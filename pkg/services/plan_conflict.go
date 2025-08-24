package services

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"goname/internal/models"
	"goname/pkg/log"

	"go.uber.org/zap"
)

// PlanConflictResolver handles conflict resolution for rename plans
type PlanConflictResolver struct {
	strategy ConflictStrategy
}

// NewPlanConflictResolver creates a new plan conflict resolver
func NewPlanConflictResolver(strategy ConflictStrategy) *PlanConflictResolver {
	return &PlanConflictResolver{
		strategy: strategy,
	}
}

// ResolvePlanConflicts resolves all conflicts in a plan using the configured strategy
func (pcr *PlanConflictResolver) ResolvePlanConflicts(plan *models.Plan) error {
	if plan.Resolved {
		return fmt.Errorf("plan is already resolved")
	}

	for i := range plan.Conflicts {
		conflict := &plan.Conflicts[i]
		if !conflict.Resolved {
			if err := pcr.resolveConflict(plan, conflict); err != nil {
				log.Error("failed to resolve conflict", zap.Error(err), zap.String("conflict_id", conflict.ID))
				return fmt.Errorf("failed to resolve conflict %s: %w", conflict.ID, err)
			}
		}
	}

	// Update plan status
	pcr.updatePlanStatus(plan)
	return nil
}

// resolveConflict resolves a single conflict
func (pcr *PlanConflictResolver) resolveConflict(plan *models.Plan, conflict *models.Conflict) error {
	switch conflict.ConflictType {
	case models.ConflictTypeMultipleSource:
		return pcr.resolveMultipleSourceConflict(plan, conflict)
	case models.ConflictTypeTargetExists:
		return pcr.resolveTargetExistsConflict(plan, conflict)
	default:
		return fmt.Errorf("unknown conflict type: %s", conflict.ConflictType)
	}
}

// resolveMultipleSourceConflict resolves conflicts where multiple files want the same target
func (pcr *PlanConflictResolver) resolveMultipleSourceConflict(plan *models.Plan, conflict *models.Conflict) error {
	modifications := make(map[string]string)

	// Find the operations involved in this conflict
	operations := make([]*models.PlannedOperation, 0, len(conflict.OperationIDs))
	for i := range plan.Operations {
		op := &plan.Operations[i]
		for _, opID := range conflict.OperationIDs {
			if op.ID == opID {
				operations = append(operations, op)
				break
			}
		}
	}

	if len(operations) == 0 {
		return fmt.Errorf("no operations found for conflict")
	}

	switch pcr.strategy {
	case SkipConflict:
		// Skip all but the first operation
		for i := 1; i < len(operations); i++ {
			operations[i].Status = models.OperationStatusSkipped
			operations[i].Error = "skipped due to conflict"
		}

	case AppendNumber, AppendTimestamp:
		// First operation gets the original target, others get modified names
		for i := 1; i < len(operations); i++ {
			op := operations[i]
			newTarget, err := pcr.generateAlternativeTarget(op.TargetPath, pcr.strategy, i)
			if err != nil {
				return fmt.Errorf("failed to generate alternative target for operation %s: %w", op.ID, err)
			}
			op.TargetPath = newTarget
			op.TargetName = filepath.Base(newTarget)
			modifications[op.ID] = newTarget
		}

	case Overwrite:
		// This doesn't make sense for multiple source conflicts - treat as skip
		for i := 1; i < len(operations); i++ {
			operations[i].Status = models.OperationStatusSkipped
			operations[i].Error = "skipped due to conflict (overwrite strategy)"
		}

	default:
		return fmt.Errorf("unsupported strategy for multiple source conflict: %d", pcr.strategy)
	}

	// Mark conflict as resolved
	conflict.Resolved = true
	conflict.Resolution = models.ConflictResolution{
		Strategy:      pcr.getStrategyName(),
		Modifications: modifications,
		Timestamp:     time.Now(),
	}

	return nil
}

// resolveTargetExistsConflict resolves conflicts where target file already exists
func (pcr *PlanConflictResolver) resolveTargetExistsConflict(plan *models.Plan, conflict *models.Conflict) error {
	if len(conflict.OperationIDs) != 1 {
		return fmt.Errorf("target exists conflict should have exactly one operation, got %d", len(conflict.OperationIDs))
	}

	// Find the operation
	var operation *models.PlannedOperation
	for i := range plan.Operations {
		op := &plan.Operations[i]
		if op.ID == conflict.OperationIDs[0] {
			operation = op
			break
		}
	}

	if operation == nil {
		return fmt.Errorf("operation not found for conflict")
	}

	modifications := make(map[string]string)

	switch pcr.strategy {
	case SkipConflict:
		operation.Status = models.OperationStatusSkipped
		operation.Error = "skipped due to existing target file"

	case AppendNumber, AppendTimestamp:
		newTarget, err := pcr.generateAlternativeTarget(operation.TargetPath, pcr.strategy, 1)
		if err != nil {
			return fmt.Errorf("failed to generate alternative target: %w", err)
		}
		operation.TargetPath = newTarget
		operation.TargetName = filepath.Base(newTarget)
		modifications[operation.ID] = newTarget

	case Overwrite:
		// Keep original target, file will be overwritten
		// No changes needed to operation

	default:
		return fmt.Errorf("unsupported strategy for target exists conflict: %d", pcr.strategy)
	}

	// Mark conflict as resolved
	conflict.Resolved = true
	conflict.Resolution = models.ConflictResolution{
		Strategy:      pcr.getStrategyName(),
		Modifications: modifications,
		Timestamp:     time.Now(),
	}

	return nil
}

// generateAlternativeTarget generates an alternative target path using the specified strategy
func (pcr *PlanConflictResolver) generateAlternativeTarget(originalPath string, strategy ConflictStrategy, index int) (string, error) {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(originalPath, ext)

	switch strategy {
	case AppendNumber:
		return fmt.Sprintf("%s (%d)%s", base, index, ext), nil
	case AppendTimestamp:
		timestamp := time.Now().Format("20060102_150405")
		return fmt.Sprintf("%s_%s%s", base, timestamp, ext), nil
	default:
		return "", fmt.Errorf("unsupported strategy: %d", strategy)
	}
}

// updatePlanStatus updates the overall plan status after conflict resolution
func (pcr *PlanConflictResolver) updatePlanStatus(plan *models.Plan) {
	allResolved := true
	for _, conflict := range plan.Conflicts {
		if !conflict.Resolved {
			allResolved = false
			break
		}
	}

	if allResolved {
		plan.Resolved = true
		// Update operation statuses from conflicted to ready where applicable
		for i := range plan.Operations {
			op := &plan.Operations[i]
			if op.Status == models.OperationStatusConflicted && op.Error == "" {
				op.Status = models.OperationStatusReady
			}
		}
	}

	// Recalculate summary
	plan.Summary = pcr.calculateSummary(plan)
}

// calculateSummary recalculates the plan summary
func (pcr *PlanConflictResolver) calculateSummary(plan *models.Plan) models.PlanSummary {
	summary := models.PlanSummary{
		TotalOperations: len(plan.Operations),
		TotalConflicts:  len(plan.Conflicts),
	}

	for _, op := range plan.Operations {
		switch op.Status {
		case models.OperationStatusReady:
			summary.ReadyOperations++
		case models.OperationStatusConflicted:
			summary.ConflictedOperations++
		case models.OperationStatusSkipped:
			summary.SkippedOperations++
		case models.OperationStatusError:
			summary.ErrorOperations++
		}
	}

	for _, conflict := range plan.Conflicts {
		if conflict.Resolved {
			summary.ResolvedConflicts++
		}
	}

	return summary
}

// getStrategyName returns a string representation of the conflict strategy
func (pcr *PlanConflictResolver) getStrategyName() string {
	switch pcr.strategy {
	case SkipConflict:
		return "skip"
	case AppendNumber:
		return "append_number"
	case AppendTimestamp:
		return "append_timestamp"
	case PromptUser:
		return "prompt_user"
	case Overwrite:
		return "overwrite"
	default:
		return "unknown"
	}
}

// SetStrategy changes the conflict resolution strategy
func (pcr *PlanConflictResolver) SetStrategy(strategy ConflictStrategy) {
	pcr.strategy = strategy
}

// GetStrategy returns the current conflict resolution strategy
func (pcr *PlanConflictResolver) GetStrategy() ConflictStrategy {
	return pcr.strategy
}
