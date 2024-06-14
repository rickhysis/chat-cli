package main

import (
	"chat-app/server"
	"chat-app/storage"
	"log"
	"os"
)

func main() {
	dbURL := os.Getenv("DB_PATH")
	if dbURL == "" {
		dbURL = "data/chat.db"
	}

	// Initialize SQLite database
	err := storage.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer storage.CloseDB()

	// Create some initial chat rooms (for demonstration)
	room1 := server.CreateChatRoom("General")
	room2 := server.CreateChatRoom("Random")

	// Example of creating users (for demonstration)
	user1ID, err := storage.CreateUser("user1", "password1")
	if err != nil {
		log.Fatalf("Error creating user1: %v", err)
	}
	user2ID, err := storage.CreateUser("user2", "password2")
	if err != nil {
		log.Fatalf("Error creating user2: %v", err)
	}

	// Example of adding users to chat rooms (for demonstration)
	user1, err := storage.GetUserByID(user1ID)
	if err != nil {
		log.Fatalf("Error getting user1: %v", err)
	}
	user2, err := storage.GetUserByID(user2ID)
	if err != nil {
		log.Fatalf("Error getting user2: %v", err)
	}
	room1.JoinChatRoom(user1)
	room2.JoinChatRoom(user2)

	// Start WebSocket server
	serverAddr := ":8080"
	server.StartServer(serverAddr)
}
