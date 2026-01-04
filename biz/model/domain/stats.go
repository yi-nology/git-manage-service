package domain

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
