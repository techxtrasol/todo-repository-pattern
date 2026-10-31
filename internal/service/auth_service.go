package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/techxtrasol/todo-app/internal/config"
	"github.com/techxtrasol/todo-app/internal/db"
	"github.com/techxtrasol/todo-app/internal/middleware"
	"github.com/techxtrasol/todo-app/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh token")
)

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthService struct {
	users    repository.UserRepository
	refresh  repository.RefreshTokenRepository
	cfg      *config.Config
}

func NewAuthService(users repository.UserRepository, refresh repository.RefreshTokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{users: users, refresh: refresh, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (db.User, AuthTokens, error) {
	if email == "" || password == "" {
		return db.User{}, AuthTokens{}, errors.New("email and password are required")
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return db.User{}, AuthTokens{}, ErrEmailTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, AuthTokens{}, err
	}
	user, err := s.users.Create(ctx, email, string(hash), "user")
	if err != nil {
		return db.User{}, AuthTokens{}, err
	}
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return user, AuthTokens{}, err
	}
	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (db.User, AuthTokens, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return db.User{}, AuthTokens{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return db.User{}, AuthTokens{}, ErrInvalidCredentials
	}
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return user, AuthTokens{}, err
	}
	return user, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (AuthTokens, error) {
	hash := hashToken(refreshToken)
	rt, err := s.refresh.GetByHash(ctx, hash)
	if err != nil {
		return AuthTokens{}, ErrInvalidRefresh
	}
	if rt.Revoked {
		return AuthTokens{}, ErrInvalidRefresh
	}
	if time.Now().After(rt.ExpiresAt.Time) {
		return AuthTokens{}, ErrInvalidRefresh
	}
	user, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return AuthTokens{}, ErrInvalidRefresh
	}
	// rotate: revoke old, issue new
	_ = s.refresh.Revoke(ctx, hash)
	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return AuthTokens{}, err
	}
	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.refresh.Revoke(ctx, hashToken(refreshToken))
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (db.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *AuthService) issueTokens(ctx context.Context, user db.User) (AuthTokens, error) {
	now := time.Now()
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
			Issuer:    "todo-app",
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return AuthTokens{}, err
	}

	// opaque refresh token
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return AuthTokens{}, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(raw)
	_ = s.refresh.RevokeAllForUser(ctx, user.ID) // optional: revoke previous
	_, err = s.refresh.Create(ctx, user.ID, hashToken(refreshToken), now.Add(s.cfg.RefreshTokenTTL))
	if err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}