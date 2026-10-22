package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestServeShutsDownWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	cfg := &config{
		addr:            "127.0.0.1:0",
		shutdownTimeout: time.Second,
	}
	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		ReadHeaderTimeout: time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := serve(ctx, server, cfg, logger); err != nil {
		t.Fatal(err)
	}
}
