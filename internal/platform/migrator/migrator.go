package migrator

import (
	"codeflow/internal/platform/logger"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(db_uri string) error {
	ctx := context.Background()
	l := logger.Get()

	conn, err := pgxpool.New(ctx, db_uri)

	if err != nil {
		return fmt.Errorf("Unable to connect to the db:%w", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("Could not connect to database on startup: %v", err)
	}

	l.Info("PostGres is connected")

	migrationRunner(conn)

	return nil

}

func migrationRunner(conn *pgxpool.Pool) {
	ctx := context.Background()

	files, err := os.ReadDir("migrations")
	if err != nil {
		fmt.Printf("Got error in reading files:%v", err)
		return
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Fatalf("failed to start transaction: %v", err)
		return
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS migration_files (id SERIAL PRIMARY KEY,filename TEXT UNIQUE NOT NULL,created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);")
	if err != nil {
		log.Fatalf("Failed to run the migration table creation query: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		dup, err := tx.Exec(ctx, "INSERT INTO migration_files (filename) VALUES ($1) ON CONFLICT (filename) DO NOTHING;", file.Name())
		if err != nil {
			log.Fatalf("The migration for the file: %s with error: %s", file.Name(), err)
			return
		}

		if dup.RowsAffected() == 0 {
			fmt.Println("file skipped:" + file.Name())
			continue
		}

		path := filepath.Join("migrations", file.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)

		}

		ext := filepath.Ext(file.Name())
		if ext != ".sql" {
			continue
		}

		_, err = tx.Exec(ctx, string(data))
		if err != nil {
			log.Printf("Error: Failed to commit to migration: %s", err)
			break
		}

		fmt.Println("Inserted:", file.Name())

	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit migration: %v", err)
	}
}
