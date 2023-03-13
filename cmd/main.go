package main

import (
	"fmt"
	"github.com/JouleJ/socnet/core"
	"github.com/JouleJ/socnet/internal"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
    "io"
    "strconv"
)

const (
    newsFeedPostCount = 1000
)

func main() {
	rm := internal.NewResourceManager()
	signupHtml, err := core.GetFirstResourceByRegexp(rm, `.*signup\.html$`)
	if err != nil {
		log.Fatalf("Failed to find signup.html resource due to %v\n", err)
	}

	loginHtml, err := core.GetFirstResourceByRegexp(rm, `.*login\.html$`)
	if err != nil {
		log.Fatalf("Failed to find login.html resource due to %v\n", err)
	}
    
    mkPostHtml, err := core.GetFirstResourceByRegexp(rm, `.*mkpost\.html$`)
    if err != nil {
        log.Fatalf("Failed to find mkpost.html resource due to %v\n", err)
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
        db := internal.NewDatabase()
        defer db.Close()

        internal.BeginHtml(w)
        defer internal.EndHtml(w)

		var login string
		tokenCookie, err := r.Cookie("socnet_token")
		if err == nil && tokenCookie != nil {
			login, err = internal.VerifyToken(tokenCookie.Value)
		}

		log.Printf("/homepage login=%v\n", login)

		if len(login) == 0 || err != nil {
			log.Printf("Failed to verify token due to %v", err)
			internal.WriteErrorString(w, "You are not logged in")
			io.WriteString(w, `<p class="error">Please visit <a href="/signup">sign up</a> or <a href="/login">log in</a> page</p>`)
			return
		}

        html, err := internal.RenderUserByLogin(login, db)
        if err != nil {
            log.Printf("Failed to render user %v: %v\n", login, err)
            internal.WriteErrorString(w, "Cannot render user")
            return
        }

		fmt.Fprintf(w, html)
        w.Write(mkPostHtml.Content())
	})

    r.Get("/newsfeed", func(w http.ResponseWriter, r *http.Request) {
        db := internal.NewDatabase()
        defer db.Close()

        internal.BeginHtml(w)
        defer internal.EndHtml(w)

        log.Printf("/newsfeed postCount=%v", newsFeedPostCount)

        ps, err := db.GetNewestPosts(newsFeedPostCount)
        if err != nil {
            log.Printf("Failed to load news feed: %v\n", err)

            internal.WriteErrorString(w, "Cannot load newsfeed")
            return
        }

        for _, p := range ps {
            html, err := internal.RenderPost(&p)
            if err != nil {
                log.Printf("Failed to render post: id=%v, err=%v\n", p.Id, err)
            }

            fmt.Fprintf(w, html)
        }
    })

    r.Get("/post", func(w http.ResponseWriter, r *http.Request) {
        db := internal.NewDatabase()
        defer db.Close()

        internal.BeginHtml(w)
        defer internal.EndHtml(w)

        id, err := strconv.Atoi(r.URL.Query().Get("id"))
        if err != nil {
            log.Printf("Invalid post id: %v\n", err)
            internal.WriteErrorString(w, "Cannot show post with such id")
            return
        }

        log.Printf("/post id=%v\n", id)
        html, err := internal.RenderPostById(id, db)
        if err != nil {
            log.Printf("Failed to render post %v: %v\n", id, err)
            internal.WriteErrorString(w, "Cannot render post")
            return
        }

		fmt.Fprintf(w, html)
    })

	r.Post("/do_signup", func(w http.ResponseWriter, r *http.Request) {
        db := internal.NewDatabase()
        defer db.Close()

		r.ParseForm()
		login := r.Form.Get("login")
		password := []byte(r.Form.Get("password"))
		h := internal.GetHash(password)
		bio := []byte(r.Form.Get("bio"))

		log.Printf("/do_signup: %v %v %v\n", login, password, bio)

		u, err := db.FindUser(login)
		if u != nil && err == nil {
			log.Printf("Failed to create user: user already exists\n")

            internal.BeginHtml(w)
            defer internal.EndHtml(w)
            internal.WriteErrorString(w, "User already exists")
			return
		}

		u = &core.User{Login: login, PasswordHash: h, Bio: bio}
		err = db.CreateUser(u)
		if err != nil {
			log.Printf("Failed to create user: %v\n", err)

            internal.BeginHtml(w)
            defer internal.EndHtml(w)
            internal.WriteErrorString(w, "User cannot be created")
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	r.Post("/do_login", func(w http.ResponseWriter, r *http.Request) {
        db := internal.NewDatabase()
        defer db.Close()

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

            internal.BeginHtml(w)
            defer internal.EndHtml(w)
            internal.WriteErrorString(w, "Cannot log in")
			return
		}

		http.SetCookie(w, &http.Cookie{Name: "socnet_token", Value: internal.MakeToken(u.Login)})
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
	})

    r.Post("/do_post", func(w http.ResponseWriter, r *http.Request) {
        db := internal.NewDatabase()
        defer db.Close()

        r.ParseForm()
        postContent := []byte(r.Form.Get("postContent"))

        if len(postContent) == 0 {
            internal.BeginHtml(w)
            defer internal.EndHtml(w)

			internal.WriteErrorString(w, "Empty posts are not allowed")
            return
        }

		var login string
		tokenCookie, err := r.Cookie("socnet_token")
		if err == nil && tokenCookie != nil {
			login, err = internal.VerifyToken(tokenCookie.Value)
		}

        if len(login) == 0 {
            internal.BeginHtml(w)
            defer internal.EndHtml(w)

			internal.WriteErrorString(w, "You are not logged in")
			io.WriteString(w, `<p class="error">Please visit <a href="/signup">sign up</a> or <a href="/login">log in</a> page</p>`)
			return
        }

        log.Printf("/do_post login=%v postContent=%v", login, postContent)

        u, err := db.FindUser(login)
        if err != nil {
            log.Printf("Failed to find user: %v\n", err)

            internal.BeginHtml(w)
            defer internal.EndHtml(w)

            internal.WriteErrorString(w, "You are logged in as non-existant user")
            return
        }

        p := &core.Post{Author: u, Content: postContent}
        err = db.CreatePost(p)
        if err != nil {
            log.Printf("Failed to create post: %v\n", err)

            internal.BeginHtml(w)
            defer internal.EndHtml(w)

            internal.WriteErrorString(w, "Failed to create post")
            return
        }

		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
    })

	http.ListenAndServe(":80", r)
}
