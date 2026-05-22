package execution

import "context"

type ExecutionRepository interface {
	Create(ctx context.Context, execution *Execution) error
	FindExecutionByID(ctx context.Context, id string) (*Execution, error)
	FindExecutionsByUser(ctx context.Context, userID string) ([]*Execution, error)
	UpdateExecutionStatus(ctx context.Context, execution *Execution) error
}
