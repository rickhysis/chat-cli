package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// Client represents a chat client
type Client struct {
	conn *websocket.Conn
}

// NewClient creates a new chat client
func NewClient() *Client {
	return &Client{}
}

// Connect connects the client to the WebSocket server
func (c *Client) Connect(addr string) error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Message represents a message sent to the server
type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Send sends a message to the WebSocket server
func (c *Client) Send(message Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = c.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Printf("Send error: %v", err)
		return err
	}
	return nil
}

// Read reads a message from the WebSocket server
func (c *Client) Read() (string, error) {
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(message), nil
}

// Authenticate authenticates the user with the server
func (c *Client) Authenticate(username, password string) error {
	authMessage := Message{
		Type:    "auth",
		Payload: fmt.Sprintf("%s %s", username, password),
	}
	return c.Send(authMessage)
}

// JoinRoom joins a chat room
func (c *Client) JoinRoom(roomName string) error {
	joinMessage := Message{
		Type:    "join",
		Payload: roomName,
	}
	return c.Send(joinMessage)
}

// SendMessageToRoom sends a message to a chat room
func (c *Client) SendMessageToRoom(roomName, message string) error {
	roomMessage := Message{
		Type:    "room",
		Payload: fmt.Sprintf("%s %s", roomName, message),
	}
	return c.Send(roomMessage)
}

// SendDirectMessage sends a direct message to a user
func (c *Client) SendDirectMessage(username, message string) error {
	directMessage := Message{
		Type:    "dm",
		Payload: fmt.Sprintf("%s %s", username, message),
	}
	return c.Send(directMessage)
}

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "localhost:8080"
	}
	client := NewClient()
	err := client.Connect(serverURL)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter username:")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Println("Enter password:")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	err = client.Authenticate(username, password)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("Authenticated successfully")

	go func() {
		for {
			message, err := client.Read()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			fmt.Println("Received:", message)
		}
	}()

	for {
		fmt.Println("Enter command (/join [room], /room [room] [message], /dm [user] [message], /exit):")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		parts := strings.SplitN(command, " ", 3)

		switch parts[0] {
		case "/join":
			if len(parts) < 2 {
				fmt.Println("Usage: /join [room]")
				continue
			}
			err := client.JoinRoom(parts[1])
			if err != nil {
				log.Printf("Failed to join room: %v", err)
			}
		case "/room":
			if len(parts) < 3 {
				fmt.Println("Usage: /room [room] [message]")
				continue
			}
			err := client.SendMessageToRoom(parts[1], parts[2])
			if err != nil {
				log.Printf("Failed to send message to room: %v", err)
			}
		case "/dm":
			if len(parts) < 3 {
				fmt.Println("Usage: /dm [user] [message]")
				continue
			}
			err := client.SendDirectMessage(parts[1], parts[2])
			if err != nil {
				log.Printf("Failed to send direct message: %v", err)
			}
		case "/exit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
