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
}

func NewUserServicePostgres(db *sql.DB) UserService {
	return userServicePostgres{
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
	email = strings.ToLower(email) // postgres is not case sensitive
	binHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user := User{
		Email:        email,
		PasswordHash: string(binHash),
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
