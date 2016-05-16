package repository

import "time"

type BotLog struct {
	Id              int
	Message         string
	CreatedDateTime time.Time
}
