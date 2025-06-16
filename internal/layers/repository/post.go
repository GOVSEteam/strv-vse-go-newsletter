package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"database/sql"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

//go:embed queries/post/create.sql
var createPostQuery string

//go:embed queries/post/get_by_id.sql
var getPostByIDQuery string

//go:embed queries/post/list_by_newsletter_id.sql
var listPostsByNewsletterIDQuery string

//go:embed queries/post/count_by_newsletter_id.sql
var countPostsByNewsletterIDQuery string

//go:embed queries/post/delete.sql
var deletePostQuery string



// PostUpdate defines the fields that can be updated for a post.
// Only non-nil fields will be updated in the database.
type PostUpdate struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

// dbPost is an internal struct used for scanning database rows.
// It maps directly to the 'posts' table schema.
type dbPost struct {
	ID           string     `db:"id"`
	NewsletterID string     `db:"newsletter_id"`
	Title        string     `db:"title"`
	Content      string     `db:"content"`
	PublishedAt  *time.Time `db:"published_at"` // Pointer to handle NULL
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// toModel converts a dbPost to a models.Post domain object.
func (dbP *dbPost) toModel() models.Post {
	return models.Post{
		ID:           dbP.ID,
		NewsletterID: dbP.NewsletterID,
		Title:        dbP.Title,
		Content:      dbP.Content,
		PublishedAt:  dbP.PublishedAt,
		CreatedAt:    dbP.CreatedAt,
		UpdatedAt:    dbP.UpdatedAt,
	}
}



// PostRepository defines the interface for post data access operations.
type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error)
	ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error)
	UpdatePost(ctx context.Context, postID string, updates PostUpdate) (*models.Post, error)
	SetPostPublished(ctx context.Context, postID string, publishedAt time.Time) (*models.Post, error)
	SetPostUnpublished(ctx context.Context, postID string) (*models.Post, error)
	DeletePost(ctx context.Context, postID string) error
}

type postgresPostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new instance of postgresPostRepository.
func NewPostRepository(db *sql.DB) PostRepository {
	return &postgresPostRepository{db: db}
}

func (r *postgresPostRepository) CreatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	now := time.Now().UTC()
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	if post.UpdatedAt.IsZero() {
		post.UpdatedAt = now
	}
	if post.ID == "" {
		post.ID = uuid.NewString()
	}

	var createdPostDB dbPost
	err := r.db.QueryRowContext(ctx, createPostQuery,
		post.ID, post.NewsletterID, post.Title, post.Content, post.PublishedAt, post.CreatedAt, post.UpdatedAt,
	).Scan(
		&createdPostDB.ID, &createdPostDB.NewsletterID, &createdPostDB.Title, &createdPostDB.Content,
		&createdPostDB.PublishedAt, &createdPostDB.CreatedAt, &createdPostDB.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				return nil, fmt.Errorf("post repo: CreatePost: %w", apperrors.ErrConflict)
			case "23503": // foreign_key_violation
				return nil, fmt.Errorf("post repo: CreatePost: newsletter not found or invalid: %w", apperrors.ErrNotFound)
			}
		}
		return nil, fmt.Errorf("post repo: CreatePost: scan: %w", err)
	}
	model := createdPostDB.toModel()
	return &model, nil
}

func (r *postgresPostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	var p dbPost
	err := r.db.QueryRowContext(ctx, getPostByIDQuery, postID).Scan(
		&p.ID, &p.NewsletterID, &p.Title, &p.Content, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post repo: GetPostByID: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("post repo: GetPostByID: scan: %w", err)
	}
	model := p.toModel()
	return &model, nil
}

func (r *postgresPostRepository) ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error) {
	rows, err := r.db.QueryContext(ctx, listPostsByNewsletterIDQuery, newsletterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("post repo: ListPostsByNewsletterID: query: %w", err)
	}
	defer rows.Close()

	var dbPosts []dbPost
	for rows.Next() {
		var p dbPost
		if errScan := rows.Scan(&p.ID, &p.NewsletterID, &p.Title, &p.Content, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt); errScan != nil {
			return nil, 0, fmt.Errorf("post repo: ListPostsByNewsletterID: scan: %w", errScan)
		}
		dbPosts = append(dbPosts, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("post repo: ListPostsByNewsletterID: rows error: %w", err)
	}

	posts := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		posts = append(posts, dbPosts[i].toModel())
	}

	var totalCount int
	err = r.db.QueryRowContext(ctx, countPostsByNewsletterIDQuery, newsletterID).Scan(&totalCount)
	if err != nil {
		return posts, 0, fmt.Errorf("post repo: ListPostsByNewsletterID: count query: %w", err)
	}

	return posts, totalCount, nil
}

func (r *postgresPostRepository) UpdatePost(ctx context.Context, postID string, updates PostUpdate) (*models.Post, error) {
	updatedAt := time.Now().UTC()
	
	// Build dynamic SET clause based on provided fields
	var setParts []string
	var args []interface{}
	argIndex := 1
	
	if updates.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *updates.Title)
		argIndex++
	}
	
	if updates.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *updates.Content)
		argIndex++
	}
	

	
	if len(setParts) == 0 {
		return nil, fmt.Errorf("post repo: UpdatePost: no fields to update")
	}
	
	// Always update updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, updatedAt)
	argIndex++
	
	// Add postID as final parameter
	args = append(args, postID)
	
	// Build final query
	query := fmt.Sprintf(`
		UPDATE posts 
		SET %s 
		WHERE id = $%d 
		RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)
	
	var updatedPostDB dbPost
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&updatedPostDB.ID, &updatedPostDB.NewsletterID, &updatedPostDB.Title, &updatedPostDB.Content,
		&updatedPostDB.PublishedAt, &updatedPostDB.CreatedAt, &updatedPostDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post repo: UpdatePost: %w", apperrors.ErrPostNotFound)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation
				return nil, fmt.Errorf("post repo: UpdatePost: %w", apperrors.ErrConflict)
			case "23503": // foreign_key_violation
				return nil, fmt.Errorf("post repo: UpdatePost: newsletter not found or invalid: %w", apperrors.ErrNotFound)
			}
		}
		return nil, fmt.Errorf("post repo: UpdatePost: scan: %w", err)
	}
	model := updatedPostDB.toModel()
	return &model, nil
}

func (r *postgresPostRepository) SetPostPublished(ctx context.Context, postID string, publishedAt time.Time) (*models.Post, error) {
	updatedAt := time.Now().UTC()

	var updatedPostDB dbPost
	err := r.db.QueryRowContext(ctx, `
		UPDATE posts 
		SET published_at = $1, updated_at = $2 
		WHERE id = $3 
		RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at`,
		publishedAt, updatedAt, postID,
	).Scan(
		&updatedPostDB.ID, &updatedPostDB.NewsletterID, &updatedPostDB.Title, &updatedPostDB.Content,
		&updatedPostDB.PublishedAt, &updatedPostDB.CreatedAt, &updatedPostDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post repo: SetPostPublished: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("post repo: SetPostPublished: %w", err)
	}
	model := updatedPostDB.toModel()
	return &model, nil
}

func (r *postgresPostRepository) SetPostUnpublished(ctx context.Context, postID string) (*models.Post, error) {
	updatedAt := time.Now().UTC()

	var updatedPostDB dbPost
	err := r.db.QueryRowContext(ctx, `
		UPDATE posts 
		SET published_at = NULL, updated_at = $1 
		WHERE id = $2 
		RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at`,
		updatedAt, postID,
	).Scan(
		&updatedPostDB.ID, &updatedPostDB.NewsletterID, &updatedPostDB.Title, &updatedPostDB.Content,
		&updatedPostDB.PublishedAt, &updatedPostDB.CreatedAt, &updatedPostDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post repo: SetPostUnpublished: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("post repo: SetPostUnpublished: %w", err)
	}
	model := updatedPostDB.toModel()
	return &model, nil
}

func (r *postgresPostRepository) DeletePost(ctx context.Context, postID string) error {
	cmdTag, err := r.db.ExecContext(ctx, deletePostQuery, postID)
	if err != nil {
		return fmt.Errorf("post repo: DeletePost: exec: %w", err)
	}
	rowsAffected, err := cmdTag.RowsAffected()
	if err != nil {
		return fmt.Errorf("post repo: DeletePost: checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post repo: DeletePost: %w", apperrors.ErrPostNotFound)
	}
	return nil
}
