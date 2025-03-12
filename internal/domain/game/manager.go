package game

import (
    "fmt" 
    "sync"
    "time"
)

type GameManager struct {
    games map[string]*Game
    mu    sync.RWMutex
    onGameUpdate func(gameID string, game *Game)
}

func NewGameManager() *GameManager {
    return &GameManager{
        games: make(map[string]*Game),
    }
}

func (m *GameManager) SetUpdateCallback(callback func(gameID string, game *Game)) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.onGameUpdate = callback
}

// แก้ไขฟังก์ชัน CreateGame
func (m *GameManager) CreateGame(playerID string) (*Game, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    game := &Game{
        ID: generateGameID(),
        Players: []*Player{
            {
                ID:      playerID,
                Health:  100,
                Gold:    0,
                Level:   1,
                Attack:  10,    // เพิ่มค่าโจมตีเริ่มต้น
                Defense: 5,     // เพิ่มค่าป้องกันเริ่มต้น
            },
        },
        Status:    StatusWaiting,
        Actions:   make([]GameAction, 0),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    m.games[game.ID] = game

    if m.onGameUpdate != nil {
        m.onGameUpdate(game.ID, game)
    }

    return game, nil
}

// แก้ไขฟังก์ชัน JoinGame
func (m *GameManager) JoinGame(gameID string, playerID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    game, exists := m.games[gameID]
    if !exists {
        return ErrGameNotFound
    }

    if game.Status != StatusWaiting {
        return ErrGameNotJoinable
    }

    for _, p := range game.Players {
        if p.ID == playerID {
            return ErrPlayerAlreadyInGame
        }
    }

    game.Players = append(game.Players, &Player{
        ID:      playerID,
        Health:  100,
        Gold:    0,
        Level:   1,
        Attack:  10,    // เพิ่มค่าโจมตีเริ่มต้น
        Defense: 5,     // เพิ่มค่าป้องกันเริ่มต้น
    })

    if len(game.Players) >= 2 {
        game.Status = StatusPlaying
    }

    game.UpdatedAt = time.Now()

    if m.onGameUpdate != nil {
        m.onGameUpdate(game.ID, game)
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
            m.onGameUpdate(game.ID, game)
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

    // ทำความสะอาดเกมที่ไม่ได้ใช้งาน
    m.cleanupInactiveGames()

    waitingGames := make([]*Game, 0)
    for _, game := range m.games {
        if game.Status == StatusWaiting && len(game.Players) < 2 {
            waitingGames = append(waitingGames, game)
        }
    }
    return waitingGames
}

// เพิ่มเมธอดสำหรับ auto matching
func (m *GameManager) AutoMatch(playerID string) (*Game, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // ทำความสะอาดเกมที่ไม่ได้ใช้งาน
    m.cleanupInactiveGames()

    // ตรวจสอบว่าผู้เล่นอยู่ในเกมอื่นหรือไม่
    for _, game := range m.games {
        for _, p := range game.Players {
            if p.ID == playerID {
                return nil, ErrPlayerAlreadyInGame
            }
        }
    }

    // หาเกมที่กำลังรอ
    var waitingGame *Game
    for _, game := range m.games {
        if game.Status == StatusWaiting && len(game.Players) < 2 {
            waitingGame = game
            break
        }
    }

    // ถ้าไม่มีเกมที่รอ ให้สร้างเกมใหม่
    if waitingGame == nil {
        waitingGame = &Game{
            ID: generateGameID(),
            Players: []*Player{
                {
                    ID:      playerID,
                    Health:  100,
                    Gold:    0,
                    Level:   1,
                    Attack:  10,
                    Defense: 5,
                },
            },
            Status:    StatusWaiting,
            Actions:   make([]GameAction, 0),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }
        m.games[waitingGame.ID] = waitingGame
    } else {
        // เข้าร่วมเกมที่รออยู่
        waitingGame.Players = append(waitingGame.Players, &Player{
            ID:      playerID,
            Health:  100,
            Gold:    0,
            Level:   1,
            Attack:  10,
            Defense: 5,
        })
        waitingGame.Status = StatusPlaying
        waitingGame.UpdatedAt = time.Now()
    }

    if m.onGameUpdate != nil {
        m.onGameUpdate(waitingGame.ID, waitingGame)
    }

    return waitingGame, nil
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
