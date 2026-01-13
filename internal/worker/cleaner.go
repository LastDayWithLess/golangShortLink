package worker

import (
	"context"
	"short_link/internal/logger"
	"short_link/internal/repository/database"
	"time"
)

type CleanerConfig struct {
	Interval  time.Duration
	BatchSize int
}

type Cleaner struct {
	repo     *database.LinkRepository
	logger   *logger.Logger
	config   CleanerConfig
	stopChan chan struct{}
}

func NewCleaner(repo *database.LinkRepository, logger *logger.Logger, config CleanerConfig) *Cleaner {
	return &Cleaner{
		repo:     repo,
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (c *Cleaner) Start() {
	c.logger.Info("starting link cleaner worker...")

	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.runCleanup()
		case <-c.stopChan:
			c.logger.Info("worker stoped")
			return
		}
	}
}

func (c *Cleaner) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	c.logger.Info("Starting cleanup")

	tx, err := c.repo.BeginTx(ctx)
	if err != nil {
		c.logger.Error("Failed to begin transaction",
			logger.ErrorField(err))
		return
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				c.logger.Error("Failed to rollback transaction",
					logger.ErrorField(rbErr))
			}
		}
	}()

	expLinks, err := c.repo.FindExpiredLinks(ctx, c.config.BatchSize, tx)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	err = c.repo.DeleteExpiredShortLinks(ctx, expLinks, tx)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	err = c.repo.DeleteExpiredOriginalLinks(ctx, tx)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		c.logger.Error("Failed to commit transaction",
			logger.ErrorField(err))
	}

	c.logger.Info("cleanup finished")
}

func (c *Cleaner) Stop() {
	close(c.stopChan)
}
