package domain

// Rating scale: a self-set skill level, stored and displayed as the integer 1–5.
// The neutral median is used for any Player whose rating is unknown, so an
// all-unrated field has zero rating gaps and the scheduler behaves as it did
// before ratings existed. See ADR 0006.
const (
	MinRating    = 1
	MaxRating    = 5
	MedianRating = 3
)

// IsValidRating reports whether r is within the 1–5 rating scale.
func IsValidRating(r int) bool {
	return r >= MinRating && r <= MaxRating
}

// NormalizeRating maps any out-of-range rating (including the zero value of an
// unrated player) to the neutral median. Schedulers use this so an all-unrated
// field has zero rating gaps and behaves exactly as it did before ratings
// existed. See ADR 0006.
func NormalizeRating(r int) int {
	if !IsValidRating(r) {
		return MedianRating
	}
	return r
}

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
				"courts":   courts,
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
