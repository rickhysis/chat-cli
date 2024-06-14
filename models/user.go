package models

import "github.com/gorilla/websocket"

// User represents a chat user
type User struct {
	ID       int64           `json:"id"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Conn     *websocket.Conn `json:"-"`
}
