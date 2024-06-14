package storage

import (
	"database/sql"
	"errors"
	"log"

	"chat-app/models"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB // DB is the global database connection pool

// InitDB initializes the SQLite database
func InitDB(dbURL string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbURL)
	if err != nil {
		return err
	}

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return err
	}

	log.Printf("Connected to database: %s\n", dbURL)
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Error closing database: %v\n", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}

// Function to create necessary tables if they don't exist
func createTables() error {
	// Example table creation for users
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Example table creation for chat rooms
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS chat_rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Example table creation for messages
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (room_id) REFERENCES chat_rooms(id),
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Additional tables can be created here as needed

	return nil
}

// CreateUser creates a new user
func CreateUser(username, password string) (int64, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

	// Insert user into database
	result, err := DB.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("User '%s' created with ID %d\n", username, userID)
	return userID, nil
}

func GetUserByID(id int64) (*models.User, error) {
	row := DB.QueryRow("SELECT id, username, password FROM users WHERE id = ?", id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	row := DB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func AuthenticateUser(username, password string) (*models.User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// CreateRoom creates a new chat room
func CreateRoom(name string) (int64, error) {
	// Insert room into database
	result, err := DB.Exec("INSERT INTO chat_rooms(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}

	roomID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Chat room '%s' created with ID %d\n", name, roomID)
	return roomID, nil
}

// CreateMessage creates a new message in a chat room
func CreateMessage(roomID, senderID int64, content string) (int64, error) {
	// Insert message into database
	result, err := DB.Exec("INSERT INTO messages(room_id, sender_id, content) VALUES(?, ?, ?)", roomID, senderID, content)
	if err != nil {
		return 0, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Message created with ID %d in room ID %d\n", messageID, roomID)
	return messageID, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
