package data

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

var conn *sql.DB

type MessageOrder string

var Latest MessageOrder = "DESC"
var Oldest MessageOrder = "ASC"

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

func (dao Dao) CreateUser(name string) (*User, error) {
	//we may need some kind of db level locking here
	query := `
	INSERT INTO USERS (name, created_at)
	VALUES ($1, NOW())
	RETURNING id, name
	`
	var user User

	err := dao.QueryRow(query, name).Scan(&user.ID, &user.Name)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, &UserNameNotAvailable{name}
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return &user, nil
}

func (dao Dao) GetMessageCount() (int, error) {
	var count int
	query := `SELECT COUNT(1) FROM messages`
	err := dao.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error retrieving message count: %w", err)
	}
	return count, nil
}
func (dao Dao) GetUserByName(name string) (*User, error) {
	var user User
	query := `SELECT id, name FROM users WHERE name = $1`
	err := dao.QueryRow(query, name).Scan(&user.ID, &user.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &UserNotFoundError
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return &user, nil
}

func (dao Dao) GetReceivedMessages(userID int, size int, offset int, minTime time.Time) ([]MessagePayload, error) {
	//Get other messages other than current users messages
	query := `
	SELECT m.id, u.name, m.text, m.created_at
	FROM messages m
	JOIN users u on m.user_id = u.id
	WHERE u.id != $1
	AND m.created_at >= $2
	ORDER BY m.created_at ASC
	LIMIT $3 OFFSET $4
	`
	rows, err := dao.Query(query, userID, minTime, size, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages for user %d: %w", userID, err)
	}
	defer rows.Close()
	var messages []MessagePayload

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Name, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning message row: %w", err)
		}
		//get associated user name
		messages = append(messages, MessagePayload{
			Message: msg,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}
	return messages, nil
}

func (dao Dao) InsertUserMessage(userID int, message string, createdAt time.Time) error {
	log.Println("inserting with time: ", createdAt)
	query := `
	INSERT INTO messages (user_id, text, created_at)
	VALUES ($1, $2, $3)
	`
	_, err := dao.Exec(query, userID, message, createdAt)

	if err != nil {
		return fmt.Errorf("error inserting message for user ID %d: %w", userID, err)
	}
	return nil
}

func (dao Dao) GetMessages(size int, offset int, mo MessageOrder, maxTime time.Time) ([]Message, error) {

	query := fmt.Sprintf(
		`
		SELECT m.id, u.name, m.text, m.created_at FROM
		messages m JOIN users u on m.user_id = u.id
		WHERE m.created_at <= $3
		ORDER BY m.created_at %s LIMIT $1 OFFSET $2
		`, mo)
	rows, err := dao.Query(query, size, offset, maxTime)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages %w", err)
	}

	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Name, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("error retrieving message row %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}
	return messages, nil
}

func (dao Dao) GetMessageStream(size int, offset int, mo MessageOrder) ([]Message, error) {

	query := fmt.Sprintf(
		`
		SELECT m.id, u.name, m.text, m.created_at FROM
		messages m JOIN users u on m.user_id = u.id
		ORDER BY m.created_at %s LIMIT $1 OFFSET $2
		`, mo)
	rows, err := dao.Query(query, size, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages %w", err)
	}

	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Name, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("error retrieving message row %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over messages: %w", err)
	}
	return messages, nil
}
