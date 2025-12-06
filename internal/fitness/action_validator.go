package fitness

import (
	"fmt"
)

// ActionValidator validates actions based on game observations
type ActionValidator struct {
	gridWidth  int
	gridHeight int
	mountains  [][]bool
}

// NewActionValidator creates a new action validator
func NewActionValidator() *ActionValidator {
	return &ActionValidator{}
}

// setGridDimensions extracts grid dimensions from mountain data
func (av *ActionValidator) setGridDimensions(mountains [][]bool) {
	av.gridWidth = len(mountains)
	av.gridHeight = len(mountains[0])
}
func (av *ActionValidator) SetMountains(mountain_array [][]bool) {
	av.mountains = mountain_array
	av.setGridDimensions(mountain_array)
}
func rowHasTrue(row []bool) bool {
	for _, v := range row {
		if v {
			return true
		}
	}
	return false
}
func applyMask(values []float64, mask []float64) {
	for i := range values {
		values[i] *= mask[i] // 0 wipes, 1 keeps
	}
}

// ValidXActionMask Returns all x values that have a potential y value
func (av *ActionValidator) ValidXActionMask(owned_cells [][]bool) ([]float64, error) {
	validXactions := make([]float64, av.gridWidth)
	count := 0
	for i := range av.gridWidth {
		if rowHasTrue(owned_cells[i]) {
			validXactions[i] = 1.0
		} else {
			count += 1
		}
	}

	if count >= av.gridWidth {
		return nil, fmt.Errorf("No valid X actions(impossible)")
	}
	return validXactions, nil
}

// ValidYActionMask Returns all x values that have a potential y value
func (av *ActionValidator) ValidYActionMask(x int, owned_cells [][]bool) ([]float64, error) {
	validYactions := make([]float64, av.gridHeight)
	invalidYaction := 0
	for i := range len(owned_cells[x]) {
		if owned_cells[x][i] {
			validYactions[i] = 1.0
			continue
		}
		invalidYaction += 1
	}

	if invalidYaction >= av.gridHeight {
		return nil, fmt.Errorf("No valid Y actions(impossible)")
	}

	return validYactions, nil
}

func (av *ActionValidator) DirectionActionMask(x, y int) ([]float64, error) {
	validDirectionMask := make([]float64, 4)

	directions := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Up, Down, Left, Right
	}
	invalidCount := 0
	for i, dir := range directions {
		adjX, adjY := x+dir[0], y+dir[1]
		if adjX >= av.gridWidth || adjX < 0 || adjY >= av.gridHeight || adjY < 0 {

			invalidCount += 1
			continue
		}
		if av.mountains[adjX][adjY] {
			invalidCount += 1
			continue
		}
		validDirectionMask[i] = 1.0

	}
	if invalidCount >= 4 {
		return nil, fmt.Errorf("No Valid Directions available(impossible)")
	}
	return validDirectionMask, nil
}

func (av *ActionValidator) SelectValidAction(actionOutputs [][]float64, owned_cells [][]bool) ([]int, error) {

	selectedActions := make([]int, len(actionOutputs))
	probabilities := CalculateSoftmax(actionOutputs[0])
	selectedActions[0] = SampleAction(probabilities)
	if selectedActions[0] == 1 {
		return selectedActions, nil
	}

	Xprobabilities := CalculateSoftmax(actionOutputs[1])
	validXActions, err := av.ValidXActionMask(owned_cells)
	if err != nil {
		return nil, fmt.Errorf("failed to find valid X value", err.Error())
	}
	applyMask(Xprobabilities, validXActions)
	selectedActions[1] = SampleAction(Xprobabilities)

	Yprobabilities := CalculateSoftmax(actionOutputs[2])
	validYActions, err := av.ValidYActionMask(selectedActions[1], owned_cells)
	if err != nil {
		return nil, fmt.Errorf("failed to find valid Y value", err.Error())
	}
	applyMask(Yprobabilities, validYActions)
	selectedActions[2] = SampleAction(Yprobabilities)

	directionProbabilities := CalculateSoftmax(actionOutputs[3])
	validDirections, err := av.DirectionActionMask(selectedActions[1], selectedActions[2])
	if err != nil {
		return nil, fmt.Errorf("failed to find valid Direction value", err.Error())
	}
	applyMask(directionProbabilities, validDirections)
	selectedActions[3] = SampleAction(directionProbabilities)

	splitProbabilities := CalculateSoftmax(actionOutputs[4])
	selectedActions[4] = SampleAction(splitProbabilities)
	return selectedActions, nil
}
