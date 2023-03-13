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
    query := `
INSERT INTO posts (id, author, content) 
SELECT COUNT(*) + 1, ?, ? FROM posts;
`

    result, err := db.impl.Exec(
        query,
        p.Author.Id,
        p.Content)

    if err != nil {
        return err
    }

	lastInsertId, err := result.LastInsertId()
	p.Id = int(lastInsertId)
	return err
}

func (db *database) CreateComment(c *core.Comment) error {
    query := `
INSERT INTO comments (id, author, commented_post, content)
SELECT COUNT(*) + 1, ?, ?, ? FROM comments;
`

    result, err := db.impl.Exec(
        query,
        c.Author.Id,
        c.CommentedPost.Id,
        c.Content)

    if err != nil {
        return err
    }

    lastInsertId, err := result.LastInsertId()
    c.Id = int(lastInsertId)
    return err
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
    rows, err := db.impl.Query(
        "SELECT author, content FROM posts WHERE id = ?;",
        id)

    if err != nil || rows == nil {
        return nil, err
    }

    defer rows.Close()

    p := &core.Post{Id: id}
    if rows.Next() {
        var authorId int

        rows.Scan(&authorId, &p.Content)

        p.Author, err = db.LoadUser(authorId)
        if err != nil {
            return nil, err
        }
    } else {
        return nil, fmt.Errorf("Failed to scan rows\n")
    }

    return p, nil
}

func (db *database) GetPostsByUser(u *core.User) ([]core.Post, error) {
    rows, err := db.impl.Query(
        `SELECT id, content FROM posts WHERE author = ?;`,
        u.Id)

    if err != nil || rows == nil {
        return nil, fmt.Errorf("Failed to list posts of user %v:%v due to %v\n", u.Id, u.Login, err)
    }
    defer rows.Close()

    ps := []core.Post{}
    for rows.Next() {
        p := core.Post{Author: u}
        rows.Scan(&p.Id, &p.Content)

        ps = append(ps, p)
    }

    return ps, nil
}

func (db *database) GetCommentsByPost(p *core.Post) ([]core.Comment, error) {
    rows, err := db.impl.Query(
        `SELECT c.id, c.content, u.id, u.login, u.password_hash, u.bio
         FROM comments as c
         INNER JOIN users as u
         ON u.id == c.author
         WHERE c.commented_post = ?;`,
        p.Id)

    if err != nil || rows == nil {
        return nil, fmt.Errorf("Failed to get comments to post %v due to %v\n", p.Id, err)
    }
    defer rows.Close()

    cs := []core.Comment{}
    for rows.Next() {
        u := &core.User{}
        c := core.Comment{CommentedPost: p, Author: u}
        rows.Scan(
            &c.Id,
            &c.Content,
            &c.Author.Id,
            &c.Author.Login,
            &c.Author.PasswordHash,
            &c.Author.Bio)

        cs = append(cs, c)
    }

    return cs, nil
}

func (db *database) GetNewestPosts(count int) ([]core.Post, error) {
    rows, err := db.impl.Query(
        `SELECT p.id, p.content, u.id, u.login, u.password_hash, u.bio
         FROM posts as p
         INNER JOIN users as u
         ON u.id = p.author
         ORDER BY p.id DESC;`)

    if err != nil || rows == nil {
        return nil, fmt.Errorf("Failed to list %v newst posts due to %v\n", count, err)
    }
    defer rows.Close()

    ps := make([]core.Post, 0, count)
    for i := 0; i < count; i++ {
        if !rows.Next() {
            break
        }

        u := &core.User{}
        p := core.Post{Author: u}
        rows.Scan(
            &p.Id,
            &p.Content,
            &p.Author.Id,
            &p.Author.Login,
            &p.Author.PasswordHash,
            &p.Author.Bio)

        ps = append(ps, p)
    }

    return ps, nil
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

func (db *database) FindUser(login string) (*core.User, error) {
	rows, err := db.impl.Query(
		"SELECT id, bio, password_hash FROM users WHERE login = ?;",
		login)

	if err != nil || rows == nil {
		return nil, err
	}
	defer rows.Close()

	u := &core.User{Login: login}
	if rows.Next() {
		rows.Scan(&u.Id, &u.Bio, &u.PasswordHash)
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
