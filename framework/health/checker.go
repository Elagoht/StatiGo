package health

import (
	"context"
	"sync"
	"time"
)

// CheckResult represents the result of a health check.
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthStatus represents overall health status.
type HealthStatus struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks,omitempty"`
}

// Checker performs health checks on external dependencies.
type Checker struct {
	timeout time.Duration
	checks  []CheckFunc
}

// CheckFunc is a function that performs a health check.
type CheckFunc func(ctx context.Context) CheckResult

// NewChecker creates a new health checker.
func NewChecker(timeout time.Duration) *Checker {
	return &Checker{
		timeout: timeout,
		checks:  make([]CheckFunc, 0),
	}
}

// AddCheck registers a new health check.
func (c *Checker) AddCheck(check CheckFunc) {
	c.checks = append(c.checks, check)
}

// CheckAll runs all health checks in parallel.
func (c *Checker) CheckAll(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Status: "up",
		Checks: make([]CheckResult, 0),
	}

	// If no checks registered, return healthy
	if len(c.checks) == 0 {
		return status
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var wg sync.WaitGroup
	resultsChan := make(chan CheckResult, len(c.checks))

	// Run all checks in parallel
	for _, check := range c.checks {
		wg.Add(1)
		go func(checkFn CheckFunc) {
			defer wg.Done()
			resultsChan <- checkFn(timeoutCtx)
		}(check)
	}

	// Close results channel when all checks complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		status.Checks = append(status.Checks, result)
		if result.Status != "up" {
			status.Status = "degraded"
		}
	}

	return status
}
