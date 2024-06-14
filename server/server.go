package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"chat-app/models"
	"chat-app/storage"

	"github.com/gorilla/websocket"
)

// WebSocketServer represents the WebSocket server
type WebSocketServer struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]*models.User
	mutex    sync.Mutex
	rooms    map[string]*ChatRoom
}

// Message represents a message sent to the server
type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]*models.User),
		rooms:   make(map[string]*ChatRoom),
	}
}

// StartServer starts the WebSocket server
func StartServer(addr string) error {
	server := NewWebSocketServer()
	http.HandleFunc("/ws", server.handleConnections)
	log.Println("Starting server on", addr)
	return http.ListenAndServe(addr, nil)
}

func (server *WebSocketServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			server.handleDisconnect(conn)
			break
		}

		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		log.Println("Payload", message.Payload)

		switch message.Type {
		case "auth":
			server.handleAuth(conn, message.Payload)
		case "join":
			server.handleJoin(conn, message.Payload)
		case "room":
			server.handleRoomMessage(conn, message.Payload)
		case "dm":
			server.handleDirectMessage(conn, message.Payload)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

func (server *WebSocketServer) handleAuth(conn *websocket.Conn, payload string) {
	parts := strings.SplitN(payload, " ", 2)
	if len(parts) != 2 {
		log.Printf("Invalid auth payload: %s", payload)
		return
	}
	username, password := parts[0], parts[1]

	user, err := storage.AuthenticateUser(username, password)
	if err != nil {
		log.Printf("Authentication failed for user %s: %v", username, err)
		return
	}

	// Set the WebSocket connection
	user.Conn = conn

	// Update clients map
	server.mutex.Lock()
	server.clients[conn] = user
	server.mutex.Unlock()

	// Log successful authentication
	log.Printf("User %s authenticated successfully", username)
}

func (server *WebSocketServer) handleJoin(conn *websocket.Conn, roomName string) {
	server.mutex.Lock()
	user, ok := server.clients[conn]
	server.mutex.Unlock()
	if !ok {
		log.Printf("User not authenticated")
		return
	}

	room, exists := server.rooms[roomName]
	if !exists {
		room = CreateChatRoom(roomName)
		server.rooms[roomName] = room
	}

	err := room.JoinChatRoom(user)
	if err != nil {
		log.Printf("Error adding user to room: %v", err)
		return
	}

	log.Printf("User %s joined room %s", user.Username, roomName)
}

func (server *WebSocketServer) handleRoomMessage(conn *websocket.Conn, payload string) {
	parts := strings.SplitN(payload, " ", 2)
	if len(parts) != 2 {
		log.Printf("Invalid room message payload: %s", payload)
		return
	}
	roomName, message := parts[0], parts[1]

	server.mutex.Lock()
	user, ok := server.clients[conn]
	server.mutex.Unlock()
	if !ok {
		log.Printf("User not authenticated")
		return
	}

	room, exists := server.rooms[roomName]
	if !exists {
		log.Printf("Room %s does not exist", roomName)
		return
	}

	room.BroadcastMessage(user, message)
}

func (server *WebSocketServer) handleDirectMessage(conn *websocket.Conn, payload string) {
	parts := strings.SplitN(payload, " ", 2)
	if len(parts) != 2 {
		log.Printf("Invalid direct message payload: %s", payload)
		return
	}
	username, message := parts[0], parts[1]

	server.mutex.Lock()
	sender, ok := server.clients[conn]
	server.mutex.Unlock()
	if !ok {
		log.Printf("User not authenticated")
		return
	}

	receiver, err := storage.GetUserByUsername(username)
	if err != nil {
		log.Printf("User %s not found: %v", username, err)
		return
	}

	// Ensure the receiver is also authenticated and has a valid connection
	server.mutex.Lock()
	if receiverUser, ok := server.clients[receiver.Conn]; ok && receiverUser.Conn != nil {
		server.sendMessageToUser(receiverUser, sender.Username, message)
	} else {
		log.Printf("User %s has no active connection", receiver.Username)
	}
	server.mutex.Unlock()
}

func (server *WebSocketServer) sendMessageToUser(user *models.User, sender, message string) {
	err := user.Conn.WriteJSON(Message{
		Type:    "dm",
		Payload: fmt.Sprintf("%s: %s", sender, message),
	})
	if err != nil {
		log.Printf("Error sending direct message to user %s: %v", user.Username, err)
	}
}

// handleDisconnect handles a client disconnecting
func (server *WebSocketServer) handleDisconnect(conn *websocket.Conn) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if user, ok := server.clients[conn]; ok {
		for _, room := range server.rooms {
			room.LeaveChatRoom(user)
		}
		delete(server.clients, conn)
	}
	log.Printf("User disconnected")
}
