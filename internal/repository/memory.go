package repository

import (
    "fmt"
    "sync"
)

type MemoryPlayerRepository struct {
    mu      sync.RWMutex
    players map[string]*Player
}

type Player struct {
    ID       string
    Username string
    Password string
    Email    string
}

func NewMemoryPlayerRepository() *MemoryPlayerRepository {
    return &MemoryPlayerRepository{
        players: make(map[string]*Player),
    }
}

func (r *MemoryPlayerRepository) FindByUsername(username string) (*Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, player := range r.players {
        if player.Username == username {
            return player, nil
        }
    }

    return nil, fmt.Errorf("player not found")
}

func (r *MemoryPlayerRepository) Save(player *Player) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, existingPlayer := range r.players {
        if existingPlayer.Username == player.Username {
            return fmt.Errorf("username already exists")
        }
    }


    if player.ID == "" {
        player.ID = fmt.Sprintf("player_%d", len(r.players)+1)
    }

    r.players[player.ID] = player
    return nil
}