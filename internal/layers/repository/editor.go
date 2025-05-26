package repository

import (
	"database/sql"
	"errors"
)

// ErrEditorNotFound is returned when an editor is not found
var ErrEditorNotFound = errors.New("editor not found")

type Editor struct {
	ID          string
	FirebaseUID string
	Email       string
}

type EditorRepository interface {
	InsertEditor(firebaseUID, email string) (*Editor, error)
	GetEditorByFirebaseUID(firebaseUID string) (*Editor, error)
}

type PostgresEditorRepo struct {
	db *sql.DB
}

func EditorRepo(db *sql.DB) EditorRepository {
	return &PostgresEditorRepo{db: db}
}

func (r *PostgresEditorRepo) InsertEditor(firebaseUID, email string) (*Editor, error) {
	row := r.db.QueryRow(`INSERT INTO editors (firebase_uid, email) VALUES ($1, $2) RETURNING id, firebase_uid, email`, firebaseUID, email)
	var e Editor
	if err := row.Scan(&e.ID, &e.FirebaseUID, &e.Email); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PostgresEditorRepo) GetEditorByFirebaseUID(firebaseUID string) (*Editor, error) {
	row := r.db.QueryRow(`SELECT id, firebase_uid, email FROM editors WHERE firebase_uid = $1`, firebaseUID)
	var e Editor
	if err := row.Scan(&e.ID, &e.FirebaseUID, &e.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEditorNotFound
		}
		return nil, err
	}
	return &e, nil
}
