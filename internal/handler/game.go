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
    "github.com/tem-mars/tft-game-server/internal/middleware"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}



type GameHandler struct {
    gameManager *game.GameManager
    log        logger.Logger
    secret     string
    connections map[string]*websocket.Conn
    mu         sync.RWMutex  // เปลี่ยนจาก sync.Mutex เป็น sync.RWMutex
}


func NewGameHandler(gameManager *game.GameManager, log logger.Logger, secret string) *GameHandler {
    handler := &GameHandler{
        gameManager: gameManager,
        log:        log,
        secret:     secret,
        connections: make(map[string]*websocket.Conn),
    }

    // เปลี่ยนจาก SetUpdateCallback เป็น SetOnGameUpdate
    gameManager.SetOnGameUpdate(handler.broadcastGameState)

    return handler
}

func (h *GameHandler) CreateGame(c *gin.Context) {
    h.log.Info("Creating game...") // เพิ่ม logging

    claims, exists := c.Get("claims")
    if !exists {
        h.log.Error("No claims found")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
        return
    }

    userClaims, ok := claims.(*middleware.Claims)
    if !ok {
        h.log.Error("Invalid claims type")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims type"})
        return
    }

    h.log.Info("Creating game for player", logger.String("playerID", userClaims.PlayerID))

    game, err := h.gameManager.CreateGame(userClaims.PlayerID)
    if err != nil {
        h.log.Error("Failed to create game", logger.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.log.Info("Game created successfully", 
        logger.String("gameID", game.ID),
        logger.String("playerID", userClaims.PlayerID))

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "Game created successfully",
        "game": game,
    })
}

func (h *GameHandler) JoinGame(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
        return
    }

    userClaims, ok := claims.(*middleware.Claims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims type"})
        return
    }

    gameID := c.Param("gameId")
    err := h.gameManager.JoinGame(gameID, userClaims.PlayerID)
    if err != nil {
        h.log.Error("Failed to join game", 
            logger.String("gameID", gameID),
            logger.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Successfully joined game"})
}

func (h *GameHandler) broadcastGameState(game *game.Game) {
    h.log.Info("Broadcasting game state", 
        logger.String("gameID", game.ID),
        logger.Int("playerCount", len(game.Players)))

    message := map[string]interface{}{
        "type": "game_state",
        "game": game,
    }

    // ส่งข้อมูลให้ทุกคนที่เกี่ยวข้องกับเกม
    for _, player := range game.Players {
        h.mu.RLock()
        conn, exists := h.connections[player.ID]
        h.mu.RUnlock()

        if exists {
            if err := conn.WriteJSON(message); err != nil {
                h.log.Error("Failed to send game state",
                    logger.String("playerID", player.ID),
                    logger.Error(err))
            } else {
                h.log.Info("Sent game state to player",
                    logger.String("playerID", player.ID))
            }
        }
    }
}




func (h *GameHandler) GetWaitingGames(c *gin.Context) {
    h.log.Info("Getting waiting games...")

    games := h.gameManager.GetWaitingGames()
    
    h.log.Info("Found waiting games", logger.Int("count", len(games)))

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "games": games,
    })
}

func (h *GameHandler) AutoMatch(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No claims found"})
        return
    }

    userClaims, ok := claims.(*middleware.Claims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims type"})
        return
    }

    game, err := h.gameManager.AutoMatch(userClaims.PlayerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ส่ง game state ให้ทุกคนในเกม
    h.broadcastGameState(game)

    c.JSON(http.StatusOK, gin.H{
        "message": "Match found",
        "game": game,
    })
}

func (h *GameHandler) HandleWebSocket(c *gin.Context) {
    h.log.Info("New WebSocket connection attempt")

    // รับ token จาก query parameter
    token := c.Query("token")
    if token == "" {
        h.log.Error("No token provided")
        c.String(http.StatusUnauthorized, "No token provided")
        return
    }

    // ลบ prefix "Bearer " ถ้ามี
    token = strings.TrimPrefix(token, "Bearer ")
    token = strings.TrimSpace(token)

    h.log.Info("Parsing token", logger.String("token", token))

    // สร้าง claims struct
    claims := &middleware.Claims{}
    
    // parse token ด้วย Claims struct ที่สร้างไว้
    parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
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

    // ใช้ PlayerID จาก claims โดยตรง
    playerID := claims.PlayerID
    if playerID == "" {
        h.log.Error("No player ID in claims")
        c.String(http.StatusUnauthorized, "Invalid player ID")
        return
    }

    h.log.Info("Token validated successfully", 
        logger.String("playerID", playerID))

    // Upgrade connection to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        h.log.Error("Failed to upgrade connection", logger.Error(err))
        return
    }
    defer conn.Close()

    h.log.Info("WebSocket connection established", 
        logger.String("playerID", playerID))

    // เก็บ connection ในแมพ
    h.mu.Lock()
    h.connections[playerID] = conn
    h.mu.Unlock()

    // ส่งข้อความต้อนรับ
    welcome := map[string]interface{}{
        "type": "welcome",
        "message": "Connected to game server",
        "playerID": playerID,
    }
    
    if err := conn.WriteJSON(welcome); err != nil {
        h.log.Error("Failed to send welcome message", logger.Error(err))
        return
    }

    // Cleanup เมื่อจบการเชื่อมต่อ
    defer func() {
        h.mu.Lock()
        delete(h.connections, playerID)
        h.mu.Unlock()
        h.log.Info("Player disconnected", logger.String("playerID", playerID))
    }()

    // รับข้อความจาก WebSocket
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
    
        // จัดการข้อความตาม type
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
    c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *GameHandler) getPlayerClaims(c *gin.Context) (*middleware.Claims, error) {
    claims, exists := c.Get("claims")
    if !exists {
        return nil, fmt.Errorf("no claims found")
    }

    playerClaims, ok := claims.(*middleware.Claims)
    if !ok {
        return nil, fmt.Errorf("invalid claims type")
    }

    return playerClaims, nil
}


func (h *GameHandler) BuyItem(c *gin.Context) {
    claims, err := h.getPlayerClaims(c)
    if err != nil {
        h.log.Error("Failed to get player claims", logger.Error(err))
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    var itemAction game.ItemAction
    if err := c.ShouldBindJSON(&itemAction); err != nil {
        h.log.Error("Failed to bind item action", logger.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // เพิ่ม logging
    h.log.Info("Attempting to buy item",
        logger.String("playerID", claims.PlayerID),
        logger.String("gameID", itemAction.GameID),
        logger.String("itemID", itemAction.ItemID))

    err = h.gameManager.BuyItem(itemAction.GameID, claims.PlayerID, itemAction.ItemID)
    if err != nil {
        h.log.Error("Failed to buy item", logger.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.log.Info("Item purchased successfully",
        logger.String("playerID", claims.PlayerID),
        logger.String("itemID", itemAction.ItemID))

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "Item purchased successfully",  // เพิ่มเครื่องหมาย comma ตรงนี้
    })
}