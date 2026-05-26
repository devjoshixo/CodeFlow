package postgres

import (
	"codeflow/internal/execution"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExecutionRepo struct {
	db *pgxpool.Pool
}

func NewExecutionRepo(db *pgxpool.Pool) *ExecutionRepo {
	return &ExecutionRepo{db: db}
}

func (r *ExecutionRepo) Create(ctx context.Context, execution *execution.Execution) error {
	_, err := r.db.Exec(ctx, "INSERT INTO executions (id, user_id, language, code, status, exit_code, duration_ms, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);", execution.ID, execution.UserID, execution.Language, execution.Code, execution.Status, execution.ExitCode, execution.DurationMs, execution.CreatedAt, execution.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	return nil
}

func (r *ExecutionRepo) MarkQueued(ctx context.Context, id string) error {
	cmdTag, err := r.db.Exec(ctx, "UPDATE executions SET status=$1, updated_at=NOW() WHERE id=$2", execution.StatusQueued, id)
	if err != nil {
		return fmt.Errorf("failed to update execution to queue status")
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("execution not found")
	}
	return nil
}

func (r *ExecutionRepo) FindExecutionByID(ctx context.Context, id string) (*execution.Execution, error) {
	exec := &execution.Execution{}
	row := r.db.QueryRow(ctx, "SELECT id, user_id, language, code, status, output, error, exit_code, duration_ms, created_at, updated_at FROM executions WHERE id=$1;", id)
	err := row.Scan(
		&exec.ID,
		&exec.UserID,
		&exec.Language,
		&exec.Code,
		&exec.Status,
		&exec.Output,
		&exec.Error,
		&exec.ExitCode,
		&exec.DurationMs,
		&exec.CreatedAt,
		&exec.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		fmt.Printf("error is:%v", err)
		return nil, fmt.Errorf("could not find any execution with the given id: %w", err)
	}

	return exec, nil
}

func (r *ExecutionRepo) FindExecutionsByUser(ctx context.Context, userID string) ([]*execution.Execution, error) {
	rows, err := r.db.Query(ctx, "SELECT id, user_id, language, code, status, output, error, exit_code, duration_ms, created_at, updated_at FROM executions WHERE user_id=$1 ORDER BY created_at DESC;", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch executions: %w", err)
	}

	var executions []*execution.Execution

	for rows.Next() {
		exec := &execution.Execution{}

		err := rows.Scan(
			&exec.ID,
			&exec.UserID,
			&exec.Language,
			&exec.Code,
			&exec.Status,
			&exec.Output,
			&exec.Error,
			&exec.ExitCode,
			&exec.DurationMs,
			&exec.CreatedAt,
			&exec.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		executions = append(executions, exec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return executions, nil
}

func (r *ExecutionRepo) UpdateExecutionStatus(ctx context.Context, execution *execution.Execution) error {
	cmdTag, err := r.db.Exec(ctx, "UPDATE executions SET status=$1, output=$2, error=$3, exit_code=$4, duration_ms=$5, updated_at=$6 WHERE id=$7", execution.Status, execution.Output, execution.Error, execution.ExitCode, execution.DurationMs, execution.UpdatedAt, execution.ID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("execution not found")
	}

	return nil
}
