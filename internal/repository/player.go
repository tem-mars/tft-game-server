package repository

import (
    "context"
    "github.com/tem-mars/tft-game-server/internal/domain/player"
)

type PlayerRepository interface {
    Create(ctx context.Context, player *player.Player) error
    GetByID(ctx context.Context, id string) (*player.Player, error)
    GetByEmail(ctx context.Context, email string) (*player.Player, error)
    Update(ctx context.Context, player *player.Player) error
    UpdateStats(ctx context.Context, stats *player.Stats) error
    GetStats(ctx context.Context, playerID string) (*player.Stats, error)
}