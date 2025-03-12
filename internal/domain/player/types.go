package player

import (
    "time"
)

type Player struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"` // ไม่แสดงใน JSON
    MMR       int       `json:"mmr"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Stats struct {
    PlayerID    string `json:"player_id"`
    GamesPlayed int    `json:"games_played"`
    Wins        int    `json:"wins"`
    Top4        int    `json:"top4"`
    AvgPlace    float64 `json:"avg_place"`
}