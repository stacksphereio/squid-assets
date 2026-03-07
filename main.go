package main

import (
	"context"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"squid-assets/internal/db"
	"squid-assets/internal/featureflags"
	mw "squid-assets/internal/http/middleware"
	"squid-assets/internal/logger"
	"squid-assets/internal/storage"
)

func main() {
	// 1) DB init
	sqlDB, err := db.Init()
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer sqlDB.Close()

	// 2) Feature flags init (non-fatal)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := featureflags.Init(ctx, ""); err != nil {
		log.Printf("feature flags init warning: %v", err)
	} else {
		log.Printf("feature flags ready: offline=%v, logLevel=%s",
			featureflags.Values().Offline.IsEnabled(nil),
			featureflags.Values().LogLevel.GetValue(nil))
	}
	defer featureflags.Shutdown()

	// 2a) Initialize levelled logger from flag & watch for flips
	logger.Init(featureflags.Values().LogLevel.GetValue(nil))
	logger.Infof("log level set to %s", logger.GetLevel())

	go func() {
		prev := featureflags.Values().LogLevel.GetValue(nil)
		for {
			time.Sleep(5 * time.Second)
			cur := featureflags.Values().LogLevel.GetValue(nil)
			if cur != prev {
				logger.SetLevel(cur)
				logger.Infof("log level changed to %s", logger.GetLevel())
				prev = cur
			}
		}
	}()

	// 4) Router
	r := mux.NewRouter()

	// 4a) Offline kill-switch middleware (placed immediately after router creation)
	offlineGate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// always allow health checks
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			// block all other requests when Offline flag is ON
			if featureflags.Values().Offline.IsEnabled(nil) {
				http.Error(w, "service temporarily offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	r.Use(offlineGate)

	// 4b) Request logger (skip noisy health endpoints)
	r.Use(mw.LogRequests(mw.WithSkips("/health", "/ready")))

	// 5) Health endpoints
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if err := sqlDB.Ping(); err != nil {
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	}).Methods(http.MethodGet)

	// 6) Inspect current flag values
	r.HandleFunc("/_flags", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"offline":              featureflags.Values().Offline.IsEnabled(nil),
			"logLevel":             featureflags.Values().LogLevel.GetValue(nil),
			"imageStorageLocation": featureflags.Values().ImageStorageLocation.GetValue(nil),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodGet)

	// 7) Asset serving endpoints
	assetsBasePath := os.Getenv("ASSETS_BASE_PATH")
	if assetsBasePath == "" {
		assetsBasePath = "./assets" // Default to bundled assets
	}

	r.PathPrefix("/assets/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract path after /assets/
		assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")

		// Determine storage backend based on feature flag
		storageLocation := featureflags.Values().ImageStorageLocation.GetValue(nil)
		var store storage.Storage

		switch storageLocation {
		case "bucket":
			// TODO: Implement bucket storage when ready
			logger.Warnf("bucket storage not yet implemented, falling back to local")
			store = storage.NewLocalStorage(assetsBasePath)
		case "local":
			fallthrough
		default:
			store = storage.NewLocalStorage(assetsBasePath)
		}

		data, err := store.Get(assetPath)
		if err != nil {
			logger.Debugf("asset not found: %s (%v)", assetPath, err)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Detect content type - check if it's SVG regardless of extension
		contentType := mime.TypeByExtension(filepath.Ext(assetPath))

		// Detect SVG files by content (even if named .jpg)
		if len(data) > 4 && (string(data[:4]) == "<svg" || string(data[:5]) == "<?xml") {
			contentType = "image/svg+xml"
		} else if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache
		w.Write(data)
	}).Methods(http.MethodGet)

	s := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Infof("squid-assets listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
