package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage keeps pictures in an S3-compatible bucket: Cloudflare R2 in
// production, a local S3 server in development. Objects are private; clients get short-lived
// presigned URLs only for pictures they are allowed to see.
type Storage struct {
	client *minio.Client
	bucket string

	// Presigned URLs carry a signature that changes every second, so handing out a
	// fresh one each time makes browsers download the same photo again and again.
	// Each key keeps one URL until it's close to expiring.
	mu   sync.Mutex
	urls map[string]signedURL
}

type signedURL struct {
	url     string
	expires time.Time
}

const (
	urlTTL       = 6 * time.Hour
	urlReuse     = 5 * time.Hour // hand out the same URL for this long, then sign a new one
	cacheForever = "public, max-age=31536000, immutable"
)

const maxImageBytes = 8 << 20

var imageTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/heic": "heic",
	"image/gif":  "gif",
}

func NewStorage(ctx context.Context, cfg Config) (*Storage, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, err
	}
	s := &Storage{client: client, bucket: cfg.S3Bucket}
	if cfg.S3CreateBucket {
		exists, err := client.BucketExists(ctx, cfg.S3Bucket)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{Region: cfg.S3Region}); err != nil {
				return nil, err
			}
		}
	}
	return s, nil
}

// readImage reads the "file" part of a multipart request, checks it is a picture
// and shrinks it (longest side at most maxSide) before it is stored.
func readImage(w http.ResponseWriter, r *http.Request, maxSide int) ([]byte, string, string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImageBytes+1<<20)
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "Envie uma imagem de até 8 MB.")
		return nil, "", "", false
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxImageBytes {
		writeError(w, 400, "Envie uma imagem de até 8 MB.")
		return nil, "", "", false
	}
	ctype := http.DetectContentType(data)
	ext, ok := imageTypes[ctype]
	if !ok && isHEIC(data) {
		ctype, ext, ok = "image/heic", "heic", true
	}
	if !ok {
		writeError(w, 400, "Formato não suportado. Use JPG, PNG, WEBP ou HEIC.")
		return nil, "", "", false
	}
	data, ctype, ext = optimizeImage(data, ctype, ext, maxSide)
	return data, ctype, ext, true
}

func isHEIC(b []byte) bool {
	if len(b) < 12 || string(b[4:8]) != "ftyp" {
		return false
	}
	switch string(b[8:12]) {
	case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1":
		return true
	}
	return false
}

func (s *Storage) Put(ctx context.Context, prefix, ext, ctype string, data []byte) (string, error) {
	key := fmt.Sprintf("%s/%s.%s", prefix, uuid.NewString(), ext)
	// Keys are unique per upload and never rewritten, so browsers may keep them forever.
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: ctype, CacheControl: cacheForever})
	return key, err
}

func (s *Storage) Delete(ctx context.Context, key string) {
	if s == nil || key == "" {
		return
	}
	_ = s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// URL returns a presigned GET URL, or "" for an empty key or when storage is off.
func (s *Storage) URL(ctx context.Context, key *string) string {
	if s == nil || key == nil || *key == "" {
		return ""
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.urls == nil {
		s.urls = map[string]signedURL{}
	}
	if c, ok := s.urls[*key]; ok && now.Before(c.expires.Add(urlReuse-urlTTL)) {
		return c.url
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, *key, urlTTL, url.Values{
		// Lets R2 send the long cache header even for photos uploaded before it was set.
		"response-cache-control": []string{cacheForever},
	})
	if err != nil {
		return ""
	}
	if len(s.urls) > 10_000 { // a safety valve; there are far fewer photos than this
		clear(s.urls)
	}
	s.urls[*key] = signedURL{url: u.String(), expires: now.Add(urlTTL)}
	return u.String()
}
