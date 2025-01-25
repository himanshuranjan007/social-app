package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("hr01")

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

var db *sql.DB

func initDB() {
	var err error
	connStr := os.Getenv("DB_CONN")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	log.Println("Database connection established")
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Error hashing password"}`, http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", creds.Email, string(hashedPassword))
	if err != nil {
		log.Printf("Error saving user to database: %v", err)
		http.Error(w, `{"error": "Error saving user to database"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("User created: %s", creds.Email)
	w.WriteHeader(http.StatusCreated)
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE email=$1", creds.Email).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		} else {
			http.Error(w, `{"error": "Error querying database"}`, http.StatusInternalServerError)
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Email: creds.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, `{"error": "Error generating token"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	log.Printf("User logged in: %s", creds.Email)
	w.WriteHeader(http.StatusOK)
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
} // fuck cors

type Post struct {
	ID      int       `json:"id"`
	Author  string    `json:"author"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
}

func createPostsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		author VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		created TIMESTAMP NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create posts table: %v", err)
	}
	fmt.Println("Posts table created or already exists.")
	return nil
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO posts (author, content, created) VALUES ($1, $2, $3)", post.Author, post.Content, time.Now())
	if err != nil {
		log.Printf("Error saving post to database: %v", err)
		http.Error(w, `{"error": "Error saving post to database"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Post created by: %s", post.Author)
	w.WriteHeader(http.StatusCreated)
}

// need more features , user sessions and user id logging with posts and need to create commnet section too
// dont know how to ship comment section neither in ui nor in backend
// high hopes..........
// finally ahhhhh....

func GetPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, author, content, created FROM posts ORDER BY created DESC")
	if err != nil {
		log.Printf("Error querying posts from database: %v", err)
		http.Error(w, `{"error": "Error querying posts from database"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Author, &post.Content, &post.Created)
		if err != nil {
			log.Printf("Error scanning post: %v", err)
			http.Error(w, `{"error": "Error scanning post"}`, http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func main() {
	if err := godotenv.Load(); err != nil { // loading env using godotenv.Load()
		log.Fatal("Error loading env file")
	} else {
		fmt.Println("env loaded successfooly")
	}
	initDB()

	defer db.Close()

	err := createPostsTable(db) //creating a posts table if does not exist in db already
	if err != nil {
		log.Fatalf("Error creating posts table: %v", err)
	}

	http.Handle("/signup", enableCors(http.HandlerFunc(SignUp)))
	http.Handle("/signin", enableCors(http.HandlerFunc(SignIn)))
	http.Handle("/createpost", enableCors(http.HandlerFunc(CreatePost)))
	http.Handle("/getposts", enableCors(http.HandlerFunc(GetPosts)))
// so shit for cors 


	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
