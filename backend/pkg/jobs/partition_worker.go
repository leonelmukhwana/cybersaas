package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type PartitionWorker struct {
	db     *pgxpool.Pool
	cron   *cron.Cron
	logger *slog.Logger
}

func NewPartitionWorker(db *pgxpool.Pool, logger *slog.Logger) *PartitionWorker {
	return &PartitionWorker{
		db: db,
		// Standard 5-field cron parser (Minute, Hour, Day of Month, Month, Day of Week)
		cron:   cron.New(),
		logger: logger,
	}
}

func (w *PartitionWorker) Start(ctx context.Context) error {
	// Run once immediately at startup
	if err := w.EnsurePartitions(ctx); err != nil {
		w.logger.Error("Failed initial session partition check", "error", err)
	}

	// Schedule to run every 1st day of the month at 02:00 AM ("0 2 1 * *")
	_, err := w.cron.AddFunc("0 2 1 * *", func() {
		jobCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := w.EnsurePartitions(jobCtx); err != nil {
			w.logger.Error("Failed scheduled session partition creation", "error", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule partition cron: %w", err)
	}

	w.cron.Start()
	w.logger.Info("Partition maintenance worker started successfully")
	return nil
}

func (w *PartitionWorker) Stop() {
	ctx := w.cron.Stop()
	<-ctx.Done() // Wait for running jobs to finish gracefully
	w.logger.Info("Partition maintenance worker stopped")
}

func (w *PartitionWorker) EnsurePartitions(ctx context.Context) error {
	currentYear := time.Now().UTC().Year()

	// Ensure current year, next year, and year after next are partitioned
	for _, year := range []int{currentYear, currentYear + 1, currentYear + 2} {
		_, err := w.db.Exec(ctx, "SELECT create_yearly_session_partition($1)", year)
		if err != nil {
			return fmt.Errorf("failed creating partition for year %d: %w", year, err)
		}
	}

	w.logger.Info("Verified/created session partitions", "target_years", []int{currentYear, currentYear + 1, currentYear + 2})
	return nil
}
