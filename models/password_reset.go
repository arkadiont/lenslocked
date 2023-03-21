package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/arkadiont/lenslocked/rand"
	"strings"
	"time"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID     int
	UserID uint
	// Token is only set when a PasswordReset is being created
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService interface {
	Create(email string) (*PasswordReset, error)
	Consume(token string) (*User, error)
}

type passResetOption func(service *passwordResetService)

func WithBytesPerTokenReset(bytesPerToken int) passResetOption {
	return func(s *passwordResetService) {
		if bytesPerToken > s.BytesPerToken {
			s.BytesPerToken = bytesPerToken
		}
	}
}

func NewPasswordResetService(db *sql.DB, opts ...passResetOption) PasswordResetService {
	s := &passwordResetService{
		DB:            db,
		BytesPerToken: MinBytesPerToken,
		Duration:      DefaultResetDuration,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type passwordResetService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each password reset token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be used.
	BytesPerToken int
	// Duration is the amount of time that a PasswordReset is valid for. Default DefaultResetDuration
	Duration time.Duration
}

func (p passwordResetService) Create(email string) (*PasswordReset, error) {
	email = strings.ToLower(email)
	var userId uint
	row := p.DB.QueryRow(`SELECT id FROM users WHERE email = $1;`, email)
	err := row.Scan(&userId)
	if err != nil {
		// TODO consider return specific err when user not exists
		return nil, fmt.Errorf("create: %w", err)
	}
	token, err := rand.String(p.BytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	duration := p.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	pwReset := PasswordReset{
		UserID:    userId,
		Token:     token,
		TokenHash: p.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = p.DB.QueryRow(`
		INSERT INTO password_reset (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3) ON CONFLICT(user_id) DO 
		UPDATE SET token_hash = $2, expires_at = $3 RETURNING id;`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create password_reset: %w", err)
	}
	return &pwReset, nil
}

func (p passwordResetService) Consume(token string) (*User, error) {
	var user User
	var pwReset PasswordReset
	hash := p.hash(token)
	row := p.DB.QueryRow(`SELECT p.id, p.expires_at, u.id, u.email, u.password_hash
	FROM users u, password_reset p 
	WHERE u.id = p.user_id and p.token_hash = $1;`, hash)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}
	err = p.delete(pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	return &user, nil
}

func (p passwordResetService) delete(id int) error {
	_, err := p.DB.Exec(`DELETE FROM password_reset WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (p passwordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
