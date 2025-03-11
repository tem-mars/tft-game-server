package app

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

type App struct {
    cfg    *Config
    log    logger.Logger
    server *http.Server
}

func New(cfg *Config, log logger.Logger) (*App, error) {
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(LoggerMiddleware(log))

    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    server := &http.Server{
        Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
        Handler: router,
    }

    return &App{
        cfg:    cfg,
        log:    log,
        server: server,
    }, nil
}

func (a *App) Start() error {
    a.log.Info("starting server",
        logger.String("address", a.server.Addr),
    )
    return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
    return a.server.Shutdown(ctx)
}

func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        log.Info("request processed",
            logger.String("path", path),
            logger.Int("status", status),
            logger.Duration("latency", latency),
        )
    }
}