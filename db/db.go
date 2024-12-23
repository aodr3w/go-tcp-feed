package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var conn *sql.DB

func execSQL(cmd string) error {
	_, err := conn.Exec(cmd)
	return err
}

func InitDB() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("PG_PASS")
	db := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", user, password, db)
	log.Println("postgres DSN: ", connStr)
	conn, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	createUserSQL := `CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY, 
	name VARCHAR(100) NOT NULL UNIQUE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	err = execSQL(createUserSQL)
	if err != nil {
		return err
	}

	createMessagesSQL := `CREATE TABLE IF NOT EXISTS messages (
	id SERIAL PRIMARY KEY,
	user_id INT NOT NULL REFERENCES users(id),
	text TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	err = execSQL(createMessagesSQL)
	if err != nil {
		return err
	}

	log.Println("PostgreSQL connected , tables created")
	return nil
}

type Dao struct {
	*sql.DB
}

func NewDAO() Dao {
	return Dao{
		conn,
	}
}

func (dao Dao) GetUserByName(name string) (*User, error) {
	var user User
	query := `SELECT id, name FROM users WHERE name = $1`
	err := dao.QueryRow(query, name).Scan(&user.ID, &user.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", name)
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return &user, nil
}

func (dao Dao) GetUserMessages(name string) ([]Message, error) {
	//TODO add pagination to this, offsets etc
	query := `
	SELECT m.id, u.name, m.text, m.created_at
	FROM messages m
	JOIN users u on m.user_id = u.id
	WHERE u.name = $1
	ORDER BY m.created_at ASC
	`
	rows, err := dao.Query(query, name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages for user %s: %w", name, err)
	}
	defer rows.Close()
	var messages []Message

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Name, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning message row: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}
	return messages, nil
}

func (dao Dao) InsertUserMessage(userId int, message string) error {
	query := `
	INSERT INTO messages (user_id, text, created_at)
	VALUES ($1, $2, NOW())
	`
	_, err := dao.Exec(query, userId, message)

	if err != nil {
		return fmt.Errorf("error inserting message for user ID %d: %w", userId, err)
	}
	return nil
}
