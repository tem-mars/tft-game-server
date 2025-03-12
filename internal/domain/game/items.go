package game

import (
    "fmt"
    "time"
)

// เพิ่ม error constant
var (
    ErrPlayerNotFound = fmt.Errorf("player not found")
)

var DefaultItems = map[string]*Item{
    "sword": {
        ID:          "sword",
        Name:        "Sword",
        Type:        ItemWeapon,
        Attack:      5,
        Defense:     0,
        Health:      0,
        Cost:        10,
        Description: "Increases attack by 5",
    },
    "shield": {
        ID:          "shield",
        Name:        "Shield",
        Type:        ItemArmor,
        Attack:      0,
        Defense:     5,
        Health:      0,
        Cost:        10,
        Description: "Increases defense by 5",
    },
    "potion": {
        ID:          "potion",
        Name:        "Health Potion",
        Type:        ItemPotion,
        Attack:      0,
        Defense:     0,
        Health:      20,
        Cost:        5,
        Description: "Restores 20 health",
    },
}

// เพิ่มเมธอดใน GameManager
func (m *GameManager) GetAvailableItems() []*Item {
    items := make([]*Item, 0, len(DefaultItems))
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
        return fmt.Errorf("item not found")
    }

    if player.Gold < item.Cost {
        return fmt.Errorf("not enough gold")
    }

    // หักเงินและเพิ่มไอเทม
    player.Gold -= item.Cost
    player.Items = append(player.Items, item)

    // อัพเดทค่าสถานะตามไอเทม
    player.Attack += item.Attack
    player.Defense += item.Defense
    if item.Type == ItemPotion {
        player.Health = min(100, player.Health+item.Health)
    }

    game.UpdatedAt = time.Now()

    // เพิ่มประวัติการซื้อไอเทม
    game.Actions = append(game.Actions, GameAction{
        Type:      "buy_item",
        PlayerID:  playerID,
        ItemID:    itemID,
        Timestamp: time.Now(),
    })

    if m.onGameUpdate != nil {
        m.onGameUpdate(game.ID, game)
    }

    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}