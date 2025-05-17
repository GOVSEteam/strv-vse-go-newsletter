package repository

import (
	"database/sql"
)

type Editor struct {
	ID          string
	FirebaseUID string
	Email       string
}

type EditorRepository interface {
	InsertEditor(firebaseUID, email string) (*Editor, error)
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
