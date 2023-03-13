package core

type User struct {
	Id int

	Login        string
	PasswordHash uint64

	Bio []byte
}

type Post struct {
	Id int

	Author  *User
	Content []byte
}

type Comment struct {
	Id int

	Author        *User
	CommentedPost *Post
	Content       []byte
}

type Like struct {
	Id int

	Author       *User
	LikedPost    *Post
	LikedComment *Comment
}

type Database interface {
	CreateUser(u *User) error
	CreatePost(p *Post) error
	CreateComment(c *Comment) error
	CreateLike(l *Like) error

	LoadUser(id int) (*User, error)
	LoadPost(id int) (*Post, error)
	LoadComment(id int) (*Comment, error)
	LoadLike(id int) (*Like, error)

	VerifyUser(login string, passwordHash uint64) (*User, error)
    FindUser(login string) (*User, error)

    GetPostsByUser(*User) ([]Post, error)
    GetNewestPosts(count int) ([]Post, error)
    GetCommentsByPost(*Post) ([]Comment, error)

	Close()
}
