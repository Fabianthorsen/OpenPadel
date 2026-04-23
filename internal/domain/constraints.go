package domain

// ValidationError represents a constraint violation with a structured error code and parameters.
type ValidationError struct {
	Code   string                 `json:"code"`
	Params map[string]interface{} `json:"params"`
}

// AmericanoConstraints validates an Americano session configuration.
// Returns a list of validation errors; empty list means all constraints are satisfied.
func AmericanoConstraints(courts, playerCount int) []ValidationError {
	var errs []ValidationError

	// Courts validation: 1-20
	if courts < 1 || courts > 20 {
		errs = append(errs, ValidationError{
			Code:   "americano_invalid_courts",
			Params: map[string]interface{}{"courts": courts, "min": 1, "max": 20},
		})
	}

	// Players validation: >= courts * 4
	minPlayers := courts * 4
	if playerCount < minPlayers {
		errs = append(errs, ValidationError{
			Code: "americano_insufficient_players",
			Params: map[string]interface{}{
				"required": minPlayers,
				"current":  playerCount,
			},
		})
	}

	return errs
}

// MexicanoConstraints validates a Mexicano session configuration.
// Returns a list of validation errors; empty list means all constraints are satisfied.
func MexicanoConstraints(courts, playerCount int) []ValidationError {
	var errs []ValidationError

	// Courts validation: 1-20
	if courts < 1 || courts > 20 {
		errs = append(errs, ValidationError{
			Code:   "mexicano_invalid_courts",
			Params: map[string]interface{}{"courts": courts, "min": 1, "max": 20},
		})
	}

	// Players validation: exactly courts * 4 (no bench)
	requiredPlayers := courts * 4
	if playerCount != requiredPlayers {
		errs = append(errs, ValidationError{
			Code: "mexicano_player_count_mismatch",
			Params: map[string]interface{}{
				"required": requiredPlayers,
				"current":  playerCount,
			},
		})
	}

	return errs
}
