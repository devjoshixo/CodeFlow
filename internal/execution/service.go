package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ExecutionService struct {
	executionRepo ExecutionRepository
	redis         *redis.Client
}

func NewExecutionService(executionRepo ExecutionRepository, redis *redis.Client) *ExecutionService {
	return &ExecutionService{
		executionRepo: executionRepo,
		redis:         redis,
	}
}

func (s *ExecutionService) Submit(ctx context.Context, execution *Execution) (*Execution, error) {
	if err := execution.Validate(); err != nil {
		return nil, fmt.Errorf("error in validating code: %w", err)
	}

	newExec := &Execution{
		ID:         uuid.New().String(),
		UserID:     execution.UserID,
		Language:   execution.Language,
		Code:       execution.Code,
		Status:     StatusPending,
		ExitCode:   1,
		DurationMs: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.executionRepo.Create(ctx, newExec); err != nil {
		return nil, fmt.Errorf("failed to insert the execution: %w", err)
	}

	// push to Redis list
	if err := s.redis.LPush(ctx, "executions:queue", newExec.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to push execution id to redis:  %w", err)
	}
	return newExec, nil

}

func (s *ExecutionService) GetByID(ctx context.Context, id string, userId string) (*Execution, error) {
	execution, err := s.executionRepo.FindExecutionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching executions: %w", err)
	}

	if execution.UserID != userId {
		return nil, fmt.Errorf("user id does not belong to the user: %w", err)
	}

	return execution, nil
}

func (s *ExecutionService) GetByUser(ctx context.Context, userID string) ([]*Execution, error) {
	executions, err := s.executionRepo.FindExecutionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error in finding your submissions")
	}

	return executions, nil
}

func (s *ExecutionService) UpdateExecution(ctx context.Context, id string, status Status, output string, errMsg string, exitCode int, durationMs int) error {
	execution, err := s.executionRepo.FindExecutionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching executions: %w", err)
	}

	if !execution.Status.CanTransition(status) {
		return ErrInvalidStatusTransition
	}

	execution.Status = status
	execution.Output = &output
	execution.Error = &errMsg
	execution.ExitCode = exitCode
	execution.DurationMs = durationMs
	execution.UpdatedAt = time.Now()

	if err := s.executionRepo.UpdateExecutionStatus(ctx, execution); err != nil {
		return fmt.Errorf("failed to update the execution results: %w", err)
	}

	return nil
}
