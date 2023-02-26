package internal 

import (
    "fmt"
    "github.com/JouleJ/socnet/core"
    "golang.org/x/net/html"
    "strings"
)

func RenderUser(u *core.User, db core.Database) (string, error) {
    builder := &strings.Builder{}

    builder.WriteString("<table>")

    builder.WriteString("<tr>")
    builder.WriteString("<td>Login</td>")
    fmt.Fprintf(builder, "<td>%v</td>", html.EscapeString(u.Login))
    builder.WriteString("</tr>")

    builder.WriteString("<tr>")
    builder.WriteString("<td>Bio</td>")
    fmt.Fprintf(builder, "<td>%v</td>", html.EscapeString(string(u.Bio)))
    builder.WriteString("</tr>")

    ps, err := db.GetPostsByUser(u)
    if err != nil || ps == nil {
        return "", err
    }

    for _, p := range ps {
        builder.WriteString("<tr>")
        fmt.Fprintf(builder, "<td>Post %v</td>", p.Id)
        fmt.Fprintf(builder, "<td>%v</td>", html.EscapeString(string(p.Content)))
        builder.WriteString("</tr>")
    }

    builder.WriteString("</table>")

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
