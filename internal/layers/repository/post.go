package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../queries/post/create.sql
var createPostQuery string

//go:embed ../../queries/post/get_by_id.sql
var getPostByIDQuery string

//go:embed ../../queries/post/list_by_newsletter_id.sql
var listPostsByNewsletterIDQuery string

//go:embed ../../queries/post/count_by_newsletter_id.sql
var countPostsByNewsletterIDQuery string

//go:embed ../../queries/post/update.sql
var updatePostQuery string

//go:embed ../../queries/post/delete.sql
var deletePostQuery string

//go:embed ../../queries/post/mark_as_published.sql
var markPostAsPublishedQuery string

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

// toModelPtr converts a dbPost to a *models.Post domain object.
func (dbP *dbPost) toModelPtr() *models.Post {
	m := dbP.toModel()
	return &m
}

// PostRepository defines the interface for post data access operations.
type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error)
	ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error)
	UpdatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, postID string) error
	MarkPostAsPublished(ctx context.Context, postID string, publishedAt time.Time) (*models.Post, error)
}

type postgresPostRepository struct {
	db *pgxpool.Pool
}

// NewPostRepository creates a new instance of postgresPostRepository.
func NewPostRepository(db *pgxpool.Pool) PostRepository {
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
	err := r.db.QueryRow(ctx, createPostQuery,
		post.ID, post.NewsletterID, post.Title, post.Content, post.PublishedAt, post.CreatedAt, post.UpdatedAt,
	).Scan(
		&createdPostDB.ID, &createdPostDB.NewsletterID, &createdPostDB.Title, &createdPostDB.Content,
		&createdPostDB.PublishedAt, &createdPostDB.CreatedAt, &createdPostDB.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("post repo: CreatePost: %w", apperrors.ErrConflict)
		} else if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return nil, fmt.Errorf("post repo: CreatePost: newsletter not found or invalid: %w", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("post repo: CreatePost: scan: %w", err)
	}
	return createdPostDB.toModelPtr(), nil
}

func (r *postgresPostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	var p dbPost
	err := r.db.QueryRow(ctx, getPostByIDQuery, postID).Scan(
		&p.ID, &p.NewsletterID, &p.Title, &p.Content, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("post repo: GetPostByID: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("post repo: GetPostByID: scan: %w", err)
	}
	return p.toModelPtr(), nil
}

func (r *postgresPostRepository) ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error) {
	rows, err := r.db.Query(ctx, listPostsByNewsletterIDQuery, newsletterID, limit, offset)
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
	err = r.db.QueryRow(ctx, countPostsByNewsletterIDQuery, newsletterID).Scan(&totalCount)
	if err != nil {
		return posts, 0, fmt.Errorf("post repo: ListPostsByNewsletterID: count query: %w", err)
	}

	return posts, totalCount, nil
}

func (r *postgresPostRepository) UpdatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	post.UpdatedAt = time.Now().UTC()

	var updatedPostDB dbPost
	err := r.db.QueryRow(ctx, updatePostQuery,
		post.Title, post.Content, post.PublishedAt, post.UpdatedAt, post.ID,
	).Scan(
		&updatedPostDB.ID, &updatedPostDB.NewsletterID, &updatedPostDB.Title, &updatedPostDB.Content,
		&updatedPostDB.PublishedAt, &updatedPostDB.CreatedAt, &updatedPostDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("post repo: UpdatePost: %w", apperrors.ErrPostNotFound)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("post repo: UpdatePost: %w", apperrors.ErrConflict)
		} else if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return nil, fmt.Errorf("post repo: UpdatePost: newsletter not found or invalid: %w", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("post repo: UpdatePost: scan: %w", err)
	}
	return updatedPostDB.toModelPtr(), nil
}

func (r *postgresPostRepository) DeletePost(ctx context.Context, postID string) error {
	cmdTag, err := r.db.Exec(ctx, deletePostQuery, postID)
	if err != nil {
		return fmt.Errorf("post repo: DeletePost: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("post repo: DeletePost: %w", apperrors.ErrPostNotFound)
	}
	return nil
}

func (r *postgresPostRepository) MarkPostAsPublished(ctx context.Context, postID string, publishedAt time.Time) (*models.Post, error) {
	updatedAt := time.Now().UTC()

	var publishedPostDB dbPost
	err := r.db.QueryRow(ctx, markPostAsPublishedQuery, publishedAt, updatedAt, postID).Scan(
		&publishedPostDB.ID, &publishedPostDB.NewsletterID, &publishedPostDB.Title, &publishedPostDB.Content,
		&publishedPostDB.PublishedAt, &publishedPostDB.CreatedAt, &publishedPostDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("post repo: MarkPostAsPublished: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("post repo: MarkPostAsPublished: scan: %w", err)
	}
	return publishedPostDB.toModelPtr(), nil
}
