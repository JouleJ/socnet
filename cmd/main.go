package main

import (
	"fmt"
	"github.com/JouleJ/socnet/core"
	"github.com/JouleJ/socnet/internal"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
)

func writeErrorString(w io.Writer, s string) {
    io.WriteString(w, `<h1 class="error">`)
    io.WriteString(w, s)
    io.WriteString(w, `</h1>`)
}

func beginHtml(w io.Writer) {
    io.WriteString(w, `<!DOCTYPE HTML>`)
    io.WriteString(w, `<html>`)
    io.WriteString(w, `    <head>`)
    io.WriteString(w, `        <link rel="stylesheet" type="text/css" href="style.css">`)
    io.WriteString(w, `    </head>`)
    io.WriteString(w, `    <body>`)
}

func endHtml(w io.Writer) {
    io.WriteString(w, `    </body>`)
    io.WriteString(w, `</html>`)
}

func main() {
	db := internal.NewDatabase()
	defer db.Close()

	rm := internal.NewResourceManager()
	signupHtml, err := core.GetFirstResourceByRegexp(rm, `.*signup\.html$`)
	if err != nil {
		log.Fatalf("Failed to find signup.html resource due to %v\n", err)
	}

	loginHtml, err := core.GetFirstResourceByRegexp(rm, `.*login\.html$`)
	if err != nil {
		log.Fatalf("Failed to find login.html resource due to %v\n", err)
	}

    styleCss, err := core.GetFirstResourceByRegexp(rm, `.*style\.css$`)
    if err != nil {
        log.Fatalf("Failed to find style.css resource due to %v\n", err)
    }

	r := chi.NewRouter()

    r.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("/style.css")
        w.Header().Set("Content-Type", "text/css")
        w.Write(styleCss.Content())
    }) 

	r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("/signup\n")
		w.Write(signupHtml.Content())
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("/login\n")
		w.Write(loginHtml.Content())
	})

	r.Get("/homepage", func(w http.ResponseWriter, r *http.Request) {
        beginHtml(w)
        defer endHtml(w)

		var login string
		tokenCookie, err := r.Cookie("socnet_token")
		if err == nil && tokenCookie != nil {
			login, err = internal.VerifyToken(tokenCookie.Value)
		}

		log.Printf("/homepage login=%v\n", login)

		if len(login) == 0 || err != nil {
			log.Printf("Failed to verify token due to %v", err)
			writeErrorString(w, "You are not logged in")
			io.WriteString(w, `<p class="error">Please visit <a href="/signup">sign up</a> or <a href="/login">log in</a> page</p>`)
			return
		}

        html, err := internal.RenderUserByLogin(login, db)
		fmt.Fprintf(w, html)
        if err != nil {
            log.Printf("Failed to render user %v: %v\n", login, err)
            writeErrorString(w, "Cannot render user")
        }
	})

	r.Post("/do_signup", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		login := r.Form.Get("login")
		password := []byte(r.Form.Get("password"))
		h := internal.GetHash(password)
		bio := []byte(r.Form.Get("bio"))

		log.Printf("/do_signup: %v %v %v\n", login, password, bio)

		u, err := db.FindUser(login)
		if u != nil && err == nil {
			log.Printf("Failed to create user: user already exists\n")

            beginHtml(w)
            defer endHtml(w)
            writeErrorString(w, "User already exists")
			return
		}

		u = &core.User{Login: login, PasswordHash: h, Bio: bio}
		err = db.CreateUser(u)
		if err != nil {
			log.Printf("Failed to create user: %v\n", err)

            beginHtml(w)
            defer endHtml(w)
            writeErrorString(w, "User cannot be created")
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	r.Post("/do_login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		login := r.Form.Get("login")
		password := []byte(r.Form.Get("password"))
		h := internal.GetHash(password)

		log.Printf("/do_login %v %v\n", login, password)

		u, err := db.VerifyUser(login, h)
		if u != nil && err == nil {
			log.Printf("Login and password match, user=%v\n", *u)
		} else {
			log.Printf("Login and password do not match, err=%v\n", err)

            beginHtml(w)
            defer endHtml(w)
            writeErrorString(w, "Cannot log in")
			return
		}

		http.SetCookie(w, &http.Cookie{Name: "socnet_token", Value: internal.MakeToken(u.Login)})
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
	})

	http.ListenAndServe(":80", r)
}
