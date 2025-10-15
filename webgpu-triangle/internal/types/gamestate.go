package types

// GameState represents different states of the game
type GameState int

const (
	// TEST is a test game state
	TEST GameState = iota
)

// String returns the string representation of the game state
func (g GameState) String() string {
	switch g {
	case TEST:
		return "TEST"
	default:
		return "UNKNOWN"
	}
}
