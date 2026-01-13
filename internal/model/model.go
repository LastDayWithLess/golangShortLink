package model

import "time"

type Link struct {
	Id  int64  `json:"id" db:"id"`
	URL string `json:"url" db:"url"`
}

type ShortLink struct {
	Id            int64      `json:"id" db:"id"`
	IdURL         int64      `json:"id_url" db:"id_url"`
	ShortURL      string     `json:"short_url" db:"short_url"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	AccessedAt    *time.Time `json:"accessed_at,omitempty" db:"accessed_at"`
	AccessedCount int        `json:"accessed_count" db:"accessed_count"`
}

type LinkStatsDTO struct {
	URL           string     `json:"url"`
	ShortURL      string     `json:"short_url"`
	CreatedAt     time.Time  `json:"create_at"`
	AccessedAt    *time.Time `json:"accessed_at,omitempty"`
	AccessedCount int        `json:"accessed_count"`
}

type ErrorDTO struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type LinkDTO struct {
	URL string `json:"url"`
}
