package models

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type User struct {
	ID           uint
	Email        string
	PasswordHash string
}

type UserService interface {
	Create(email, password string) (*User, error)
	Authenticate(email, password string) (*User, error)
	UpdatePassword(userId uint, password string) error
}

func NewUserServicePostgres(db *sql.DB) UserService {
	return &userServicePostgres{
		DB: db,
	}
}

type userServicePostgres struct {
	DB *sql.DB
}

func (us userServicePostgres) Authenticate(email, password string) (*User, error) {
	user := User{
		Email: strings.ToLower(email), // postgres is not case sensitive,
	}
	row := us.DB.QueryRow(`
		SELECT id, password_hash FROM users
		WHERE email=$1`, user.Email)

	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	return &user, nil
}

func (us userServicePostgres) Create(email, password string) (*User, error) {
	var err error
	user := User{
		Email: strings.ToLower(email), // postgres is not case sensitive
	}
	user.PasswordHash, err = us.generateFromPassword(password)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	row := us.DB.QueryRow(`
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2) RETURNING id`, user.Email, user.PasswordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (us userServicePostgres) UpdatePassword(userId uint, password string) error {
	hash, err := us.generateFromPassword(password)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if _, err = us.DB.Exec(`
		UPDATE users SET password_hash = $2 
		WHERE id = $1;`, userId, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (us userServicePostgres) generateFromPassword(password string) (string, error) {
	binHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(binHash), nil
}
