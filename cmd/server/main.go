package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/tem-mars/tft-game-server/internal/app"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

func main() {
    log := logger.New()
    defer log.Sync()

    cfg, err := app.LoadConfig()
    if err != nil {
        log.Fatal("failed to load configuration", 
            logger.Error(err),
        )
    }

    application, err := app.New(cfg, log)
    if err != nil {
        log.Fatal("failed to create application", 
            logger.Error(err),
        )
    }

    // Start server in a goroutine
    go func() {
        if err := application.Start(); err != nil {
            log.Fatal("failed to start server", 
                logger.Error(err),
            )
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := application.Shutdown(ctx); err != nil {
        log.Fatal("server forced to shutdown", 
            logger.Error(err),
        )
    }

    log.Info("server exited properly")
}