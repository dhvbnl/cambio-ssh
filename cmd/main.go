package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	wish "github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/dhvbnl/cambio-ssh/cmd/cli"
)

func main() {
	host := envOrDefault("SSH_HOST", "0.0.0.0")
	port := envOrDefault("SSH_PORT", "22")
	address := host + ":" + port

	hostKeyPath := envOrDefault("SSH_HOST_KEY_PATH", filepath.Join(".ssh", "id_ed25519"))
	if err := os.MkdirAll(filepath.Dir(hostKeyPath), 0o755); err != nil {
		log.Fatalf("failed to create host key directory: %v", err)
	}

	server, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaSessionHandler),
			activeterm.Middleware(),
			logging.Middleware(),
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

func teaSessionHandler(_ ssh.Session) (tea.Model, []tea.ProgramOption) {
	return cli.NewRootModel(), []tea.ProgramOption{tea.WithAltScreen()}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
