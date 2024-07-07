package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/samocodes/go-priv-notes/crypto"
	"github.com/samocodes/go-priv-notes/helpers"
	"github.com/samocodes/go-priv-notes/types"
)

func UserRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	// returns all notes of a user
	// if user dont exists, create one and fetch the notes
	// TODO: pagination
	r.Get("/notes", func(w http.ResponseWriter, r *http.Request) {
		// gets username, and pin from url query
		username := r.URL.Query().Get("username")
		pin := r.URL.Query().Get("pin")

		if !helpers.IsValidUsername(username) || !helpers.IsValidPin(pin) {
			http.Error(w, "Either username or pin is invalid", http.StatusUnauthorized)
			return
		}

		// check if user exists or create one
		if err := execUser(username, pin, db); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// fetch user's notes
		notes, err := fetchUserNotes(username, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, notes)
	})

	return r
}

func createUser(username, pin string, db *sql.DB) error {
	hashedPin, err := crypto.AESEncrypt(pin)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO users(username, pin) VALUES (?, ?)`, username, hashedPin)
	if err != nil {
		return err
	}

	return nil
}

func execUser(username, pin string, db *sql.DB) error {
	// fetch the user
	var user types.UsersTable

	query := "SELECT username, pin, created_at FROM users WHERE username = ?"
	row := db.QueryRow(query, username)
	err := row.Scan(&user.Username, &user.Pin, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// if doesnt exists, create one
			if err := createUser(username, pin, db); err != nil {
				return err
			}

			return nil
		}

		return fmt.Errorf("invalid user credentials")
	}

	p, err := crypto.AESDecrypt(user.Pin)
	if err != nil || p != pin {
		return fmt.Errorf("invalid user credentials")
	}

	return nil
}

func fetchUserNotes(username string, db *sql.DB) ([]types.NotesTable, error) {
	var notes []types.NotesTable

	query := "SELECT id, content, username, created_at FROM notes WHERE username = ?"
	rows, err := db.Query(query, username)
	if err != nil {
		return notes, errors.New("error while fetching user's notes")
	}
	defer rows.Close()

	for rows.Next() {
		var note types.NotesTable
		if err := rows.Scan(&note.Id, &note.Content, &note.Username, &note.CreatedAt); err != nil {
			continue
		}

		notes = append(notes, note)
	}

	return notes, nil
}
