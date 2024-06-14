package models

// Room represents a chat room
type Room struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
