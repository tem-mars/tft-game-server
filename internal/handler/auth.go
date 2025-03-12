package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/tem-mars/tft-game-server/pkg/logger"
)

type AuthService interface {
    Register(username, password, email string) error
    Login(username, password string) (string, error)  
}

type AuthHandler struct {
    authService AuthService
    log        logger.Logger
}

func NewAuthHandler(authService AuthService, log logger.Logger) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        log:        log,
    }
}

type RegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Email    string `json:"email" binding:"required"`
}

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := h.authService.Register(req.Username, req.Password, req.Email); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "registration successful"})
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    token, err := h.authService.Login(req.Username, req.Password)
    if err != nil {
        c.JSON(401, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"token": token})
}