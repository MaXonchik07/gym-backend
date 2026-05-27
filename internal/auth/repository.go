package auth

import (
	"context"

	"github.com/MaXonchik07/gym-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, phone, password_hash, role, membership_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, join_date, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		user.FirstName, user.LastName, user.Email, user.Phone,
		user.PasswordHash, user.Role, user.MembershipType,
	).Scan(&user.ID, &user.JoinDate, &user.CreatedAt, &user.UpdatedAt)
	return err
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, first_name, last_name, email, phone, password_hash, role, membership_type, join_date, created_at, updated_at FROM users WHERE email = $1`
	row := r.pool.QueryRow(ctx, query, email)
	user := &models.User{}
	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Role, &user.MembershipType,
		&user.JoinDate, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *repository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET first_name=$1, last_name=$2, phone=$3, membership_type=$4, updated_at=NOW()
		WHERE id=$5
	`
	_, err := r.pool.Exec(ctx, query, user.FirstName, user.LastName, user.Phone, user.MembershipType, user.ID)
	return err
}
