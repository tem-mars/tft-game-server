package service

import (
    "context"  // เพิ่ม import
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/tem-mars/tft-game-server/internal/repository"
    "github.com/tem-mars/tft-game-server/internal/middleware"
)

type PlayerRepository interface {
    Create(ctx context.Context, username, email, password string) (*repository.Player, error)
    GetByUsername(username string) (*repository.Player, error)
    GetByID(ctx context.Context, id string) (*repository.Player, error)
    GetByEmail(ctx context.Context, email string) (*repository.Player, error)
    Update(ctx context.Context, player *repository.Player) error
    UpdateStats(ctx context.Context, stats *repository.Stats) error
    GetStats(ctx context.Context, playerID string) (*repository.Stats, error)
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
    _, err := s.playerRepo.Create(context.Background(), username, email, password)
    return err
}

func (s *AuthService) Login(username, password string) (string, error) {
    player, err := s.playerRepo.GetByUsername(username)
    if err != nil {
        return "", err
    }

    if player.Password != password {
        return "", fmt.Errorf("invalid credentials")
    }

    claims := &middleware.Claims{
        PlayerID: player.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   player.ID,
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}