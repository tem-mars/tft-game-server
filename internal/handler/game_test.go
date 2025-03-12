package handler

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/tem-mars/tft-game-server/internal/domain/game"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

func setupTestRouter() (*gin.Engine, *game.GameManager, logger.Logger) {
    gin.SetMode(gin.TestMode)
    
    router := gin.New()
    gameManager := game.NewGameManager()
    log := logger.New()

    return router, gameManager, log
}

func TestCreateGame(t *testing.T) {
    router, gameManager, log := setupTestRouter()
    handler := NewGameHandler(gameManager, log)

    router.POST("/games/:gameId", handler.CreateGame)

    t.Run("Create New Game", func(t *testing.T) {
        w := httptest.NewRecorder()
        req, _ := http.NewRequest("POST", "/games/test-game-1", nil)
        router.ServeHTTP(w, req)

        if w.Code != http.StatusCreated {
            t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
        }

        var response game.Game
        err := json.Unmarshal(w.Body.Bytes(), &response)
        if err != nil {
            t.Fatalf("failed to unmarshal response: %v", err)
        }

        if response.ID != "test-game-1" {
            t.Errorf("expected game ID test-game-1, got %s", response.ID)
        }
    })

    t.Run("Create Duplicate Game", func(t *testing.T) {
        // First request
        w1 := httptest.NewRecorder()
        req1, _ := http.NewRequest("POST", "/games/test-game-2", nil)
        router.ServeHTTP(w1, req1)

        // Second request with same ID
        w2 := httptest.NewRecorder()
        req2, _ := http.NewRequest("POST", "/games/test-game-2", nil)
        router.ServeHTTP(w2, req2)

        if w2.Code != http.StatusBadRequest {
            t.Errorf("expected status %d, got %d", http.StatusBadRequest, w2.Code)
        }
    })
}