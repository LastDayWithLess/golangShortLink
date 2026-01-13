package main

import (
	"os"
	"os/signal"
	"short_link/config"
	"short_link/internal/logger"
	"short_link/internal/repository/database"
	"short_link/internal/worker"
	"syscall"
	"time"
)

func main() {
	workerLogger, err := logger.NewLogger("logs/worker.log")
	if err != nil {
		workerLogger.Error("failed to create logger:", logger.ErrorField(err))
	}
	defer workerLogger.Close()

	workerLogger.Info("starting link cleaner worker")

	config.Init()
	cfg := config.LoadDBConfig()
	pool, err := database.NewConnection(cfg)

	if err != nil {
		workerLogger.Error(err.Error())
	}

	defer pool.CloseDB()

	var r *database.LinkRepository = database.NewLinkRepository(pool.GetDB())

	var c *worker.Cleaner = worker.NewCleaner(r, workerLogger, worker.CleanerConfig{
		Interval:  1 * time.Hour,
		BatchSize: 1000,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go c.Start()

	<-quit
	workerLogger.Info("shutting down worker...")

	c.Stop()

	workerLogger.Info("worker exited properly")
}
