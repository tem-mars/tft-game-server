package service

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/tem-mars/tft-game-server/internal/repository"  
)

type PlayerRepository interface {
    FindByUsername(username string) (*repository.Player, error)  
    Save(player *repository.Player) error  
}

type AuthService struct {
    playerRepo PlayerRepository
    jwtSecret  string
}

func NewAuthService(playerRepo PlayerRepository, jwtSecret string) *AuthService {
    return &AuthService{
        playerRepo: playerRepo,
        jwtSecret: jwtSecret,
    }
}

func (s *AuthService) Register(username, password, email string) error {
    player := &repository.Player{  
        Username: username,
        Password: password,
        Email:    email,
    }
    return s.playerRepo.Save(player)
}

func (s *AuthService) Login(username, password string) (string, error) {
    player, err := s.playerRepo.FindByUsername(username)
    if err != nil {
        return "", err
    }

    if player.Password != password {
        return "", fmt.Errorf("invalid credentials")
    }

    
    claims := jwt.MapClaims{
        "sub": player.ID,
        "username": player.Username,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", fmt.Errorf("failed to create token: %v", err)
    }

    // Debug log
    fmt.Printf("Generated token for user %s: %s\n", username, tokenString)

    return tokenString, nil
}