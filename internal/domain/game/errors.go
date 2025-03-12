package game

import "errors"

var (
    ErrGameNotFound        = errors.New("game not found")
    ErrGameNotJoinable     = errors.New("game is not joinable")
    ErrPlayerAlreadyInGame = errors.New("player is already in game")
)