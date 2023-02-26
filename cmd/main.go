package main

import (
	"github.com/go-chi/chi/v5"
    "fmt"
    "github.com/JouleJ/socnet/core"
    "github.com/JouleJ/socnet/internal"
    "io"
    "log"
    "net/http"
)

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

    r := chi.NewRouter()

    r.Get("/signup", func (w http.ResponseWriter, r *http.Request) {
        log.Printf("/signup\n")
        w.Write(signupHtml.Content())
    })

    r.Get("/login", func (w http.ResponseWriter, r *http.Request) {
        log.Printf("/login\n")
        w.Write(loginHtml.Content())
    })

    r.Get("/homepage", func (w http.ResponseWriter, r *http.Request) {
        var login string
        tokenCookie, err := r.Cookie("socnet_token")
        if err == nil && tokenCookie != nil {
            login, err = internal.VerifyToken(tokenCookie.Value)
        }

        log.Printf("/homepage login=%v\n", login)

        if len(login) == 0 || err != nil {
            log.Printf("Failed to verify token due to %v", err)
            io.WriteString(w, "<h1>You are not logged in</h1>")
            io.WriteString(w, `</p>Please visit <a href="/signup">sign up</a> or <a href="/login">log in</a> page</p>`) 
            return
        }

        fmt.Fprintf(w, "You are logged in as %v\n", login)
    })

    r.Post("/do_signup", func (w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        login := r.Form.Get("login")
        password := []byte(r.Form.Get("password"))
        h := internal.GetHash(password)
        bio := []byte(r.Form.Get("bio"))

        log.Printf("/do_signup: %v %v %v\n", login, password, bio)

        u, err := db.VerifyUser(login, h)
        if u != nil && err == nil {
            log.Printf("Failed to create user: user already exists\n")
            io.WriteString(w, "<h1>User already exists</h1>")
            return
        }

        u = &core.User{Login: login, PasswordHash: h, Bio: bio}
        err = db.CreateUser(u)
        if err != nil {
            log.Printf("Failed to create user: %v\n", err)
            io.WriteString(w, "<h1>User cannot be created</h1>")
            return
        }

        http.Redirect(w, r, "/login", http.StatusSeeOther)
    })

    r.Post("/do_login", func (w http.ResponseWriter, r *http.Request) {
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
            io.WriteString(w, "<h1>Cannot log in</h1>")
            return
        }

        http.SetCookie(w, &http.Cookie{Name: "socnet_token", Value: internal.MakeToken(u.Login)})
        http.Redirect(w, r, "/homepage", http.StatusSeeOther)
    })

    http.ListenAndServe(":80", r)
}
