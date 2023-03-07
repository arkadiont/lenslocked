package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/arkadiont/lenslocked/rand"
)

const (
	// The minimum number of bytes to be used for each session token.
	MinBytesPerToken = 32
)

type Session struct {
	ID     uint
	UserId uint
	// Token is only set when creating a new session. When look up a session this will be left empty,
	// as we only store the hash(token) in db
	Token     string
	TokenHash string
}

type SessionService interface {
	Create(userID uint) (*Session, error)
	User(token string) (*User, error)
	Delete(token string) error
}

type sessionOption func(*sessionService)

func WithBytesPerToken(bytesPerToken int) sessionOption {
	return func(s *sessionService) {
		if bytesPerToken > s.BytesPerToken {
			s.BytesPerToken = bytesPerToken
		}
	}
}

func NewSessionServicePostgres(db *sql.DB, opts ...sessionOption) SessionService {
	s := sessionService{
		DB:            db,
		BytesPerToken: MinBytesPerToken,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

type sessionService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be used.
	BytesPerToken int
}

func (ss sessionService) Create(userID uint) (*Session, error) {
	token, err := rand.String(ss.BytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	session := Session{
		UserId:    userID,
		Token:     token,
		TokenHash: ss.hash(token),
	}
	row := ss.DB.QueryRow(`
		INSERT INTO session (user_id, token_hash)
		VALUES ($1, $2) ON CONFLICT DO 
		UPDATE SET token_hash = $2 RETURNING id;`, session.UserId, session.TokenHash)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &session, nil
}

func (ss sessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	_, err := ss.DB.Exec(`DELETE FROM session WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (ss sessionService) User(token string) (*User, error) {
	var user User
	tokenHash := ss.hash(token)
	row := ss.DB.QueryRow(`SELECT u.id, u.email, u.password_hash FROM users u, session s
	WHERE u.id = s.user_id AND s.token_hash = $1`, tokenHash)
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	return &user, nil
}

func (ss sessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
