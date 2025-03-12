package game

import (
    "testing"
)

func TestNewGame(t *testing.T) {
    gameID := "test-game-1"
    game := NewGame(gameID)

    if game.ID != gameID {
        t.Errorf("expected game ID %s, got %s", gameID, game.ID)
    }

    if game.State != StateWaiting {
        t.Errorf("expected initial state %s, got %s", StateWaiting, game.State)
    }

    if game.Round != 0 {
        t.Errorf("expected initial round 0, got %d", game.Round)
    }

    if len(game.Players) != 0 {
        t.Errorf("expected empty players map, got %d players", len(game.Players))
    }
}