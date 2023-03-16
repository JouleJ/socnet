package internal

import (
	"fmt"
	"github.com/JouleJ/socnet/core"
	"golang.org/x/net/html"
	"io"
	"log"
	"strings"
)

func WriteErrorString(w io.Writer, s string) {
	io.WriteString(w, `<h1 class="error">`)
	io.WriteString(w, s)
	io.WriteString(w, `</h1>`)
}

func WriteMessageString(w io.Writer, s string) {
	io.WriteString(w, `<h1 class="message">`)
	io.WriteString(w, s)
	io.WriteString(w, `</h1>`)
}

func WriteHeaderFooterContent(w io.Writer) {
	io.WriteString(w, `<nav>`)
	io.WriteString(w, `<a href="/newsfeed"> News Feed </a>`)
	io.WriteString(w, `<a href="/homepage"> Home Page </a>`)
	io.WriteString(w, `<a href="/login"> Login </a>`)
	io.WriteString(w, `<a href="/signup"> Sign Up </a>`)
	io.WriteString(w, `</nav>`)
}

func WriteHeader(w io.Writer) {
	io.WriteString(w, `<header>`)
	WriteHeaderFooterContent(w)
	io.WriteString(w, `</header>`)
}

func WriteFooter(w io.Writer) {
	io.WriteString(w, `<footer>`)
	WriteHeaderFooterContent(w)
	io.WriteString(w, `</footer>`)
}

func BeginHtml(w io.Writer) {
	io.WriteString(w, `<!DOCTYPE HTML>`)
	io.WriteString(w, `<html>`)
	io.WriteString(w, `    <head>`)
	io.WriteString(w, `        <link rel="stylesheet" type="text/css" href="style.css">`)
	io.WriteString(w, `    </head>`)
	io.WriteString(w, `    <body>`)
	WriteHeader(w)
}

func EndHtml(w io.Writer) {
	WriteFooter(w)
	io.WriteString(w, `    </body>`)
	io.WriteString(w, `</html>`)
}

func RenderUser(u *core.User, db core.Database) (string, error) {
	builder := &strings.Builder{}

	builder.WriteString(`<table>`)

	builder.WriteString(`<tr>`)
	builder.WriteString(`<td class="rowname">Login</td>`)
	fmt.Fprintf(builder, `<td>%v</td>`, html.EscapeString(u.Login))
	builder.WriteString(`</tr>`)

	builder.WriteString(`<tr>`)
	builder.WriteString(`<td class="rowname">Bio</td>`)
	fmt.Fprintf(builder, `<td>%v</td>`, html.EscapeString(string(u.Bio)))
	builder.WriteString(`</tr>`)

	ps, err := db.GetPostsByUser(u)
	if err != nil || ps == nil {
		return "", err
	}

	for _, p := range ps {
		builder.WriteString(`<tr>`)
		fmt.Fprintf(builder, `<td class="rowname"><a href="/post?id=%v">Post link</a></td>`, p.Id)
		fmt.Fprintf(builder, `<td class="post"><code><pre>%v</pre></code></td>`, html.EscapeString(string(p.Content)))
		builder.WriteString(`</tr>`)
	}

	builder.WriteString(`</table>`)

	return builder.String(), nil
}

func RenderUserByLogin(login string, db core.Database) (string, error) {
	u, err := db.FindUser(login)
	if err != nil {
		return "", fmt.Errorf("Failed to find user: login=%v, err=%v\n", login, err)
	}

	html, err := RenderUser(u, db)
	return html, err
}

func RenderUserById(id int, db core.Database) (string, error) {
	u, err := db.LoadUser(id)
	if err != nil {
		return "", fmt.Errorf("Failed to find user: id=%v, err=%v\n", id, err)
	}

	html, err := RenderUser(u, db)
	return html, err
}

func RenderPost(p *core.Post, db core.Database) (string, error) {
	builder := &strings.Builder{}

	builder.WriteString(`<table>`)

	builder.WriteString(`<tr>`)
	builder.WriteString(`<td class="rowname">Author</td>`)
	fmt.Fprintf(builder, `<td><a href="/user?id=%v">%v</a></td>`, p.Author.Id, html.EscapeString(p.Author.Login))
	builder.WriteString(`</tr>`)

	builder.WriteString(`<tr>`)
	builder.WriteString(`<td class="rowname">Post link</td>`)
	fmt.Fprintf(builder, `<td><a href="/post?id=%v">Post %v</a></td>`, p.Id, p.Id)
	builder.WriteString(`</tr>`)

	builder.WriteString(`<tr>`)
	builder.WriteString(`<td class="rowname">Content</td>`)
	fmt.Fprintf(builder, `<td class="post"><code><pre>%v</pre></code></td>`, html.EscapeString(string(p.Content)))
	builder.WriteString(`</tr>`)

	cs, err := db.GetCommentsByPost(p)
	if err != nil {
		log.Printf("Failed to get comments in RenderPost: %v\n", err)
	}

	for _, c := range cs {
		builder.WriteString(`<tr>`)
		fmt.Fprintf(builder, `<td class="rowname">Comment by <a href="/user?id=%v">%v</a></td>`, c.Author.Id, html.EscapeString(c.Author.Login))
		fmt.Fprintf(builder, `<td class="post"><code><pre>%v</pre></code></td`, html.EscapeString(string(c.Content)))
		builder.WriteString(`</tr>`)
	}

	builder.WriteString(`</table>`)

	return builder.String(), nil
}

func RenderPostById(id int, db core.Database) (string, error) {
	p, err := db.LoadPost(id)
	if err != nil {
		return "", fmt.Errorf("Failed to load post: id=%v, err=%v\n", id, err)
	}

	html, err := RenderPost(p, db)
	return html, err
}
