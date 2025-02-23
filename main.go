package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Felamande/oidcserver/config"
	"github.com/Felamande/oidcserver/exampleop"
	"github.com/Felamande/oidcserver/storage"
)

func getUserStore(cfg *config.Config) (storage.UserStore, error) {
	if cfg.UsersFile == "" {
		return storage.NewUserStore(fmt.Sprintf("http://localhost:%s/", cfg.Port)), nil
	}
	return storage.StoreFromFile(cfg.UsersFile)
}

func main() {
	cfg := config.FromEnvVars(&config.Config{Port: "9998"})
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
	)

	//which gives us the issuer: http://localhost:9998/
	// issuer := fmt.Sprintf("http://localhost:%s/", cfg.Port)
	issuer := os.Getenv("ISSUER")
	if issuer == "" {
		issuer = fmt.Sprintf("http://localhost:%s/", cfg.Port)
	}

	storage.RegisterClients(
		storage.WebClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), cfg.RedirectURI...),
	)

	// the OpenIDProvider interface needs a Storage interface handling various checks and state manipulations
	// this might be the layer for accessing your database
	// in this example it will be handled in-memory
	store, err := getUserStore(cfg)
	if err != nil {
		logger.Error("cannot create UserStore", "error", err)
		os.Exit(1)
	}
	storage := storage.NewStorage(store)
	// generate random key
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		logger.Error("cannot generate random key", "error", err)
		os.Exit(1)
	}
	keyStr := hex.EncodeToString(key)

	router := exampleop.SetupServer(issuer, keyStr, storage, logger, false)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	logger.Info("server listening, press ctrl+c to stop", "addr", issuer)
	if server.ListenAndServe() != http.ErrServerClosed {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
