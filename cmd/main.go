package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	wish "charm.land/wish/v2"
	bm "charm.land/wish/v2/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/dhvbnl/cambio-ssh/cmd/cli"
)

func main() {
	host := envOrDefault("SSH_HOST", "0.0.0.0")
	port := envOrDefault("SSH_PORT", "23234")
	address := host + ":" + port

	hostKeyPath := envOrDefault("SSH_HOST_KEY_PATH", filepath.Join(".ssh", "id_ed25519"))
	if err := os.MkdirAll(filepath.Dir(hostKeyPath), 0o755); err != nil {
		log.Fatalf("failed to create host key directory: %v", err)
	}

	server, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bm.Middleware(func(_ ssh.Session) (tea.Model, []tea.ProgramOption) {
				return cli.NewGameModel(), nil
			}),
		),
	)
	if err != nil {
		log.Fatalf("failed to create ssh server: %v", err)
	}

	log.Printf("blackjack ssh server listening on %s", address)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil {
		if err != ssh.ErrServerClosed {
			log.Fatalf("ssh server error: %v", err)
		}
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
