package internal

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
)

// TODO: use cryptographic hash
func GetHash(password []byte) uint64 {
    h, magic1, magic2 := uint64(0), uint64(6863), uint64(7919)
    salt := []byte(os.Getenv("SALT"))

    log.Printf("SALT=%v\n", salt)
    if len(salt) == 0 {
        log.Fatalf("SALT is empty\n")
    }

    for _, b := range password {
        h += uint64(b)
        h *= magic1
    }

    for _, b := range salt {
        h += uint64(b)
        h *= magic1
    }

    for _, b := range password {
        h ^= uint64(b)
        h *= magic2
    }

    for _, b := range salt {
        h ^= uint64(b)
        h *= magic2
    }

    return h
}

// TODO: use JWT
func MakeToken(login string) string {
    loginBytes := []byte(login)
    return fmt.Sprintf("%v:%v", login, GetHash(loginBytes))
}

func VerifyToken(token string) (string, error) {
    i := strings.Index(token, ":")
    if i < 0 {
        return "", fmt.Errorf("Failed to find ':'")
    }

    login := token[:i]
    if len(login) == 0 {
        return "", fmt.Errorf("Login is empty")
    }

    h, err := strconv.ParseUint(token[i+1:], 10, 64) 
    if err != nil {
        return "", fmt.Errorf("Failed to parse hash: %v", err)
    }

    loginBytes := []byte(login)
    if GetHash(loginBytes) != h {
        return "", fmt.Errorf("Login and hash do not match")
    }

    return login, nil
}
