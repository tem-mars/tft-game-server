package game

import "time"

type ItemType string
type GameStatus string
type ActionType string

const (
    StatusWaiting  GameStatus = "waiting"
    StatusPlaying  GameStatus = "playing"
    StatusFinished GameStatus = "finished"

    ActionAttack  ActionType = "attack"
    ActionBuyItem ActionType = "buy_item"
    ActionUseItem ActionType = "use_item"
)

// ลบ constants ที่ซ้ำกันออก เหลือแค่ชุดเดียว
const (
    ItemTypeWeapon ItemType = "weapon"
    ItemTypeArmor  ItemType = "armor"
    ItemTypePotion ItemType = "potion"
)

type Player struct {
    ID        string  `json:"id"`
    Username  string  `json:"username"`
    Health    int     `json:"health"`
    Gold      int     `json:"gold"`      // เหลือ Gold แค่ตัวเดียว
    Level     int     `json:"level"`
    Attack    int     `json:"attack"`
    Defense   int     `json:"defense"`
    Inventory []Item  `json:"inventory"` // เปลี่ยนจาก Items เป็น Inventory
}

type GameAction struct {
    Type      ActionType `json:"type"`
    PlayerID  string     `json:"player_id"`
    TargetID  string     `json:"target_id,omitempty"`
    ItemID    string     `json:"item_id,omitempty"`
    Timestamp time.Time  `json:"timestamp"`
}

type Game struct {
    ID        string       `json:"id"`
    Players   []*Player    `json:"players"`
    Status    GameStatus   `json:"status"`
    Actions   []GameAction `json:"actions"`
    CreatedAt time.Time    `json:"created_at"`
    UpdatedAt time.Time    `json:"updated_at"`
}

type ItemAction struct {
    Type     ActionType `json:"type"`
    GameID   string     `json:"game_id"` 
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