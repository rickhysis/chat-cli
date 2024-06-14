package models

// Message represents a chat message
type Message struct {
	ID        int64  `json:"id"`
	RoomID    int64  `json:"room_id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
