package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	DbUser     = "myuser"
	DbPassword = "mypassword"
	DbName     = "mydatabase"
	DbHost     = "postgres"
	DbPort     = "5432"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ShiftReceive struct {
	Datum  string                   `json:"datum"`
	Day    string                   `json:"day"`
	Time   string                   `json:"time"`
	Trade  bool                     `json:"trade"`
	Uid    string                   `json:"uid"`
	Search []map[string]interface{} `json:"search"`
}

type App struct {
	DB *sql.DB
}

// Auth middleware
func (app *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sessionID")
		if err != nil || !app.isValidSession(cookie.Value) {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
			if err != nil {
				return
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *App) isValidSession(sessionID string) bool {
	var username string
	err := app.DB.QueryRow("SELECT username FROM sessions WHERE sessionID=$1 AND timeout > NOW()", sessionID).Scan(&username)
	return err == nil
}

// Login handler
func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.handleLoginPost(w, r)
	case http.MethodPut:
		app.handleLoginPut(w, r)
	case http.MethodGet:
		app.handleLoginGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (app *App) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid request payload\"}"))
		if err != nil {
			return
		}
		return
	}

	// Check user credentials against the database
	var storedPassword string
	err = app.DB.QueryRow("SELECT password FROM user_base WHERE username=$1", user.Username).Scan(&storedPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Invalid credentials\"}"))
		if err != nil {
			return
		}
		return
	}

	// Compare the stored hashed password with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Invalid credentials\"}"))
		if err != nil {
			return
		}
		return
	}

	// Generate a new session ID
	sessionID := generateSessionID()

	// Store the session in the database
	timeout := time.Now().Add(24 * time.Hour) // 24 hour session timeout
	_, err = app.DB.Exec("INSERT INTO sessions (sessionID, username, timeout) VALUES ($1, $2, $3)", sessionID, user.Username, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to create session\"}"))
		if err != nil {
			return
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	})
	_, err = w.Write([]byte("{\"message\": \"Login successful\"}"))
	if err != nil {
		return
	}
}

func (app *App) handleLoginPut(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid request payload\"}"))
		if err != nil {
			return
		}
		return
	}

	// Hash the user's password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to hash password\"}"))
		if err != nil {
			return
		}
		return
	}

	// Store user credentials in the database
	_, err = app.DB.Exec("INSERT INTO user_base (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to create user\"}"))
		if err != nil {
			return
		}
		return
	}

	// Generate a new session ID
	sessionID := generateSessionID()

	// Store the session in the database
	timeout := time.Now().Add(24 * time.Hour) // 24 hour session timeout
	_, err = app.DB.Exec("INSERT INTO sessions (sessionID, username, timeout) VALUES ($1, $2, $3)", sessionID, user.Username, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to create session\"}"))
		if err != nil {
			return
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	})
	_, err = w.Write([]byte("{\"message\": \"User created and login successful\"}"))
	if err != nil {
		return
	}
}

func (app *App) handleLoginGet(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		// No session cookie, respond with loggedIn: false
		response := map[string]bool{"loggedIn": false}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
		return
	}

	if app.isValidSession(cookie.Value) {
		// Valid session, respond with loggedIn: true
		response := map[string]bool{"loggedIn": true}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	} else {
		// Invalid session, respond with loggedIn: false
		response := map[string]bool{"loggedIn": false}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	}
}

// Logout handler
func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.handleLogoutPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (app *App) handleLogoutPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized\"}"))
		if err != nil {
			return
		}
		return
	}

	_, err = app.DB.Exec("DELETE FROM sessions WHERE sessionID=$1", cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to logout\"}"))
		if err != nil {
			return
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1, // Delete cookie
	})
	_, err = w.Write([]byte("{\"message\": \"Logout successful\"}"))
	if err != nil {
		return
	}
}

// Private handler
func (app *App) shiftHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.shiftHandlerGet(w, r)
	case http.MethodPost:
		app.shiftHandlerPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (app *App) shiftHandlerGet(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	var username string
	err = app.DB.QueryRow("SELECT username FROM sessions WHERE sessionID=$1", cookie.Value).Scan(&username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to get username from session\"}"))
		if err != nil {
			return
		}
		return
	}

	rows, err := app.DB.Query("SELECT shiftID, date, time, day, TRADE, search_early, search_evening, search_night FROM shifts WHERE username=$1", username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to fetch shifts\"}"))
		if err != nil {
			return
		}
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var shifts []map[string]interface{}
	for rows.Next() {
		var shiftID, date, timeV, day string
		var trade, searchEarly, searchEvening, searchNight bool
		err := rows.Scan(&shiftID, &date, &timeV, &day, &trade, &searchEarly, &searchEvening, &searchNight)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("{\"message\": \"Failed to scan shift\"}"))
			if err != nil {
				return
			}
			return
		}

		var earlyCount string
		var lateCount string
		var nightCount string

		err = app.DB.QueryRow("SELECT count(*) FROM shifts WHERE date=$1 AND trade=true AND time='früh'", date).Scan(&earlyCount)
		err = app.DB.QueryRow("SELECT count(*) FROM shifts WHERE date=$1 AND trade=true AND time='spät'", date).Scan(&lateCount)
		err = app.DB.QueryRow("SELECT count(*) FROM shifts WHERE date=$1 AND trade=true AND time='nacht'", date).Scan(&nightCount)

		shift := map[string]interface{}{
			"uid":   shiftID,
			"datum": date,
			"time":  timeV,
			"day":   day,
			"trade": trade,
			"search": []map[string]interface{}{
				{"selected": searchEarly, "name": "früh", "offers": earlyCount},
				{"selected": searchEvening, "name": "spät", "offers": lateCount},
				{"selected": searchNight, "name": "nacht", "offers": nightCount},
			},
		}
		shifts = append(shifts, shift)
	}

	if err = rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to iterate over shifts\"}"))
		if err != nil {
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(shifts)
	if err != nil {
		return
	}
}

func (app *App) shiftHandlerPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	var username string
	err = app.DB.QueryRow("SELECT username FROM sessions WHERE sessionID=$1", cookie.Value).Scan(&username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to get username from session\"}"))
		if err != nil {
			return
		}
		return
	}

	var shift ShiftReceive
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid request payload\"}"))
		if err != nil {
			return
		}
		return
	}

	shiftID := generateSessionID() // Use generateSessionID to create a unique ID for the shift

	_, err = app.DB.Exec("INSERT INTO shifts (shiftID, username, date, time, day, TRADE, search_early, search_evening, search_night) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		shiftID, username, shift.Datum, shift.Time, shift.Day, false, false, false, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to add shift\"}"))
		if err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("{\"message\": \"{\"message\": \"Shift added successfully\"}"))
	if err != nil {
		return
	}
}

// New handler for /shifts/{id}
func (app *App) shiftByIDHandler(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid shift ID\"}"))
		if err != nil {
			return
		}
		return
	}
	shiftID := pathSegments[2]

	switch r.Method {
	case http.MethodGet:
		app.shiftByIDGet(w, r, shiftID)
	case http.MethodPut:
		app.shiftByIDPut(w, r, shiftID)
	case http.MethodDelete:
		app.shiftByIDDelete(w, r, shiftID)
	case http.MethodPatch:
		app.shiftByIDPatch(w, r, shiftID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (app *App) shiftByIDGet(w http.ResponseWriter, r *http.Request, shiftID string) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	var shift ShiftReceive
	err = app.DB.QueryRow("SELECT date, time, day, TRADE, search_early, search_evening, search_night FROM shifts WHERE shiftID=$1", shiftID).Scan(&shift.Datum, &shift.Time, &shift.Day, &shift.Trade, &shift.Search[0], &shift.Search[1], &shift.Search[2])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("{\"message\": \"Shift not found\"}"))
			if err != nil {
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("{\"message\": \"Failed to get shift\"}"))
			if err != nil {
				return
			}
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(shift)
	if err != nil {
		return
	}
}

func (app *App) shiftByIDPut(w http.ResponseWriter, r *http.Request, shiftID string) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	var shift ShiftReceive
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid request payload\"}"))
		if err != nil {
			return
		}
		return
	}

	_, err = app.DB.Exec("UPDATE shifts SET date=$1, time=$2, day=$3, TRADE=$4, search_early=$5, search_evening=$6, search_night=$7 WHERE shiftID=$8",
		shift.Datum, shift.Time, shift.Day, shift.Trade,
		getSearchValue(shift.Search, "früh"), getSearchValue(shift.Search, "spät"), getSearchValue(shift.Search, "nacht"),
		shiftID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to update shift\"}"))
		if err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("{\"message\": \"Shift updated successfully\"}"))
	if err != nil {
		return
	}
}

func (app *App) shiftByIDDelete(w http.ResponseWriter, r *http.Request, shiftID string) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	_, err = app.DB.Exec("DELETE FROM shifts WHERE shiftID=$1", shiftID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to delete shift\"}"))
		if err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("{\"message\": \"Shift deleted successfully\"}"))
	if err != nil {
		return
	}
}

func (app *App) shiftByIDPatch(w http.ResponseWriter, r *http.Request, shiftID string) {
	cookie, err := r.Cookie("sessionID")
	if err != nil || !app.isValidSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("{\"message\": \"Unauthorized"))
		if err != nil {
			return
		}
		return
	}

	var shift ShiftReceive
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"message\": \"Invalid request payload\"}"))
		if err != nil {
			return
		}
		return
	}

	tx, err := app.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to begin transaction\"}"))
		if err != nil {
			return
		}
		return
	}

	// Update the current shift
	query := "UPDATE shifts SET date=$1, time=$2, day=$3, TRADE=$4, search_early=$5, search_evening=$6, search_night=$7 WHERE shiftID=$8"
	_, err = tx.Exec(query, shift.Datum, shift.Time, shift.Day, shift.Trade,
		getSearchValue(shift.Search, "früh"), getSearchValue(shift.Search, "spät"), getSearchValue(shift.Search, "nacht"),
		shiftID)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("{\"message\": \"Failed to update shift\"}"))
		if err != nil {
			return
		}
		return
	}

	if shift.Trade {
		// Check for a matching shift to trade with
		var matchingShiftID, matchingUsername string
		err = tx.QueryRow(`
			SELECT shiftID, username FROM shifts
			WHERE date=$1 AND TRADE=true
			AND ((search_early=true AND $2='früh') OR (search_evening=true AND $2='spät') OR (search_night=true AND $2='nacht'))
			AND shiftID <> $3
		`, shift.Datum, shift.Time, shiftID).Scan(&matchingShiftID, &matchingUsername)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // No matching shift found, not an error
		} else if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte("{\"message\": \"Failed to find matching shift\"}"))
			if err != nil {
				return
			}
			return
		} else {
			// Swap usernames
			var currentUsername string
			err = tx.QueryRow("SELECT username FROM shifts WHERE shiftID=$1", shiftID).Scan(&currentUsername)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("{\"message\": \"Failed to get current username\"}"))
				if err != nil {
					return
				}
				return
			}

			_, err = tx.Exec("UPDATE shifts SET username=$1, trade=false, search_early=false, search_evening=false, search_night=false WHERE shiftID=$2", matchingUsername, shiftID)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("{\"message\": \"Failed to update shift with new username\"}"))
				if err != nil {
					return
				}
				return
			}

			_, err = tx.Exec("UPDATE shifts SET username=$1, trade=false, search_early=false, search_evening=false, search_night=false WHERE shiftID=$2", currentUsername, matchingShiftID)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("{\"message\": \"Failed to update matching shift with new username\"}"))
				if err != nil {
					return
				}
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("{\"message\": \"Failed to commit transaction\"}"))
		if err != nil {
			return
		}
		return
	}

	// Return the updated shift to the frontend
	updatedShift := ShiftReceive{
		Datum: shift.Datum,
		Day:   shift.Day,
		Time:  shift.Time,
		Trade: false, // trade is set to false after swap
		Uid:   shiftID,
		Search: []map[string]interface{}{
			{"name": "früh", "selected": getSearchValue(shift.Search, "früh")},
			{"name": "spät", "selected": getSearchValue(shift.Search, "spät")},
			{"name": "nacht", "selected": getSearchValue(shift.Search, "nacht")},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updatedShift)
	if err != nil {
		return
	}
}

// Helper function to get search value
func getSearchValue(search []map[string]interface{}, name string) bool {
	for _, item := range search {
		if item["name"] == name {
			return item["selected"].(bool)
		}
	}
	return false
}

// Initialize the database
func initDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", DbHost, DbPort, DbUser, DbPassword, DbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_base (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			sessionID TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			timeout TIMESTAMP NOT NULL,
			FOREIGN KEY (username) REFERENCES user_base(username) ON DELETE CASCADE 
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS shifts (
			shiftID TEXT PRIMARY KEY,
			username TEXT,
			date TEXT NOT NULL,
			time TEXT NOT NULL,
			day TEXT NOT NULL,
			TRADE BOOLEAN default false,
			search_early BOOLEAN default false,
			search_evening BOOLEAN default false,
			search_night BOOLEAN default false,
			FOREIGN KEY (username) REFERENCES user_base(username) ON DELETE CASCADE 				  
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Generate a new session ID
func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// Main function
func main() {
	db := initDB()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	app := &App{DB: db}

	http.HandleFunc("/login", app.loginHandler)
	http.HandleFunc("/logout", app.logoutHandler)
	http.Handle("/shifts", app.authMiddleware(http.HandlerFunc(app.shiftHandler)))
	http.Handle("/shifts/", app.authMiddleware(http.HandlerFunc(app.shiftByIDHandler))) // Note the trailing slash

	fmt.Println("Server is running on port 4010")
	log.Fatal(http.ListenAndServe(":4010", nil))
}
