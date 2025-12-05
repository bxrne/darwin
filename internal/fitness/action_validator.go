package fitness

import (
	"fmt"
)

// ActionValidator validates actions based on game observations
type ActionValidator struct {
	gridWidth  int
	gridHeight int
}

// NewActionValidator creates a new action validator
func NewActionValidator() *ActionValidator {
	return &ActionValidator{}
}

// validateBasicAction performs basic validation without game state
func (av *ActionValidator) validateBasicAction(action []int) bool {
	// Check basic action format
	if len(action) != 5 {
		return false
	}

	// pass_turn (0 or 1)
	if action[0] < 0 || action[0] > 1 {
		return false
	}

	// target_x and target_y should be non-negative
	if action[1] < 0 || action[2] < 0 {
		return false
	}

	// split (0 or 1)
	if action[4] < 0 || action[4] > 1 {
		return false
	}

	return true
}

// ValidateAction validates a single action and returns whether it's valid
func (av *ActionValidator) ValidateAction(action []int, observation map[string]any) bool {
	if len(action) != 5 {
		return false
	}

	// If no observation info provided, fall back to basic validation
	if len(observation) == 0 {
		return av.validateBasicAction(action)
	}

	// Extract grid info from observation
	// observation["info"]["info"] contains the mountain grid
	outerInfo, ok := observation["info"].(map[string]any)
	if !ok {
		// If no info structure, fall back to basic validation
		return av.validateBasicAction(action)
	}

	// Extract mountain data from nested info
	mountains, ok := outerInfo["info"].([]any)
	if !ok {
		// If no mountain data, fall back to basic validation
		return av.validateBasicAction(action)
	}

	av.setGridDimensions(mountains)
	return av.validateGameAction(action, mountains, observation)
}

// validateGameAction validates action based on current game state
func (av *ActionValidator) validateGameAction(action []int, mountains []any, observation map[string]any) bool {
	// If passing turn, only need to check pass_turn flag
	if action[0] > 0 {
		// When passing, other action components should be 0 or ignored
		return action[4] >= 0 // split can be 0 or 1, but should be valid
	}

	// For movement actions, validate target coordinates
	targetX := action[1]
	targetY := action[2]

	// Check bounds
	if targetX < 0 || targetX >= av.gridWidth || targetY < 0 || targetY >= av.gridHeight {
		return false
	}

	// Check if target is a mountain
	if av.isMountain(targetX, targetY, mountains) {
		return false
	}

	// Additional validation based on owned cells
	ownedCells, ok := observation["owned_cells"].([]any)
	if ok {
		// Can only move to adjacent cells from owned territory
		if !av.isAdjacentToOwned(targetX, targetY, ownedCells) {
			return false
		}
	}

	return true
}

// setGridDimensions extracts grid dimensions from mountain data
func (av *ActionValidator) setGridDimensions(mountains []any) {
	if av.gridWidth > 0 && av.gridHeight > 0 {
		return // Already set
	}

	if len(mountains) > 0 {
		// Handle different possible row types
		switch row := mountains[0].(type) {
		case []any:
			av.gridWidth = len(row)
			av.gridHeight = len(mountains)
		case []bool:
			av.gridWidth = len(row)
			av.gridHeight = len(mountains)
		default:
			// Try to convert to []any
			if rowSlice, ok := mountains[0].([]any); ok {
				av.gridWidth = len(rowSlice)
				av.gridHeight = len(mountains)
			}
		}
	}
}

// isMountain checks if a coordinate contains a mountain
func (av *ActionValidator) isMountain(x, y int, mountains []any) bool {
	if y < 0 || y >= len(mountains) {
		return false
	}

	// Handle different row types
	switch row := mountains[y].(type) {
	case []any:
		if x < 0 || x >= len(row) {
			return false
		}
		mountain, ok := row[x].(bool)
		return ok && mountain
	case []bool:
		if x < 0 || x >= len(row) {
			return false
		}
		return row[x]
	default:
		return false
	}
}

// isAdjacentToOwned checks if a position is adjacent to owned cells
func (av *ActionValidator) isAdjacentToOwned(x, y int, ownedCells []any) bool {
	if len(ownedCells) == 0 {
		return true // No owned cells data, allow movement
	}

	// Check if any adjacent cell is owned
	directions := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Up, Down, Left, Right
	}

	for _, dir := range directions {
		adjX, adjY := x+dir[0], y+dir[1]

		// Check bounds
		if adjX >= 0 && adjX < av.gridWidth && adjY >= 0 && adjY < av.gridHeight {
			if av.isOwned(adjX, adjY, ownedCells) {
				return true
			}
		}
	}

	return false
}

// isOwned checks if a coordinate is owned by the player
func (av *ActionValidator) isOwned(x, y int, ownedCells []any) bool {
	if y < 0 || y >= len(ownedCells) {
		return false
	}

	// Handle different row types
	switch row := ownedCells[y].(type) {
	case []any:
		if x < 0 || x >= len(row) {
			return false
		}
		owned, ok := row[x].(bool)
		return ok && owned
	case []bool:
		if x < 0 || x >= len(row) {
			return false
		}
		return row[x]
	default:
		return false
	}
}

// CreateActionMask creates a binary mask for valid actions
// Returns a mask where true means action is valid
func (av *ActionValidator) CreateActionMask(observation map[string]any) [][]bool {
	// Extract grid dimensions from observation
	outerInfo, ok := observation["info"].(map[string]any)
	if ok {
		if mountains, ok := outerInfo["info"].([]any); ok {
			av.setGridDimensions(mountains)
		}
	}

	mask := make([][]bool, 5) // 5 action components

	// Component 0: pass_turn (always valid as 0 or 1)
	mask[0] = []bool{true, true}

	// Component 1: target_x (valid range depends on grid width)
	if av.gridWidth > 0 {
		mask[1] = make([]bool, av.gridWidth)
		for i := range mask[1] {
			mask[1][i] = true
		}
	} else {
		mask[1] = []bool{true} // Default if grid size unknown
	}

	// Component 2: target_y (valid range depends on grid height)
	if av.gridHeight > 0 {
		mask[2] = make([]bool, av.gridHeight)
		for i := range mask[2] {
			mask[2][i] = true
		}
	} else {
		mask[2] = []bool{true} // Default if grid size unknown
	}

	// Component 3: unused (always 0)
	mask[3] = []bool{true}

	// Component 4: split (always valid as 0 or 1)
	mask[4] = []bool{true, true}

	return mask
}

// ValidateActionWithMask validates an action using a pre-computed mask
func (av *ActionValidator) ValidateActionWithMask(action []int, mask [][]bool) bool {
	if len(action) != 5 || len(mask) != 5 {
		return false
	}

	for i, component := range action {
		if i < len(mask) {
			// Check if component value is within valid range for this mask
			if component >= 0 && component < len(mask[i]) {
				if !mask[i][component] {
					return false
				}
			} else {
				return false // Component value out of mask range
			}
		}
	}

	return true
}

// FilterInvalidActions applies validation to a batch of actions and returns valid ones
func (av *ActionValidator) FilterInvalidActions(actions [][]int, observation map[string]any) [][]int {
	var validActions [][]int

	for _, action := range actions {
		if av.ValidateAction(action, observation) {
			validActions = append(validActions, action)
		}
	}

	return validActions
}

// GetInvalidActionReason returns a human-readable reason why an action is invalid
func (av *ActionValidator) GetInvalidActionReason(action []int, observation map[string]any) string {
	if len(action) != 5 {
		return "Action must have exactly 5 components"
	}

	// Basic format validation
	if action[0] < 0 || action[0] > 1 {
		return "pass_turn must be 0 or 1"
	}
	if action[4] < 0 || action[4] > 1 {
		return "split must be 0 or 1"
	}

	// Extract game state for detailed validation
	info, ok := observation["info"].(map[string]any)
	if !ok {
		return "" // No detailed info available
	}

	mountains, ok := info["info"].([]any)
	if !ok {
		return "" // No mountain data available
	}

	av.setGridDimensions(mountains)

	// If not passing turn, validate movement
	if action[0] == 0 {
		targetX, targetY := action[1], action[2]

		if targetX < 0 || targetX >= av.gridWidth || targetY < 0 || targetY >= av.gridHeight {
			return fmt.Sprintf("Target coordinates (%d,%d) out of bounds", targetX, targetY)
		}

		if av.isMountain(targetX, targetY, mountains) {
			return fmt.Sprintf("Target coordinates (%d,%d) contain a mountain", targetX, targetY)
		}

		ownedCells, ok := observation["owned_cells"].([]any)
		if ok && !av.isAdjacentToOwned(targetX, targetY, ownedCells) {
			return fmt.Sprintf("Target coordinates (%d,%d) not adjacent to owned territory", targetX, targetY)
		}
	}

	return "" // Action is valid
}
