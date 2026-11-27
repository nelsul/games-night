// Command api is the Games Night HTTP API. It also serves the built web app.
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsul/games-night/api/games"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Config struct {
	Addr           string
	DatabaseURL    string
	PublicURL      string
	StaticDir      string
	CookieSecure   bool
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	MailFrom       string
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3Bucket       string
	S3UseSSL       bool
	S3Region       string
	S3CreateBucket bool
	Repo           string // owner/name on GitHub, for the contributors list
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() (Config, error) {
	c := Config{
		Addr:           env("ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		PublicURL:      strings.TrimRight(env("PUBLIC_URL", "http://localhost:5173"), "/"),
		StaticDir:      os.Getenv("STATIC_DIR"),
		CookieSecure:   env("COOKIE_SECURE", "false") == "true",
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       env("SMTP_PORT", "587"),
		SMTPUser:       os.Getenv("SMTP_USER"),
		SMTPPass:       os.Getenv("SMTP_PASS"),
		MailFrom:       os.Getenv("MAIL_FROM"),
		S3Endpoint:     os.Getenv("S3_ENDPOINT"),
		S3AccessKey:    os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:    os.Getenv("S3_SECRET_KEY"),
		S3Bucket:       env("S3_BUCKET", "games-night"),
		S3UseSSL:       env("S3_USE_SSL", "true") == "true",
		S3Region:       env("S3_REGION", "auto"),
		S3CreateBucket: env("S3_CREATE_BUCKET", "false") == "true",
		Repo:           env("GITHUB_REPO", "nelsul/games-night"),
	}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL is required")
	}
	return c, nil
}

type App struct {
	cfg          Config
	db           *pgxpool.Pool
	store        *Storage // nil when S3 is not configured: profile photos are then disabled
	mail         *Mailer
	hub          *Hub
	bots         *botRunner
	contributors *contributorsCache
	log          *slog.Logger
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := connectDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	hubMigrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	if err := migrate(ctx, db, log, hubMigrations, "public", ""); err != nil {
		return fmt.Errorf("migrate hub: %w", err)
	}
	// Each game keeps its tables in its own schema, migrated as its own role.
	for _, e := range games.All() {
		if err := setupGameDB(ctx, db, e); err != nil {
			return fmt.Errorf("set up %s: %w", e.Info().ID, err)
		}
		if err := migrate(ctx, db, log, e.Migrations(), e.Info().ID, games.Role(e.Info().ID)); err != nil {
			return fmt.Errorf("migrate %s: %w", e.Info().ID, err)
		}
	}

	var store *Storage
	if cfg.S3Endpoint != "" && cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		if store, err = NewStorage(ctx, cfg); err != nil {
			return fmt.Errorf("storage: %w", err)
		}
	} else {
		log.Warn("S3 is not configured: profile photos are disabled")
	}

	app := &App{cfg: cfg, db: db, store: store, mail: NewMailer(cfg, log), hub: NewHub(), bots: newBotRunner(), contributors: &contributorsCache{}, log: log}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	app.settleAllRooms(ctx)
	app.wakeAllBots(ctx)
	log.Info("listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func connectDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < 30; i++ {
		db, err := pgxpool.New(ctx, url)
		if err == nil {
			if err = db.Ping(ctx); err == nil {
				return db, nil
			}
			db.Close()
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("connect to database: %w", lastErr)
}

// migrate applies every *.sql file in fsys not yet recorded in <schema>.schema_migrations,
// in name order, each in its own transaction. With a role (games), everything runs as it.
func migrate(ctx context.Context, db *pgxpool.Pool, log *slog.Logger, fsys fs.FS, schema, role string) error {
	asRole := func(tx pgx.Tx) error {
		if role == "" {
			return nil
		}
		_, err := tx.Exec(ctx, `SET LOCAL ROLE `+pgx.Identifier{role}.Sanitize())
		return err
	}
	table := pgx.Identifier{schema, "schema_migrations"}.Sanitize()
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if err := asRole(tx); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS `+table+` (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`)
		return err
	})
	if err != nil {
		return err
	}
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE name = $1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sql, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		err = pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			if err := asRole(tx); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			_, err := tx.Exec(ctx, `INSERT INTO `+table+` (name) VALUES ($1)`, name)
			return err
		})
		if err != nil {
			return err
		}
		log.Info("migration applied", "schema", schema, "name", name)
	}
	return nil
}

func (a *App) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"ok": "true"}) })
		r.Get("/games", a.handleListGames)
		r.Get("/contributors", a.handleContributors)
		r.Post("/auth/request-code", a.handleRequestCode)
		r.Post("/auth/verify", a.handleVerifyCode)
		r.Post("/auth/logout", a.handleLogout)

		r.Group(func(r chi.Router) {
			r.Use(a.requireUser)
			r.Get("/me", a.handleMe)
			r.Patch("/me", a.handleUpdateMe)
			r.Post("/me/avatar", a.handleUploadAvatar)
			r.Delete("/me/avatar", a.handleDeleteAvatar)
			r.Get("/me/ws", a.handleUserSocket)
			r.Get("/inbox", a.handleInbox)
			r.Get("/handles/available", a.handleHandleAvailable)
			r.Get("/users/search", a.handleSearchUsers)
			r.Get("/users/{id}", a.handleGetUser)

			r.Get("/friends", a.handleListFriends)
			r.Delete("/friends/{id}", a.handleRemoveFriend)
			r.Get("/friend-requests", a.handleListRequests)
			r.Post("/friend-requests", a.handleSendRequest)
			r.Post("/friend-requests/{id}/{action}", a.handleRequestAction)
			r.Post("/invites/{id}/decline", a.handleDeclineInvite)

			r.Get("/rooms", a.handleListRooms)
			r.Post("/rooms", a.handleCreateRoom)
			r.Get("/rooms/{code}", a.handleGetRoom)
			r.Post("/rooms/{code}/join", a.handleJoinRoom)
			r.Post("/rooms/{code}/leave", a.handleLeaveRoom)
			r.Get("/rooms/{code}/ws", a.handleRoomSocket)
			r.Get("/rooms/{code}/friends", a.handleRoomFriends)
			r.Post("/rooms/{code}/invites", a.handleInviteToRoom)
			r.Put("/rooms/{code}/options", a.handleRoomOptions)
			r.Post("/rooms/{code}/start", a.handleStartMatch)
			r.Post("/rooms/{code}/action", a.handleMatchAction)
			r.Post("/rooms/{code}/stop", a.handleStopMatch)
			r.Post("/rooms/{code}/bots", a.handleAddBots)
			r.Delete("/rooms/{code}/bots", a.handleRemoveBots)
			r.Get("/rooms/{code}/messages", a.handleListMessages)
			r.Post("/rooms/{code}/messages", a.handleSendMessage)
		})
		r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeError(w, 404, "Não encontrado.") })
	})

	if a.cfg.StaticDir != "" {
		r.Handle("/*", spaHandler(a.cfg.StaticDir))
	}
	return r
}

// spaHandler serves the built web app and falls back to index.html for client routes.
func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Clean("/" + r.URL.Path)
		if info, err := os.Stat(filepath.Join(dir, p)); err == nil && !info.IsDir() {
			if strings.HasPrefix(p, "/_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
