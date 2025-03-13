package repository

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type Player struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"` // ไม่แสดงใน JSON
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Stats     *Stats    `json:"stats"`
}

type Stats struct {
    PlayerID  string    `json:"player_id"`
    Wins      int       `json:"wins"`
    Losses    int       `json:"losses"`
    Gold      int       `json:"gold"`
    Level     int       `json:"level"`
    UpdatedAt time.Time `json:"updated_at"`
}

type MemoryPlayerRepository struct {
    mu      sync.RWMutex
    players map[string]*Player // key: username
    stats   map[string]*Stats // key: playerID
}

func NewMemoryPlayerRepository() *MemoryPlayerRepository {
    return &MemoryPlayerRepository{
        players: make(map[string]*Player),
        stats:   make(map[string]*Stats),
    }
}

// Interface methods
type PlayerRepository interface {
    Create(ctx context.Context, username, email, password string) (*Player, error)
    GetByUsername(username string) (*Player, error)
    GetByID(ctx context.Context, id string) (*Player, error)
    GetByEmail(ctx context.Context, email string) (*Player, error)
    Update(ctx context.Context, player *Player) error
    UpdateStats(ctx context.Context, stats *Stats) error
    GetStats(ctx context.Context, playerID string) (*Stats, error)
}

// Implementation
func (r *MemoryPlayerRepository) Create(ctx context.Context, username, email, password string) (*Player, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.players[username]; exists {
        return nil, fmt.Errorf("username already exists")
    }

    now := time.Now()
    player := &Player{
        ID:        fmt.Sprintf("player_%d", len(r.players)+1),
        Username:  username,
        Email:     email,
        Password:  password,
        CreatedAt: now,
        UpdatedAt: now,
        Stats: &Stats{
            Gold:      100,
            Level:     1,
            UpdatedAt: now,
        },
    }

    r.players[username] = player
    player.Stats.PlayerID = player.ID
    r.stats[player.ID] = player.Stats

    return player, nil
}

func (r *MemoryPlayerRepository) GetByUsername(username string) (*Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    player, exists := r.players[username]
    if !exists {
        return nil, fmt.Errorf("player not found")
    }

    return player, nil
}

func (r *MemoryPlayerRepository) GetByID(ctx context.Context, id string) (*Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, player := range r.players {
        if player.ID == id {
            return player, nil
        }
    }

    return nil, fmt.Errorf("player not found")
}

func (r *MemoryPlayerRepository) GetByEmail(ctx context.Context, email string) (*Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, player := range r.players {
        if player.Email == email {
            return player, nil
        }
    }

    return nil, fmt.Errorf("player not found")
}

func (r *MemoryPlayerRepository) Update(ctx context.Context, player *Player) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // ตรวจสอบว่ามี player อยู่หรือไม่
    if _, exists := r.players[player.Username]; !exists {
        return fmt.Errorf("player not found")
    }

    player.UpdatedAt = time.Now()
    r.players[player.Username] = player

    if player.Stats != nil {
        player.Stats.UpdatedAt = time.Now()
        r.stats[player.ID] = player.Stats
    }

    return nil
}

func (r *MemoryPlayerRepository) UpdateStats(ctx context.Context, stats *Stats) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    stats.UpdatedAt = time.Now()
    r.stats[stats.PlayerID] = stats
    return nil
}

func (r *MemoryPlayerRepository) GetStats(ctx context.Context, playerID string) (*Stats, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    stats, exists := r.stats[playerID]
    if !exists {
        return nil, fmt.Errorf("stats not found")
    }

    return stats, nil
}