package repository

import (
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)


type Newsletter struct {
	ID          string    `json:"id"`
	EditorID    string    `json:"editor_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NewsletterRepository interface {
	// ListNewsletters() ([]Newsletter, error) // Old signature
	ListNewslettersByEditorID(editorID string, limit int, offset int) ([]Newsletter, int, error) // New: added editorID, limit, offset, returns total count for pagination
	CreateNewsletter(editorID, name, description string) (*Newsletter, error)
	GetNewsletterByIDAndEditorID(newsletterID string, editorID string) (*Newsletter, error)
	UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*Newsletter, error)
	DeleteNewsletter(newsletterID string, editorID string) error
	GetNewsletterByNameAndEditorID(name string, editorID string) (*Newsletter, error)
}

type PostgresNewsletterRepo struct {
	db *sql.DB
}

func NewsletterRepo(db *sql.DB) NewsletterRepository {
	return &PostgresNewsletterRepo{db: db}
}


// ListNewsletters is deprecated, use ListNewslettersByEditorID
// func (r *PostgresNewsletterRepo) ListNewsletters() ([]Newsletter, error) {
// 	rows, err := r.db.Query("SELECT id, editor_id, name, description, created_at, updated_at FROM newsletters ORDER BY created_at DESC")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 
// 	var newsletters []Newsletter
// 	for rows.Next() {
// 		var n Newsletter
// 		if err := rows.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt); err != nil {
// 			return nil, err
// 		}
// 		newsletters = append(newsletters, n)
// 	}
// 	return newsletters, rows.Err()
// }

// ListNewslettersByEditorID fetches a paginated list of newsletters for a specific editor.
// It also returns the total count of newsletters for that editor (for pagination metadata).
func (r *PostgresNewsletterRepo) ListNewslettersByEditorID(editorID string, limit int, offset int) ([]Newsletter, int, error) {
	// Query for the paginated list
	listQuery := `
		SELECT id, editor_id, name, description, created_at, updated_at 
		FROM newsletters 
		WHERE editor_id = $1 
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(listQuery, editorID, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var newsletters []Newsletter
	for rows.Next() {
		var n Newsletter
		if errScan := rows.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt); errScan != nil {
			return nil, 0, errScan
		}
		newsletters = append(newsletters, n)
	}
	if err = rows.Err(); err != nil {
		return newsletters, 0, err
	}

	// Query for the total count of items for that editor
	countQuery := `SELECT COUNT(*) FROM newsletters WHERE editor_id = $1`
	var totalCount int
	err = r.db.QueryRow(countQuery, editorID).Scan(&totalCount)
	if err != nil {
		return newsletters, 0, err // Return already fetched newsletters if count fails, or handle error differently
	}

	return newsletters, totalCount, nil
}

func (r *PostgresNewsletterRepo) CreateNewsletter(editorID, name, description string) (*Newsletter, error) {
	query := `
		INSERT INTO newsletters (editor_id, name, description) 
		VALUES ($1, $2, $3) 
		RETURNING id, editor_id, name, description, created_at, COALESCE(updated_at, created_at) as updated_at`

	row := r.db.QueryRow(query, editorID, name, description)
	var n Newsletter
	if err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetNewsletterByIDAndEditorID fetches a newsletter by its ID and verifies editor ownership.
func (r *PostgresNewsletterRepo) GetNewsletterByIDAndEditorID(newsletterID string, editorID string) (*Newsletter, error) {
	query := `
		SELECT id, editor_id, name, description, created_at, updated_at 
		FROM newsletters 
		WHERE id = $1 AND editor_id = $2`
	row := r.db.QueryRow(query, newsletterID, editorID)
	
	var n Newsletter
	err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a specific "not found" error
		}
		return nil, err
	}
	return &n, nil
}

// UpdateNewsletter updates a newsletter's name and/or description and its updated_at timestamp.
// It ensures that the update is performed by the owner.
func (r *PostgresNewsletterRepo) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*Newsletter, error) {
	// Build the SET part of the query dynamically based on which fields are provided
	// This is a bit more complex with pure database/sql.
	// For simplicity and security (preventing SQL injection if not careful with string building),
	// we will update both if provided, or keep existing if nil.
	// A more robust way would be to build the query string or use a query builder.

	// First, fetch the current newsletter to get existing values if not provided for update
	// and to ensure it exists and belongs to the editor.
	currentNewsletter, err := r.GetNewsletterByIDAndEditorID(newsletterID, editorID)
	if err != nil {
		return nil, err // Error fetching (e.g., DB connection issue)
	}
	if currentNewsletter == nil {
		return nil, sql.ErrNoRows // Newsletter not found or not owned by editor
	}

	// Use current values if new ones are not provided
	updatedName := currentNewsletter.Name
	if name != nil {
		updatedName = *name
	}

	updatedDescription := currentNewsletter.Description
	if description != nil {
		updatedDescription = *description
	}
	
	// Ensure name is not empty if provided for update
	if name != nil && *name == "" {
		// This validation should ideally be in the service or handler,
		// but as a safeguard here if name is being explicitly set to empty.
		// For this example, we'll allow it, assuming higher layers validate.
		// If name is mandatory for a newsletter, the table schema should reflect that.
	}


	query := `
		UPDATE newsletters
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND editor_id = $4
		RETURNING id, editor_id, name, description, created_at, updated_at`
	
	row := r.db.QueryRow(query, updatedName, updatedDescription, newsletterID, editorID)
	
	var n Newsletter
	errScan := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if errScan != nil {
		// If the RETURNING clause didn't find a row (e.g., id or editor_id didn't match,
		// which shouldn't happen if GetNewsletterByIDAndEditorID passed),
		// sql.ErrNoRows would be returned here.
		return nil, errScan
	}
	return &n, nil
}

// DeleteNewsletter removes a newsletter by its ID, ensuring it belongs to the editor.
func (r *PostgresNewsletterRepo) DeleteNewsletter(newsletterID string, editorID string) error {
	query := `DELETE FROM newsletters WHERE id = $1 AND editor_id = $2`
	result, err := r.db.Exec(query, newsletterID, editorID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err // Error checking rows affected
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Indicates newsletter not found or not owned by editor
	}

	return nil
}

// GetNewsletterByNameAndEditorID fetches a newsletter by its name and editor ID.
// Returns nil, nil if not found (to distinguish from other errors).
func (r *PostgresNewsletterRepo) GetNewsletterByNameAndEditorID(name string, editorID string) (*Newsletter, error) {
	query := `
		SELECT id, editor_id, name, description, created_at, updated_at 
		FROM newsletters 
		WHERE name = $1 AND editor_id = $2`
	row := r.db.QueryRow(query, name, editorID)

	var n Newsletter
	err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, no error to bubble up, service will interpret this
		}
		return nil, err // Other DB error
	}
	return &n, nil
}
