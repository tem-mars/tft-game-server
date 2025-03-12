package game

import "time"

type ItemType string
type GameStatus string
type ActionType string

const (
    StatusWaiting  GameStatus = "waiting"
    StatusPlaying  GameStatus = "playing"
    StatusFinished GameStatus = "finished"

    ActionAttack ActionType = "attack"
    ActionBuyItem ActionType = "buy_item"
    ActionUseItem ActionType = "use_item"
)

const (
    ItemWeapon ItemType = "weapon"
    ItemArmor  ItemType = "armor"
    ItemPotion ItemType = "potion"
)

type Player struct {
    ID        string  `json:"id"`
    Username  string  `json:"username"`
    Health    int     `json:"health"`
    Gold      int     `json:"gold"`
    Level     int     `json:"level"`
    Attack    int     `json:"attack"`
    Defense   int     `json:"defense"`
    Items     []*Item `json:"items"`
}

type GameAction struct {
    Type      ActionType `json:"type"`
    PlayerID  string    `json:"player_id"`
    TargetID  string    `json:"target_id,omitempty"`
    ItemID    string    `json:"item_id,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

type Game struct {
    ID        string     `json:"id"`
    Players   []*Player  `json:"players"`
    Status    GameStatus `json:"status"`
    Actions   []GameAction `json:"actions"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

type ItemAction struct {
    Type     ActionType `json:"type"`
    PlayerID string     `json:"player_id"`
    ItemID   string     `json:"item_id"`
}

type Item struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Type        ItemType `json:"type"`
    Attack      int      `json:"attack"`
    Defense     int      `json:"defense"`
    Health      int      `json:"health"`
    Cost        int      `json:"cost"`
    Description string   `json:"description"`
}
