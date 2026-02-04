package cache

import (
	"log/slog"
	"time"
)

// Revalidator handles scheduled cache revalidation.
type Revalidator struct {
	manager *Manager
	logger  *slog.Logger
	ticker  *time.Ticker
	done    chan bool
}

// NewRevalidator creates a new revalidator instance.
func NewRevalidator(manager *Manager, logger *slog.Logger) *Revalidator {
	return &Revalidator{
		manager: manager,
		logger:  logger,
		done:    make(chan bool),
	}
}

// Start begins the background revalidation worker.
// Runs daily at the specified hour (0-23).
func (rv *Revalidator) Start(revalidationHour int) {
	rv.logger.Info("starting cache revalidation worker",
		slog.Int("hour", revalidationHour),
	)

	// Calculate time until next revalidation
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), revalidationHour, 0, 0, 0, now.Location())

	// If today's revalidation time has passed, schedule for tomorrow
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	// Wait until first revalidation time
	initialDelay := time.Until(next)

	rv.logger.Info("next cache revalidation scheduled",
		slog.Time("next_run", next),
		slog.Duration("initial_delay", initialDelay),
	)

	// Create ticker for daily revalidation (24 hours)
	rv.ticker = time.NewTicker(24 * time.Hour)

	go func() {
		// Wait for initial delay
		time.Sleep(initialDelay)

		// Run first revalidation
		rv.revalidateIncremental()

		// Continue with ticker for subsequent revalidations
		for {
			select {
			case <-rv.ticker.C:
				rv.revalidateIncremental()
			case <-rv.done:
				rv.logger.Info("cache revalidation worker stopped")
				return
			}
		}
	}()
}

// Stop stops the revalidation worker.
func (rv *Revalidator) Stop() {
	if rv.ticker != nil {
		rv.ticker.Stop()
	}
	close(rv.done)
}

// revalidateIncremental marks all incremental cache entries as stale.
func (rv *Revalidator) revalidateIncremental() {
	rv.logger.Info("starting incremental cache revalidation")

	start := time.Now()
	count := rv.manager.MarkStale("incremental", true)
	duration := time.Since(start)

	rv.logger.Info("incremental cache revalidation completed",
		slog.Int("count", count),
		slog.Duration("duration", duration),
	)
}
