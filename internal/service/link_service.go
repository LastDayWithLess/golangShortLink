package service

import (
	"context"
	"database/sql"
	"fmt"
	"short_link/internal/logger"
	"short_link/internal/model"
	"short_link/internal/repository/cache"
	"short_link/internal/repository/database"
	"sync"
	"time"
)

type LinkService struct {
	repo   *database.LinkRepository
	cache  *cache.RedisClient
	Logger *logger.Logger
	mu     sync.Mutex
}

func NewLinkService(repo *database.LinkRepository, cache *cache.RedisClient, logger *logger.Logger) *LinkService {
	return &LinkService{
		repo:   repo,
		cache:  cache,
		Logger: logger,
		mu:     sync.Mutex{},
	}
}

func (s *LinkService) generateUniqueShortLink(ctx context.Context, tx *sql.Tx) (string, error) {
	const maxAttempts = 100
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < maxAttempts; i++ {
		shortLink := generationShortLink()

		v, err := s.repo.ExistsShortLink(ctx, tx, shortLink)
		if err != nil {
			return "", fmt.Errorf("Error checking the short url: %w", err)
		} else if v == nil {
			return shortLink, nil
		}
	}

	return "", ErrTooManyAttempts
}

func (s *LinkService) Create(ctx context.Context, originalURL string) (*model.LinkStatsDTO, error) {
	if err := ValidLink(originalURL); err != nil {
		s.Logger.Error(err.Error(), logger.String("originalURL", originalURL))
		return nil, err
	}

	tx, err := s.repo.BeginTx(ctx)

	if err != nil {
		s.Logger.Error("Failed to begin transaction",
			logger.String("originalURL", originalURL),
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.Logger.Error("Failed to rollback transaction",
					logger.String("originalURL", originalURL),
					logger.ErrorField(rbErr))
			}
		}
	}()

	link, err := s.repo.ExistsOriginalLink(ctx, tx, originalURL)

	if err != nil {
		s.Logger.Error(err.Error(), logger.String("originalURL", originalURL))
		return nil, err

	} else if link == nil {
		link, err = s.repo.CreateOriginalLink(ctx, tx, originalURL)

		if err != nil {
			s.Logger.Error(err.Error(), logger.String("originalURL", originalURL))
			return nil, err
		}
	}

	shortURL, err := s.generateUniqueShortLink(ctx, tx)

	if err != nil {
		s.Logger.Error(err.Error(), logger.String("originalURL", originalURL), logger.String("shortURL", shortURL))
		return nil, err
	}

	shortLink, err := s.repo.CreateShortLink(ctx, tx, shortURL, link.Id)

	if err != nil {
		s.Logger.Error(err.Error(), logger.String("originalURL", originalURL), logger.String("shortURL", shortURL))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("Failed to commit transaction",
			logger.String("originalURL", originalURL),
			logger.String("shortURL", shortURL),
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if s.cache != nil {
		err := s.cache.Set(ctx, shortURL, originalURL, 1*time.Hour)

		if err != nil {
			s.Logger.Error("Failed to cache short link",
				logger.String("shortURL", shortURL),
				logger.String("originalURL", originalURL),
				logger.ErrorField(err),
			)
		} else {
			s.Logger.Info("Short link cached",
				logger.String("shortURL", shortURL),
				logger.String("originalURL", originalURL),
			)
		}
	}

	var linkDTO model.LinkStatsDTO

	linkDTO = model.LinkStatsDTO{
		URL:           link.URL,
		ShortURL:      shortLink.ShortURL,
		CreatedAt:     shortLink.CreatedAt,
		AccessedAt:    shortLink.AccessedAt,
		AccessedCount: shortLink.AccessedCount,
	}

	s.Logger.Info("Created Short Link", logger.String("shortURL", shortURL))

	return &linkDTO, nil
}

func (s *LinkService) GetOriginalLink(ctx context.Context, shortURL string) (*string, error) {
	if len(shortURL) == 0 {
		s.Logger.Error("Bad request: size shortURL eq 0")
		return nil, ErrLinkBadRequest
	}

	link, err := s.cache.Get(ctx, shortURL)

	err = s.cache.Set(ctx, shortURL, link, 1*time.Hour)

	if err != nil {
		s.Logger.Error("Failed to cache short link",
			logger.String("shortURL", shortURL),
			logger.String("originalURL", link),
			logger.ErrorField(err),
		)
	} else {
		s.Logger.Info("Short link cached",
			logger.String("shortURL", shortURL),
			logger.String("originalURL", link),
		)
	}

	if err == nil {
		tx, err := s.repo.BeginTx(ctx)

		if err != nil {
			s.Logger.Error("Failed to begin transaction",
				logger.String("shortURL", shortURL),
				logger.ErrorField(err))
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					s.Logger.Error("Failed to rollback transaction",
						logger.String("shortURL", shortURL),
						logger.ErrorField(rbErr))
				}
			}
		}()

		err = s.repo.AccessedCountIncrement(ctx, tx, shortURL)
		if err != nil {
			s.Logger.Error("Error during data update",
				logger.String("shortURL", shortURL),
				logger.ErrorField(err))
			return nil, fmt.Errorf("error updating data: %w", err)
		} else {
			s.Logger.Info("successful data update")
		}

		if err := tx.Commit(); err != nil {
			s.Logger.Error("Failed to commit transaction",
				logger.String("originalURL", link),
				logger.String("shortURL", shortURL),
				logger.ErrorField(err))
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return &link, nil
	}

	tx, err := s.repo.BeginTx(ctx)

	if err != nil {
		s.Logger.Error("Failed to begin transaction",
			logger.String("shortURL", shortURL),
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.Logger.Error("Failed to rollback transaction",
					logger.String("shortURL", shortURL),
					logger.ErrorField(rbErr))
			}
		}
	}()

	shortLink, err := s.repo.ExistsShortLink(ctx, tx, shortURL)

	if err != nil {
		s.Logger.Error(err.Error(), logger.String("shortURL", shortURL))
		return nil, err

	} else if shortLink == nil {
		s.Logger.Error("Error: shortLink not found", logger.String("shortURL", shortURL))
		return nil, fmt.Errorf("Error: shortLink not found")
	}

	link, err = s.repo.GetOriginalLink(ctx, tx, shortURL)

	if err != nil {
		s.Logger.Error(err.Error(), logger.String("shortURL", shortURL))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("Failed to commit transaction",
			logger.String("originalURL", link),
			logger.String("shortURL", shortURL),
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.Logger.Info("the original link was obtained from a short", logger.String("shortURL", shortURL), logger.String("originalURL", link))

	return &link, nil
}

func (s *LinkService) GetAllShortLink(ctx context.Context) ([]model.LinkStatsDTO, error) {
	tx, err := s.repo.BeginTx(ctx)

	if err != nil {
		s.Logger.Error("Failed to begin transaction",
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.Logger.Error("Failed to rollback transaction",
					logger.ErrorField(rbErr))
			}
		}
	}()

	links, err := s.repo.GetAllShortLink(ctx, tx)

	if err != nil {
		s.Logger.Error("Error when receiving all records " + err.Error())

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("Failed to commit transaction",
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.Logger.Info("All records have been received")

	return links, nil
}
