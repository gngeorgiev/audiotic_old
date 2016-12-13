package models

type Track struct {
	Id        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Provider  string `json:"provider"`
	StreamUrl string `json:"streamUrl"`
	Next      string `json:"next"`
	Previous  string `json:"previous"`
}
