package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
    // Initialize the database connection
    initDB()

    r := mux.NewRouter()

    r.HandleFunc("/register", Register).Methods("POST")
    r.HandleFunc("/login", Login).Methods("POST")

    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
    var err error
    db, err = sql.Open("postgres", "user=postgres password=ghosh dbname=firstdb sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)

    fmt.Fprintln(w, "Register user successfully")
}

func Login(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var storedPassword string
    err = db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedPassword)
    if err != nil {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
        http.Error(w, "Invalid password", http.StatusUnauthorized)
        return
    }

    fmt.Fprintln(w, "Authentication successful")
}
