package main

import (
	"short_link/config"
	"short_link/internal/logger"
	"short_link/internal/repository/cache"
	"short_link/internal/repository/database"
	"short_link/internal/service"
	"short_link/internal/transport/rest"
)

func main() {
	logger, err := logger.NewLogger("logs/app.log")
	if err != nil {
		logger.Error(err.Error())
	}
	defer logger.Close()

	config.Init()
	cfg := config.LoadDBConfig()
	pool, err := database.NewConnection(cfg)

	if err != nil {
		logger.Error(err.Error())
	}

	defer pool.CloseDB()

	cfgRedis := config.LoadRedisConfig()
	rdb, err := cache.NewRedisConnect(cfgRedis)

	if err != nil {
		logger.Error(err.Error())
	}

	var r *database.LinkRepository = database.NewLinkRepository(pool.GetDB())
	var s *service.LinkService = service.NewLinkService(r, rdb, logger)
	var h *rest.HTTPHandler = rest.NewHTTPHanler(s)
	var server *rest.HTTPServer = rest.NewServer(h)

	if err := server.StartServer(); err != nil {
		logger.Error(err.Error())
	}

}
