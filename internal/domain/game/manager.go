package game

import (
    "context"  
    "fmt"
    "sync"
    "time"
    "github.com/tem-mars/tft-game-server/internal/repository"  // เพิ่ม import
)

type GameManager struct {
    mu         sync.RWMutex
    games      map[string]*Game
    playerRepo repository.PlayerRepository
    onGameUpdate func(*Game) 
}

func NewGameManager(playerRepo repository.PlayerRepository) *GameManager {
    return &GameManager{
        games:      make(map[string]*Game),
        playerRepo: playerRepo,
        onGameUpdate: func(*Game) {}, // default empty function
    }
}

func (m *GameManager) SetOnGameUpdate(callback func(*Game)) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.onGameUpdate = callback
}

func (m *GameManager) CreateGame(playerID string) (*Game, error) {
    // ดึงข้อมูล player จาก repository
    player, err := m.playerRepo.GetByID(context.Background(), playerID)
    if err != nil {
        return nil, err
    }

    game := &Game{
        ID:     generateGameID(),
        Status: "waiting",  // ใช้ string แทน constant
        Players: []*Player{{
            ID:       playerID,
            Username: player.Username,
            Health:   100,
            Gold:     player.Stats.Gold,
            Level:    player.Stats.Level,
            Attack:   10,
            Defense:  5,
        }},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    m.mu.Lock()
    m.games[game.ID] = game
    m.mu.Unlock()

    return game, nil
}

func (m *GameManager) JoinGame(gameID string, playerID string) error {
    player, err := m.playerRepo.GetByID(context.Background(), playerID)
    if err != nil {
        return err
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    game, exists := m.games[gameID]
    if !exists {
        return ErrGameNotFound
    }

    // เช็คว่าผู้เล่นอยู่ในเกมแล้วหรือไม่
    for _, p := range game.Players {
        if p.ID == playerID {
            return ErrPlayerAlreadyInGame
        }
    }

    // เพิ่มผู้เล่นใหม่
    game.Players = append(game.Players, &Player{
        ID:       playerID,
        Username: player.Username,
        Health:   100,
        Gold:     player.Stats.Gold,
        Level:    player.Stats.Level,
        Attack:   10,
        Defense:  5,
    })

    // อัพเดทสถานะเกม
    if len(game.Players) == 2 {
        game.Status = "playing"
    }

    game.UpdatedAt = time.Now()

    // เรียก callback เพื่ออัพเดทสถานะ
    if m.onGameUpdate != nil {
        m.onGameUpdate(game)
    }

    return nil
}


func (m *GameManager) GetGame(gameID string) (*Game, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    game, exists := m.games[gameID]
    if !exists {
        return nil, ErrGameNotFound
    }

    return game, nil
}

func generateGameID() string {
    return "game_" + time.Now().Format("20060102150405")
}

func (m *GameManager) ProcessAction(gameID string, action GameAction) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    game, exists := m.games[gameID]
    if !exists {
        return ErrGameNotFound
    }

    // ตรวจสอบว่าเกมกำลังเล่นอยู่
    if game.Status != StatusPlaying {
        return fmt.Errorf("game is not in playing state")
    }

    // หาผู้เล่นที่ทำการโจมตี
    var attacker, target *Player
    for _, p := range game.Players {
        if p.ID == action.PlayerID {
            attacker = p
        }
        if p.ID == action.TargetID {
            target = p
        }
    }

    if attacker == nil || target == nil {
        return fmt.Errorf("player not found")
    }

    switch action.Type {
    case ActionAttack:
        // คำนวณความเสียหาย
        damage := calculateDamage(attacker.Attack, target.Defense)
        target.Health -= damage

        // เช็คว่าผู้เล่นตายหรือไม่
        if target.Health <= 0 {
            target.Health = 0
            game.Status = StatusFinished
        }

        // เพิ่มประวัติการกระทำ
        game.Actions = append(game.Actions, action)
        game.UpdatedAt = time.Now()

        // ส่งอัพเดทให้ผู้เล่น
        if m.onGameUpdate != nil {
            m.onGameUpdate(game)
        }

    case ActionBuyItem:
        // TODO: Implement item purchase
        return fmt.Errorf("buy item not implemented yet")

    case ActionUseItem:
        // TODO: Implement item usage
        return fmt.Errorf("use item not implemented yet")
    }

    return nil
}

// เพิ่มฟังก์ชันคำนวณความเสียหาย
func calculateDamage(attack, defense int) int {
    damage := attack - (defense / 2)
    if damage < 1 {
        damage = 1
    }
    return damage
}

// เพิ่มเมธอดใหม่
func (m *GameManager) GetWaitingGames() []*Game {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var waitingGames []*Game
    for _, game := range m.games {
        if game.Status == "waiting" && len(game.Players) < 2 {
            waitingGames = append(waitingGames, game)
        }
    }
    return waitingGames
}

func (m *GameManager) AutoMatch(playerID string) (*Game, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // ค้นหาเกมที่รอผู้เล่น
    for _, game := range m.games {
        if game.Status == "waiting" && len(game.Players) < 2 {
            // ตรวจสอบว่าผู้เล่นไม่ได้อยู่ในเกมนี้
            playerInGame := false
            for _, p := range game.Players {
                if p.ID == playerID {
                    playerInGame = true
                    break
                }
            }

            if !playerInGame {
                // ดึงข้อมูล player
                player, err := m.playerRepo.GetByID(context.Background(), playerID)
                if err != nil {
                    return nil, err
                }

                // เพิ่มผู้เล่น
                game.Players = append(game.Players, &Player{
                    ID:       playerID,
                    Username: player.Username,
                    Health:   100,
                    Gold:     player.Stats.Gold,
                    Level:    player.Stats.Level,
                    Attack:   10,
                    Defense:  5,
                })

                game.Status = "playing"
                game.UpdatedAt = time.Now()

                // เรียก callback
                if m.onGameUpdate != nil {
                    m.onGameUpdate(game)
                }

                return game, nil
            }
        }
    }

    // สร้างเกมใหม่ถ้าไม่พบเกมที่รอ
    player, err := m.playerRepo.GetByID(context.Background(), playerID)
    if err != nil {
        return nil, err
    }

    game := &Game{
        ID:     generateGameID(),
        Status: "waiting",
        Players: []*Player{{
            ID:       playerID,
            Username: player.Username,
            Health:   100,
            Gold:     player.Stats.Gold,
            Level:    player.Stats.Level,
            Attack:   10,
            Defense:  5,
        }},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    m.games[game.ID] = game

    // เรียก callback
    if m.onGameUpdate != nil {
        m.onGameUpdate(game)
    }

    return game, nil
}

func (m *GameManager) cleanupInactiveGames() {
    now := time.Now()
    for id, game := range m.games {
        // ลบเกมที่ไม่มีผู้เล่นเกิน 5 นาที
        if len(game.Players) == 0 && now.Sub(game.UpdatedAt) > 5*time.Minute {
            delete(m.games, id)
            continue
        }
        // ลบเกมที่รอคนเล่นเกิน 10 นาที
        if game.Status == StatusWaiting && now.Sub(game.UpdatedAt) > 10*time.Minute {
            delete(m.games, id)
            continue
        }
        // ลบเกมที่จบแล้วเกิน 30 นาที
        if game.Status == StatusFinished && now.Sub(game.UpdatedAt) > 30*time.Minute {
            delete(m.games, id)
        }
    }
}
