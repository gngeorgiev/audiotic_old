package models

import "time"

type Track struct {
	ID         string    `json:"id"`
	Thumbnail  string    `json:"thumbnail"`
	Title      string    `json:"title"`
	Provider   string    `json:"provider"`
	StreamUrl  string    `json:"streamUrl"`
	Next       string    `json:"next"`
	Previous   string    `json:"previous"`
	LastPlayed time.Time `json:"lastPlayed" storm:"index"`
}
