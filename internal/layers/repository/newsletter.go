package repository

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Newsletter struct {
	ID          string
	EditorID    string
	Name        string
	Description string
	CreatedAt   string
}

type NewsletterRepository interface {
	ListNewsletters() ([]Newsletter, error)
	CreateNewsletter(editorID, name, description string) (*Newsletter, error)
}

type PostgresNewsletterRepo struct {
	db *sql.DB
}

func NewsletterRepo(db *sql.DB) NewsletterRepository {
	return &PostgresNewsletterRepo{db: db}
}

func (r *PostgresNewsletterRepo) ListNewsletters() ([]Newsletter, error) {
	rows, err := r.db.Query("SELECT id, editor_id, name, description, created_at FROM newsletters ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsletters []Newsletter
	for rows.Next() {
		var n Newsletter
		if err := rows.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt); err != nil {
			return nil, err
		}
		newsletters = append(newsletters, n)
	}
	return newsletters, rows.Err()
}

func (r *PostgresNewsletterRepo) CreateNewsletter(editorID, name, description string) (*Newsletter, error) {
	row := r.db.QueryRow(`INSERT INTO newsletters (editor_id, name, description) VALUES ($1, $2, $3) RETURNING id, editor_id, name, description, created_at`, editorID, name, description)
	var n Newsletter
	if err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt); err != nil {
		return nil, err
	}
	return &n, nil
}
