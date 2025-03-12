package handler

import (
    "fmt"  
    "net/http"
    "sync"
    "time" 
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/golang-jwt/jwt/v5"
    "github.com/tem-mars/tft-game-server/internal/domain/game"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Claims struct {
    PlayerID string `json:"sub"`  
    Username string `json:"username"`
    jwt.RegisteredClaims
}

type GameHandler struct {
    gameManager *game.GameManager
    log        logger.Logger
    connections map[string]*websocket.Conn
    mu         sync.Mutex
    secret     string
}


func NewGameHandler(gameManager *game.GameManager, log logger.Logger, secret string) *GameHandler {
    handler := &GameHandler{
        gameManager: gameManager,
        log:        log,
        connections: make(map[string]*websocket.Conn),
        secret:     secret,
    }

    
    gameManager.SetUpdateCallback(handler.broadcastGameState)

    return handler
}

func (h *GameHandler) CreateGame(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    
    if tokenClaims, ok := claims.(jwt.MapClaims); ok {
        playerID, ok := tokenClaims["sub"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid player id"})
            return
        }

        
        h.log.Info("Creating game", 
            logger.String("playerID", playerID))

        game, err := h.gameManager.CreateGame(playerID)
        if err != nil {
            h.log.Error("Failed to create game", logger.Error(err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        h.log.Info("Game created successfully", 
            logger.String("gameID", game.ID),
            logger.String("playerID", playerID))

        c.JSON(http.StatusCreated, game)
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
    }
}
func (h *GameHandler) JoinGame(c *gin.Context) {
    gameID := c.Param("gameId")
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if tokenClaims, ok := claims.(jwt.MapClaims); ok {
        playerID, ok := tokenClaims["sub"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid player id"})
            return
        }

        err := h.gameManager.JoinGame(gameID, playerID)
        if err != nil {
            h.log.Error("Failed to join game", logger.Error(err))
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "joined game successfully"})
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
    }
}

func (h *GameHandler) broadcastGameState(gameID string, game *game.Game) {
    h.mu.Lock()
    defer h.mu.Unlock()

    message := map[string]interface{}{
        "type": "game_state",
        "game": game,
    }

    for _, player := range game.Players {
        if conn, ok := h.connections[player.ID]; ok {
            if err := conn.WriteJSON(message); err != nil {
                h.log.Error("Failed to send game state",
                    logger.String("playerID", player.ID),
                    logger.Error(err),
                )
            }
        }
    }
}


func (h *GameHandler) GetWaitingGames(c *gin.Context) {
    games := h.gameManager.GetWaitingGames()
    c.JSON(http.StatusOK, gin.H{
        "games": games,
    })
}

func (h *GameHandler) AutoMatch(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    jwtClaims, ok := claims.(*jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
        return
    }

    playerID, ok := (*jwtClaims)["player_id"].(string)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid player ID"})
        return
    }

    game, err := h.gameManager.AutoMatch(playerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Match found",
        "game": game,
    })
}

func (h *GameHandler) HandleWebSocket(c *gin.Context) {
    h.log.Info("New WebSocket connection attempt")

    
    token := c.Query("token")
    if token == "" {
        h.log.Error("No token provided")
        c.String(http.StatusUnauthorized, "No token provided")
        return
    }

    
    token = strings.TrimPrefix(token, "Bearer ")
    token = strings.TrimSpace(token)

    h.log.Info("Parsing token", logger.String("token", token))

    
    parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(h.secret), nil
    })

    if err != nil {
        h.log.Error("Failed to parse token", logger.Error(err))
        c.String(http.StatusUnauthorized, "Invalid token")
        return
    }

    if !parsedToken.Valid {
        h.log.Error("Token is invalid")
        c.String(http.StatusUnauthorized, "Invalid token")
        return
    }

    
    claims, ok := parsedToken.Claims.(jwt.MapClaims)
    if !ok {
        h.log.Error("Failed to get claims from token")
        c.String(http.StatusUnauthorized, "Invalid claims")
        return
    }

    
    playerID, ok := claims["sub"].(string)
    if !ok {
        h.log.Error("No player ID in claims")
        c.String(http.StatusUnauthorized, "Invalid player ID")
        return
    }

    h.log.Info("Token validated successfully", 
        logger.String("playerID", playerID))

    // Upgrade connection
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        h.log.Error("Failed to upgrade connection", logger.Error(err))
        return
    }
    defer conn.Close()

    h.log.Info("WebSocket connection established", 
        logger.String("playerID", playerID))

    
    h.mu.Lock()
    h.connections[playerID] = conn
    h.mu.Unlock()

    welcome := map[string]interface{}{
        "type": "welcome",
        "message": "Connected to game server",
        "playerID": playerID,
    }
    
    if err := conn.WriteJSON(welcome); err != nil {
        h.log.Error("Failed to send welcome message", logger.Error(err))
        return
    }

    // Cleanup 
    defer func() {
        h.mu.Lock()
        delete(h.connections, playerID)
        h.mu.Unlock()
        h.log.Info("Player disconnected", logger.String("playerID", playerID))
    }()

    for {
        var message map[string]interface{}
        err := conn.ReadJSON(&message)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                h.log.Error("WebSocket read error", logger.Error(err))
            }
            return
        }
    
        h.log.Info("Received message from client", 
            logger.String("playerID", playerID),
            logger.String("messageType", fmt.Sprintf("%v", message["type"])))
    
        switch message["type"] {
        case "get_game_state":
            if gameID, ok := message["game_id"].(string); ok {
                if game, err := h.gameManager.GetGame(gameID); err == nil {
                    response := map[string]interface{}{
                        "type": "game_state",
                        "data": game,
                    }
                    if err := conn.WriteJSON(response); err != nil {
                        h.log.Error("Failed to send game state", logger.Error(err))
                    }
                }
            }
        case "attack":
            if gameID, ok := message["game_id"].(string); ok {
                if targetID, ok := message["target_id"].(string); ok {
                    action := game.GameAction{
                        Type:      game.ActionAttack,
                        PlayerID:  playerID,
                        TargetID:  targetID,
                        Timestamp: time.Now(),
                    }
                    
                    if err := h.gameManager.ProcessAction(gameID, action); err != nil {
                        h.log.Error("Failed to process attack", 
                            logger.String("gameID", gameID),
                            logger.String("playerID", playerID),
                            logger.Error(err))
                        
                        errorResponse := map[string]interface{}{
                            "type": "error",
                            "message": err.Error(),
                        }
                        conn.WriteJSON(errorResponse)
                    }
                }
            }
        }
    }

    
    
}

func (h *GameHandler) GetAvailableItems(c *gin.Context) {
    items := h.gameManager.GetAvailableItems()
    c.JSON(http.StatusOK, gin.H{
        "items": items,
    })
}

func (h *GameHandler) BuyItem(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    if tokenClaims, ok := claims.(jwt.MapClaims); ok {
        playerID, ok := tokenClaims["sub"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid player id"})
            return
        }

        gameID := c.Param("gameId")
        itemID := c.Param("itemId")

        err := h.gameManager.BuyItem(gameID, playerID, itemID)
        if err != nil {
            h.log.Error("Failed to buy item", logger.Error(err))
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "item purchased successfully"})
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
    }
}