package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//DB is the shared database connection instance
//Exposed at package level for dao functions to use

var DB *sql.DB

func execSQL(cmd string) error {
	_, err := DB.Exec(cmd)
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
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
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
		name VARCHAR(100) NOT NULL,
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
