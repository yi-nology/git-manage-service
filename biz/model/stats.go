package model

import "time"

type Commit struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
}

type LineStat struct {
	Author    string
	Email     string
	Date      time.Time
	Extension string
}

type AuthorStat struct {
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	TotalLines int            `json:"total_lines"`
	FileTypes  map[string]int `json:"file_types"`
	// Date -> Lines mapping (e.g. "2023-01-01" -> 10)
	TimeTrend map[string]int `json:"time_trend"`
}

type StatsResponse struct {
	TotalLines int           `json:"total_lines"`
	Authors    []*AuthorStat `json:"authors"`
}
