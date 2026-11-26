package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Hub keeps the live WebSocket connections of each room. The server is a single
// process, so this lives in memory; the database stays the source of truth and every
// message is a full snapshot, so a phone that reconnects never needs a replay.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[*client]struct{}
}

type client struct {
	userID string
	send   chan []byte
	kicked chan struct{}
	once   sync.Once
	code   websocket.StatusCode // why it was kicked; read only after kicked is closed
}

// Close codes the web app understands: after closeLeft it stops reconnecting.
const (
	closeLeft websocket.StatusCode = 4000 // the user left the room
	closeSlow websocket.StatusCode = 4001 // too far behind; reconnect for a fresh snapshot
)

func (c *client) kick(code websocket.StatusCode) {
	c.once.Do(func() {
		c.code = code
		close(c.kicked)
	})
}

func NewHub() *Hub { return &Hub{rooms: map[string]map[*client]struct{}{}} }

func (h *Hub) add(roomID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = map[*client]struct{}{}
	}
	h.rooms[roomID][c] = struct{}{}
}

func (h *Hub) remove(roomID string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[roomID], c)
	if len(h.rooms[roomID]) == 0 {
		delete(h.rooms, roomID)
	}
}

// Online returns who has at least one open connection to the room.
func (h *Hub) Online(roomID string) map[string]bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := map[string]bool{}
	for c := range h.rooms[roomID] {
		out[c.userID] = true
	}
	return out
}

// Send queues a message for everyone in the room. A client too slow to keep up
// is dropped; it reconnects and gets a fresh snapshot.
func (h *Hub) Send(roomID string, msg any) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.rooms[roomID] {
		select {
		case c.send <- b:
		default:
			c.kick(closeSlow)
		}
	}
}

// Kick closes a user's connections to a room (after they leave it).
func (h *Hub) Kick(roomID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.rooms[roomID] {
		if c.userID == userID {
			c.kick(closeLeft)
		}
	}
}

const (
	pingEvery    = 25 * time.Second
	writeTimeout = 10 * time.Second
)

func (a *App) handleRoomSocket(w http.ResponseWriter, r *http.Request) {
	rr, ok := a.memberRoom(w, r)
	if !ok {
		return
	}
	a.serveSocket(w, r, rr.ID, func() { a.broadcastRoom(rr.ID) })
}

// userChannel is the hub key for one person's own socket (friend requests, invites).
func userChannel(userID string) string { return "user:" + userID }

// handleUserSocket tells a signed-in person when their inbox changed; the site then refetches it.
func (a *App) handleUserSocket(w http.ResponseWriter, r *http.Request) {
	uid := currentUser(r.Context())
	a.serveSocket(w, r, userChannel(uid), func() { a.notifyUser(uid) })
}

// notifyUser pokes every open tab of a person to refresh their inbox.
func (a *App) notifyUser(userID string) {
	a.hub.Send(userChannel(userID), map[string]string{"type": "inbox"})
}

// serveSocket upgrades the request and keeps a client in the hub under key until it
// disconnects. changed runs after the client joins and after it leaves, so the others
// (and the client itself, on joining) get a fresh snapshot.
func (a *App) serveSocket(w http.ResponseWriter, r *http.Request, key string, changed func()) {
	uid := currentUser(r.Context())
	opts := &websocket.AcceptOptions{}
	if u, err := url.Parse(a.cfg.PublicURL); err == nil && u.Host != "" {
		opts.OriginPatterns = []string{u.Host}
	}
	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		return // Accept already wrote the HTTP error
	}
	defer conn.CloseNow()

	c := &client{userID: uid, send: make(chan []byte, 16), kicked: make(chan struct{})}
	a.hub.add(key, c)
	defer func() {
		a.hub.remove(key, c)
		go changed()
	}()
	changed()

	ctx := conn.CloseRead(context.Background()) // clients only listen; this notices them closing
	ping := time.NewTicker(pingEvery)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.kicked:
			conn.Close(c.code, "")
			return
		case msg := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
