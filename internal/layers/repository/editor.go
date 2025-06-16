package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"database/sql"

	"github.com/lib/pq"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

//go:embed queries/editor/insert.sql
var insertEditorQuery string

//go:embed queries/editor/get_by_firebase_uid.sql
var getEditorByFirebaseUIDQuery string

//go:embed queries/editor/get_by_id.sql
var getEditorByIDQuery string

// dbEditor is an internal struct used for scanning database rows.
// It maps directly to the 'editors' table schema.
type dbEditor struct {
	ID          string    `db:"id"`
	FirebaseUID string    `db:"firebase_uid"`
	Email       string    `db:"email"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// toModel converts a dbEditor to a models.Editor domain object.
func (dbEd *dbEditor) toModel() models.Editor {
	return models.Editor{
		ID:          dbEd.ID,
		FirebaseUID: dbEd.FirebaseUID,
		Email:       dbEd.Email,
		CreatedAt:   dbEd.CreatedAt,
		UpdatedAt:   dbEd.UpdatedAt,
	}
}



// EditorRepository defines the interface for editor data access.
type EditorRepository interface {
	InsertEditor(ctx context.Context, firebaseUID, email string) (*models.Editor, error)
	GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error)
	GetEditorByID(ctx context.Context, id string) (*models.Editor, error)
}

// PostgresEditorRepo is the PostgreSQL implementation of EditorRepository.
type PostgresEditorRepo struct {
	db *sql.DB
}

// NewPostgresEditorRepo creates a new PostgresEditorRepo.
func NewPostgresEditorRepo(db *sql.DB) EditorRepository {
	return &PostgresEditorRepo{db: db}
}

// InsertEditor creates a new editor record in the database.
func (r *PostgresEditorRepo) InsertEditor(ctx context.Context, firebaseUID, email string) (*models.Editor, error) {
	var ed dbEditor
	err := r.db.QueryRowContext(ctx, insertEditorQuery, firebaseUID, email).Scan(
		&ed.ID, &ed.FirebaseUID, &ed.Email, &ed.CreatedAt, &ed.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // 23505 is unique_violation
			// This could be for firebase_uid or email depending on table constraints
			return nil, fmt.Errorf("editor repo: InsertEditor: %w", apperrors.ErrConflict)
		}
		return nil, fmt.Errorf("editor repo: InsertEditor: scan: %w", err)
	}
	model := ed.toModel()
	return &model, nil
}

// GetEditorByFirebaseUID retrieves an editor by their Firebase UID.
func (r *PostgresEditorRepo) GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error) {
	var ed dbEditor
	err := r.db.QueryRowContext(ctx, getEditorByFirebaseUIDQuery, firebaseUID).Scan(
		&ed.ID, &ed.FirebaseUID, &ed.Email, &ed.CreatedAt, &ed.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("editor repo: GetEditorByFirebaseUID: %w", apperrors.ErrEditorNotFound)
		}
		return nil, fmt.Errorf("editor repo: GetEditorByFirebaseUID: scan: %w", err)
	}
	model := ed.toModel()
	return &model, nil
}

// GetEditorByID retrieves an editor by their database ID.
func (r *PostgresEditorRepo) GetEditorByID(ctx context.Context, id string) (*models.Editor, error) {
	var ed dbEditor
	err := r.db.QueryRowContext(ctx, getEditorByIDQuery, id).Scan(
		&ed.ID, &ed.FirebaseUID, &ed.Email, &ed.CreatedAt, &ed.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("editor repo: GetEditorByID: %w", apperrors.ErrEditorNotFound)
		}
		return nil, fmt.Errorf("editor repo: GetEditorByID: scan: %w", err)
	}
	model := ed.toModel()
	return &model, nil
}
