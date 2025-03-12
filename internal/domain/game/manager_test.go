package game

import (
    "testing"
)

func TestGameManager(t *testing.T) {
    t.Run("Create and Get Game", func(t *testing.T) {
        gm := NewGameManager()
        gameID := "test-game-1"

        // Test CreateGame
        game, err := gm.CreateGame(gameID)
        if err != nil {
            t.Fatalf("failed to create game: %v", err)
        }
        if game.ID != gameID {
            t.Errorf("expected game ID %s, got %s", gameID, game.ID)
        }

        // Test GetGame
        fetchedGame, err := gm.GetGame(gameID)
        if err != nil {
            t.Fatalf("failed to get game: %v", err)
        }
        if fetchedGame.ID != gameID {
            t.Errorf("expected game ID %s, got %s", gameID, fetchedGame.ID)
        }
    })

    t.Run("Add Player", func(t *testing.T) {
        gm := NewGameManager()
        gameID := "test-game-1"
        
        // Create game first
        _, err := gm.CreateGame(gameID)
        if err != nil {
            t.Fatalf("failed to create game: %v", err)
        }

        // Test adding player
        player := &Player{
            ID:       "player-1",
            Username: "TestPlayer",
            Health:   100,
            Gold:     0,
            Level:    1,
        }

        err = gm.AddPlayer(gameID, player)
        if err != nil {
            t.Fatalf("failed to add player: %v", err)
        }

        // Verify player was added
        game, _ := gm.GetGame(gameID)
        if len(game.Players) != 1 {
            t.Errorf("expected 1 player, got %d", len(game.Players))
        }

        addedPlayer := game.Players[player.ID]
        if addedPlayer.Username != player.Username {
            t.Errorf("expected username %s, got %s", player.Username, addedPlayer.Username)
        }
    })

    t.Run("Game Not Found", func(t *testing.T) {
        gm := NewGameManager()
        
        _, err := gm.GetGame("non-existent")
        if err == nil {
            t.Error("expected error for non-existent game, got nil")
        }
    })

    t.Run("Duplicate Game Creation", func(t *testing.T) {
        gm := NewGameManager()
        gameID := "test-game-1"

        // Create first game
        _, err := gm.CreateGame(gameID)
        if err != nil {
            t.Fatalf("failed to create first game: %v", err)
        }

        // Try to create duplicate
        _, err = gm.CreateGame(gameID)
        if err == nil {
            t.Error("expected error for duplicate game creation, got nil")
        }
    })
}