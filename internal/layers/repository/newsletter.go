package repository

import (
	"context"
	_ "embed" // Required for //go:embed
	"errors"  // For errors.Is
	"fmt"
	"time"

	"github.com/jackc/pgerrcode" // For pg error codes
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// Embedded SQL queries
//go:embed queries/newsletter/create.sql
var createNewsletterQuery string

//go:embed queries/newsletter/list_by_editor_id.sql
var listNewslettersByEditorIDQuery string

//go:embed queries/newsletter/count_by_editor_id.sql
var countNewslettersByEditorIDQuery string

//go:embed queries/newsletter/get_by_id.sql
var getNewsletterByIDQuery string

//go:embed queries/newsletter/get_by_id_and_editor_id.sql
var getNewsletterByIDAndEditorIDQuery string

//go:embed queries/newsletter/get_by_name_and_editor_id.sql
var getNewsletterByNameAndEditorIDQuery string

//go:embed queries/newsletter/update.sql
var updateNewsletterQuery string

//go:embed queries/newsletter/delete.sql
var deleteNewsletterQuery string

// dbNewsletter represents the database structure for a newsletter.
type dbNewsletter struct {
	ID          string    `db:"id"`
	EditorID    string    `db:"editor_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// toModel converts a dbNewsletter to a models.Newsletter domain object.
func (dbNl *dbNewsletter) toModel() models.Newsletter {
	return models.Newsletter{
		ID:          dbNl.ID,
		EditorID:    dbNl.EditorID,
		Name:        dbNl.Name,
		Description: dbNl.Description,
		CreatedAt:   dbNl.CreatedAt,
		UpdatedAt:   dbNl.UpdatedAt,
	}
}

// NewsletterRepository defines the interface for newsletter data access.
type NewsletterRepository interface {
	ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error)
	CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error)
	GetNewsletterByIDAndEditorID(ctx context.Context, newsletterID string, editorID string) (*models.Newsletter, error)
	UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*models.Newsletter, error)
	DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error
	GetNewsletterByNameAndEditorID(ctx context.Context, name string, editorID string) (*models.Newsletter, error)
	GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error)
}

// PostgresNewsletterRepo is the PostgreSQL implementation of NewsletterRepository.
type PostgresNewsletterRepo struct {
	db *pgxpool.Pool
}

// NewPostgresNewsletterRepo creates a new PostgresNewsletterRepo.
func NewPostgresNewsletterRepo(db *pgxpool.Pool) NewsletterRepository {
	return &PostgresNewsletterRepo{db: db}
}

// ListNewslettersByEditorID fetches a paginated list of newsletters for a specific editor.
func (r *PostgresNewsletterRepo) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error) {
	rows, err := r.db.Query(ctx, listNewslettersByEditorIDQuery, editorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: query: %w", err)
	}
	defer rows.Close()

	var dbNewsletters []dbNewsletter
	for rows.Next() {
		var nl dbNewsletter
		if errScan := rows.Scan(&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt); errScan != nil {
			return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: scan: %w", errScan)
		}
		dbNewsletters = append(dbNewsletters, nl)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: rows error: %w", err)
	}

	newsletters := make([]models.Newsletter, 0, len(dbNewsletters))
	for i := range dbNewsletters {
		newsletters = append(newsletters, dbNewsletters[i].toModel())
	}

	var totalCount int
	err = r.db.QueryRow(ctx, countNewslettersByEditorIDQuery, editorID).Scan(&totalCount)
	if err != nil {
		// Return already fetched newsletters if count fails, but log or wrap error appropriately.
		return newsletters, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: count query: %w", err)
	}

	return newsletters, totalCount, nil
}

// CreateNewsletter creates a new newsletter.
func (r *PostgresNewsletterRepo) CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, createNewsletterQuery, editorID, name, description).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("newsletter repo: CreateNewsletter: %w", apperrors.ErrConflict)
		}
		return nil, fmt.Errorf("newsletter repo: CreateNewsletter: scan: %w", err)
	}
	model := nl.toModel()
	return &model, nil
}

// GetNewsletterByIDAndEditorID fetches a newsletter by its ID and verifies editor ownership.
func (r *PostgresNewsletterRepo) GetNewsletterByIDAndEditorID(ctx context.Context, newsletterID string, editorID string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, getNewsletterByIDAndEditorIDQuery, newsletterID, editorID).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByIDAndEditorID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByIDAndEditorID: scan: %w", err)
	}
	model := nl.toModel()
	return &model, nil
}

// UpdateNewsletter updates a newsletter's name and/or description atomically.
// Uses COALESCE to only update provided fields, eliminating race conditions.
func (r *PostgresNewsletterRepo) UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, updateNewsletterQuery, name, description, newsletterID, editorID).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: UpdateNewsletter: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: UpdateNewsletter: scan: %w", err)
	}
	model := nl.toModel()
	return &model, nil
}

// DeleteNewsletter removes a newsletter by its ID, ensuring it belongs to the editor.
func (r *PostgresNewsletterRepo) DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error {
	cmdTag, err := r.db.Exec(ctx, deleteNewsletterQuery, newsletterID, editorID)
	if err != nil {
		return fmt.Errorf("newsletter repo: DeleteNewsletter: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("newsletter repo: DeleteNewsletter: %w", apperrors.ErrNewsletterNotFound)
	}
	return nil
}

// GetNewsletterByNameAndEditorID fetches a newsletter by its name and editor ID.
func (r *PostgresNewsletterRepo) GetNewsletterByNameAndEditorID(ctx context.Context, name string, editorID string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, getNewsletterByNameAndEditorIDQuery, name, editorID).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByNameAndEditorID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByNameAndEditorID: scan: %w", err)
	}
	model := nl.toModel()
	return &model, nil
}

// GetNewsletterByID fetches a newsletter by its ID.
func (r *PostgresNewsletterRepo) GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, getNewsletterByIDQuery, newsletterID).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByID: scan: %w", err)
	}
	model := nl.toModel()
	return &model, nil
}
