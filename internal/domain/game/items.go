package game

import (
    "fmt"
    "time"
)

// เพิ่ม error constants
var (
    ErrPlayerNotFound    = fmt.Errorf("player not found")
    ErrItemNotFound      = fmt.Errorf("item not found")
    ErrInsufficientGold  = fmt.Errorf("not enough gold")
)

var DefaultItems = map[string]Item{  // เปลี่ยนจาก *Item เป็น Item
    "sword": {
        ID:          "sword",
        Name:        "Sword",
        Type:        ItemTypeWeapon,  // แก้ให้ตรงกับ constant ใน types.go
        Attack:      5,
        Defense:     0,
        Health:      0,
        Cost:        10,
        Description: "Increases attack by 5",
    },
    "shield": {
        ID:          "shield",
        Name:        "Shield",
        Type:        ItemTypeArmor,   // แก้ให้ตรงกับ constant ใน types.go
        Attack:      0,
        Defense:     5,
        Health:      0,
        Cost:        10,
        Description: "Increases defense by 5",
    },
    "potion": {
        ID:          "potion",
        Name:        "Health Potion",
        Type:        ItemTypePotion,  // แก้ให้ตรงกับ constant ใน types.go
        Attack:      0,
        Defense:     0,
        Health:      20,
        Cost:        5,
        Description: "Restores 20 health",
    },
}

// แก้ไขเมธอดให้สอดคล้องกับ types ที่เปลี่ยน
func (m *GameManager) GetAvailableItems() []Item {  // เปลี่ยนจาก []*Item เป็น []Item
    items := make([]Item, 0, len(DefaultItems))
    for _, item := range DefaultItems {
        items = append(items, item)
    }
    return items
}

func (m *GameManager) BuyItem(gameID string, playerID string, itemID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    game, exists := m.games[gameID]
    if !exists {
        return ErrGameNotFound
    }

    var player *Player
    for _, p := range game.Players {
        if p.ID == playerID {
            player = p
            break
        }
    }
    if player == nil {
        return ErrPlayerNotFound
    }

    item, exists := DefaultItems[itemID]
    if !exists {
        return ErrItemNotFound
    }

    if player.Gold < item.Cost {
        return ErrInsufficientGold
    }

    // หักเงินและเพิ่มไอเทม
    player.Gold -= item.Cost
    player.Inventory = append(player.Inventory, item)  // เปลี่ยนจาก Items เป็น Inventory

    // อัพเดทค่าสถานะตามไอเทม
    player.Attack += item.Attack
    player.Defense += item.Defense
    if item.Type == ItemTypePotion {  // แก้ให้ตรงกับ constant ใน types.go
        player.Health = min(100, player.Health+item.Health)
    }

    game.UpdatedAt = time.Now()

    // เพิ่มประวัติการซื้อไอเทม
    game.Actions = append(game.Actions, GameAction{
        Type:      ActionBuyItem,  // ใช้ constant จาก types.go
        PlayerID:  playerID,
        ItemID:    itemID,
        Timestamp: time.Now(),
    })

    if m.onGameUpdate != nil {
        m.onGameUpdate(game)
    }

    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}