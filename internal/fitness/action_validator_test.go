package fitness_test

import (
	"github.com/bxrne/darwin/internal/fitness"
	"testing"
)

func TestActionValidator_BasicValidation(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Test valid action
	validAction := []int{0, 5, 3, 0, 1}
	observation := map[string]any{}

	if !validator.ValidateAction(validAction, observation) {
		t.Error("Valid action should pass validation")
	}

	// Test invalid action - wrong length
	invalidAction := []int{0, 5, 3}
	if validator.ValidateAction(invalidAction, observation) {
		t.Error("Action with wrong length should fail validation")
	}

	// Test invalid action - invalid pass_turn value
	invalidAction2 := []int{2, 5, 3, 0, 1}
	if validator.ValidateAction(invalidAction2, observation) {
		t.Error("Action with invalid pass_turn should fail validation")
	}

	// Test invalid action - invalid split value
	invalidAction3 := []int{0, 5, 3, 0, 2}
	if validator.ValidateAction(invalidAction3, observation) {
		t.Error("Action with invalid split should fail validation")
	}
}

func TestActionValidator_MountainValidation(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Create observation with mountain data - match actual format from logs
	mountains := [][]bool{
		{false, false, true},
		{false, true, false},
		{false, false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
	}

	// Test action targeting mountain - should fail
	mountainAction := []int{0, 2, 0, 0, 0} // target (2,0) is a mountain
	if validator.ValidateAction(mountainAction, info) {
		t.Error("Action targeting mountain should fail validation")
	}

	// Test action targeting empty cell - should pass
	emptyAction := []int{0, 1, 0, 0, 0} // target (1,0) is empty
	if !validator.ValidateAction(emptyAction, info) {
		t.Error("Action targeting empty cell should pass validation")
	}

	// Test pass turn action - should always pass regardless of target
	passAction := []int{1, 2, 0, 0, 0} // pass turn to mountain location
	if !validator.ValidateAction(passAction, info) {
		t.Error("Pass turn action should always pass validation")
	}
}

func TestActionValidator_BoundsValidation(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Create observation with grid info
	mountains := [][]bool{
		{false, false},
		{false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
	}

	// Test action out of bounds - should fail
	outOfBoundsAction := []int{0, 5, 5, 0, 0} // target (5,5) is out of bounds
	if validator.ValidateAction(outOfBoundsAction, info) {
		t.Error("Action out of bounds should fail validation")
	}

	// Test action within bounds - should pass
	inBoundsAction := []int{0, 1, 1, 0, 0} // target (1,1) is within bounds
	if !validator.ValidateAction(inBoundsAction, info) {
		t.Error("Action within bounds should pass validation")
	}
}

func TestActionValidator_AdjacentValidation(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Create observation with owned cells
	mountains := [][]bool{
		{false, false, false},
		{false, false, false},
		{false, false, false},
	}

	ownedCells := [][]bool{
		{true, false, false},
		{false, false, false},
		{false, false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	ownedCellsInterface := make([]any, len(ownedCells))
	for i, row := range ownedCells {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		ownedCellsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
		"owned_cells": ownedCellsInterface,
	}

	// Test action adjacent to owned cell - should pass
	adjacentAction := []int{0, 1, 0, 0, 0} // target (1,0) is adjacent to (0,0)
	if !validator.ValidateAction(adjacentAction, info) {
		t.Error("Action adjacent to owned cell should pass validation")
	}

	// Test action not adjacent to owned cell - should fail
	nonAdjacentAction := []int{0, 2, 2, 0, 0} // target (2,2) is not adjacent to owned cells
	if validator.ValidateAction(nonAdjacentAction, info) {
		t.Error("Action not adjacent to owned cell should fail validation")
	}
}

func TestActionValidator_CreateActionMask(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Set grid dimensions
	mountains := [][]bool{
		{false, false, false},
		{false, false, false},
		{false, false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
	}

	// Create action mask
	mask := validator.CreateActionMask(info)

	// Check mask structure
	if len(mask) != 5 {
		t.Errorf("Expected mask length 5, got %d", len(mask))
	}

	// Check pass_turn mask (should have 2 options: 0, 1)
	if len(mask[0]) != 2 {
		t.Errorf("Expected pass_turn mask length 2, got %d", len(mask[0]))
	}

	// Check target_x mask (should match grid width)
	if len(mask[1]) != 3 {
		t.Errorf("Expected target_x mask length 3, got %d", len(mask[1]))
	}

	// Check target_y mask (should match grid height)
	if len(mask[2]) != 3 {
		t.Errorf("Expected target_y mask length 3, got %d", len(mask[2]))
	}

	// Check unused mask (should have 1 option: 0)
	if len(mask[3]) != 1 {
		t.Errorf("Expected unused mask length 1, got %d", len(mask[3]))
	}

	// Check split mask (should have 2 options: 0, 1)
	if len(mask[4]) != 2 {
		t.Errorf("Expected split mask length 2, got %d", len(mask[4]))
	}
}

func TestActionValidator_FilterInvalidActions(t *testing.T) {
	validator := fitness.NewActionValidator()

	// Create observation with mountain
	mountains := [][]bool{
		{false, true, false},
		{false, false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
	}

	// Mix of valid and invalid actions
	actions := [][]int{
		{0, 0, 0, 0, 0}, // valid
		{0, 1, 0, 0, 0}, // invalid (mountain)
		{0, 2, 0, 0, 0}, // valid
		{1, 1, 0, 0, 0}, // valid (pass turn)
	}

	// Filter invalid actions
	validActions := validator.FilterInvalidActions(actions, info)

	// Debug: print what actions are considered valid
	t.Logf("Valid actions: %v", validActions)

	// Should have 3 valid actions (2 valid moves + 1 pass turn)
	if len(validActions) != 3 {
		t.Errorf("Expected 3 valid actions, got %d", len(validActions))
	}

	// Check that mountain action was filtered out (action with pass_turn=0 targeting mountain)
	foundMountainAction := false
	for _, action := range validActions {
		if action[0] == 0 && action[1] == 1 && action[2] == 0 { // pass_turn=0, target=(1,0)
			foundMountainAction = true
			break
		}
	}
	if foundMountainAction {
		t.Error("Mountain action should have been filtered out")
	}

	// Verify pass action is still valid
	foundPassAction := false
	for _, action := range validActions {
		if action[0] == 1 { // pass_turn=1
			foundPassAction = true
			break
		}
	}
	if !foundPassAction {
		t.Error("Pass action should be valid")
	}
}

func TestActionValidator_GetInvalidActionReason(t *testing.T) {
	validator := fitness.NewActionValidator()

	mountains := [][]bool{
		{false, true, false},
		{false, false, false},
	}

	// Convert to any slice to match actual observation format
	mountainsInterface := make([]any, len(mountains))
	for i, row := range mountains {
		rowInterface := make([]any, len(row))
		for j, val := range row {
			rowInterface[j] = val
		}
		mountainsInterface[i] = rowInterface
	}

	info := map[string]any{
		"info": map[string]any{
			"info": mountainsInterface,
		},
	}

	// Test wrong length
	reason := validator.GetInvalidActionReason([]int{0, 1, 2}, info)
	if reason != "Action must have exactly 5 components" {
		t.Errorf("Expected length error, got: %s", reason)
	}

	// Test invalid pass_turn
	reason = validator.GetInvalidActionReason([]int{2, 1, 2, 0, 0}, info)
	if reason != "pass_turn must be 0 or 1" {
		t.Errorf("Expected pass_turn error, got: %s", reason)
	}

	// Test mountain targeting
	reason = validator.GetInvalidActionReason([]int{0, 1, 0, 0, 0}, info)
	expected := "Target coordinates (1,0) contain a mountain"
	if reason != expected {
		t.Errorf("Expected mountain error, got: %s", reason)
	}

	// Test valid action should return empty reason
	reason = validator.GetInvalidActionReason([]int{0, 0, 0, 0, 0}, info)
	if reason != "" {
		t.Errorf("Expected empty reason for valid action, got: %s", reason)
	}
}
