package server

import (
	"log"
	"sync"

	"chat-app/models"
)

// ChatRoom represents a chat room where users can communicate
type ChatRoom struct {
	Name  string
	Users map[int64]*models.User // Change map key type to int64
	Mutex sync.Mutex
}

// CreateChatRoom initializes and returns a new chat room
func CreateChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		Name:  name,
		Users: make(map[int64]*models.User),
	}
}

// JoinChatRoom adds a user to the chat room
func (room *ChatRoom) JoinChatRoom(user *models.User) error {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if _, exists := room.Users[user.ID]; exists {
		return nil // User is already in the chat room
	}

	room.Users[user.ID] = user
	return nil
}

// LeaveChatRoom removes a user from the chat room
func (room *ChatRoom) LeaveChatRoom(user *models.User) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	delete(room.Users, user.ID)
}

// BroadcastMessage sends a message to all users in the chat room
func (room *ChatRoom) BroadcastMessage(sender *models.User, message string) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	for _, user := range room.Users {
		if user.Conn != nil {
			// Replace with actual message sending logic based on your application
			log.Printf("Sending message to user %s\n", user.Username)
		}
	}
}
