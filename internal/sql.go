package internal

import (
	"database/sql"
	"fmt"
	"github.com/JouleJ/socnet/core"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
)

type database struct {
	impl *sql.DB
}

func (db *database) CreateUser(u *core.User) error {
	query := `
INSERT INTO users (id, login, password_hash, bio)
SELECT COUNT(*) + 1, ?, ?, ? FROM users;
`

	result, err := db.impl.Exec(
		query,
		u.Login,
		u.PasswordHash,
		u.Bio)

	if err != nil {
		return err
	}
	lastInsertId, err := result.LastInsertId()
	u.Id = int(lastInsertId)

	return err
}

func (db *database) CreatePost(p *core.Post) error {
	return fmt.Errorf("Not yet implemented")
}

func (db *database) CreateComment(c *core.Comment) error {
	return fmt.Errorf("Not yet implemented")
}

func (db *database) CreateLike(l *core.Like) error {
	return fmt.Errorf("Not yet implemented")
}

func (db *database) LoadUser(id int) (*core.User, error) {
	rows, err := db.impl.Query(
		"SELECT login, password_hash, bio FROM users WHERE id = ?;",
		id)

	if err != nil || rows == nil {
		return nil, err
	}
	defer rows.Close()

	u := &core.User{Id: id}
	if rows.Next() {
		rows.Scan(&u.Login, &u.PasswordHash, &u.Bio)
	} else {
		return nil, fmt.Errorf("Failed to scan rows\n")
	}

	return u, nil
}

func (db *database) LoadPost(id int) (*core.Post, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

func (db *database) LoadComment(id int) (*core.Comment, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

func (db *database) LoadLike(id int) (*core.Like, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

func (db *database) VerifyUser(login string, passwordHash uint64) (*core.User, error) {
	rows, err := db.impl.Query(
		"SELECT id, bio FROM users WHERE login = ? AND password_hash = ?;",
		login,
		passwordHash)

	if err != nil || rows == nil {
		return nil, err
	}
	defer rows.Close()

	u := &core.User{Login: login, PasswordHash: passwordHash}
	if rows.Next() {
		rows.Scan(&u.Id, &u.Bio)
	} else {
		return nil, fmt.Errorf("Failed to scan rows\n")
	}

	return u, nil
}

func (db *database) Close() {
	db.impl.Close()
}

func NewDatabase() core.Database {
	volumePath := os.Getenv("VOLUME_PATH")
	log.Printf("VOLUME_PATH=%v\n", volumePath)
	if volumePath == "" {
		log.Fatal("VOLUME_PATH is empty\n")
	}

	dbPath := filepath.Join(volumePath, "database.db")
	dbImpl, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Fatalf("Failed to create database due to %v\n", err)
	}

	err = dbImpl.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database due to %v\n", err)
	}

	return &database{impl: dbImpl}
}
