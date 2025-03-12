package app

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/tem-mars/tft-game-server/internal/domain/game"
    "github.com/tem-mars/tft-game-server/internal/handler"
    "github.com/tem-mars/tft-game-server/internal/middleware"   
    "github.com/tem-mars/tft-game-server/pkg/logger"
    "github.com/tem-mars/tft-game-server/internal/service"
    "github.com/tem-mars/tft-game-server/internal/repository"
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

    // CORS middleware
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))


    playerRepo := repository.NewMemoryPlayerRepository()
    authService := service.NewAuthService(playerRepo, cfg.JWT.Secret)
    gameManager := game.NewGameManager()

    // Initialize handlers
    authHandler := handler.NewAuthHandler(authService, log)
    gameHandler := handler.NewGameHandler(gameManager, log, cfg.JWT.Secret) 

    // Public routes
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Auth routes
    router.POST("/auth/register", authHandler.Register)
    router.POST("/auth/login", authHandler.Login)


    router.GET("/games/ws", gameHandler.HandleWebSocket)

    // Protected routes
    protected := router.Group("/")
    protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    {
        protected.GET("/items", gameHandler.GetAvailableItems)
        protected.POST("/games/:gameId/buy/:itemId", gameHandler.BuyItem)
        protected.GET("/games/waiting", gameHandler.GetWaitingGames)
        protected.POST("/games/match", gameHandler.AutoMatch)
        protected.POST("/games", gameHandler.CreateGame)
        protected.POST("/games/:gameId/join", gameHandler.JoinGame)
    }

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