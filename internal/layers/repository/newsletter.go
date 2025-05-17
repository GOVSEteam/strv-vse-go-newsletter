package repository

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type NewsletterRepository interface {
	ListNewsletters() ([]string, error)
	CreateNewsletter(name string) error
}

type PostgresNewsletterRepo struct {
	db *sql.DB
}

func NewsletterRepo(db *sql.DB) NewsletterRepository {
	return &PostgresNewsletterRepo{db: db}
}

func (r *PostgresNewsletterRepo) ListNewsletters() ([]string, error) {
	rows, err := r.db.Query("SELECT name FROM newsletters ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsletters []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		newsletters = append(newsletters, name)
	}
	return newsletters, rows.Err()
}

func (r *PostgresNewsletterRepo) CreateNewsletter(name string) error {
	_, err := r.db.Exec("INSERT INTO newsletters (name) VALUES ($1)", name)
	return err
}
