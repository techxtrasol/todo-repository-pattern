package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/techxtrasol/todo-app/internal/db"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, role string) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type userRepository struct {
	queries *db.Queries
}

type refreshTokenRepository struct {
	queries *db.Queries
}

func NewAuthRepository(pool *pgxpool.Pool) (UserRepository, RefreshTokenRepository) {
	q := db.New(pool)
	return &userRepository{queries: q}, &refreshTokenRepository{queries: q}
}

func (r *userRepository) Create(ctx context.Context, email, passwordHash, role string) (db.User, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	})
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *refreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.RefreshToken, error) {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return r.queries.GetRefreshToken(ctx, tokenHash)
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.RevokeAllUserRefreshTokens(ctx, userID)
}