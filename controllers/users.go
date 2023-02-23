package controllers

import (
	"fmt"
	"github.com/arkadiont/lenslocked/models"
	"log"
	"net/http"
)

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService models.UserService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.New.Execute(w, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		log.Printf("create user err: %v", err)
		http.Error(w, "Something was wrong.", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User created: %v", user)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.SignIn.Execute(w, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Email    string
		Password string
	}{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		log.Printf("authenticate user err: %v", err)
		http.Error(w, "Something was wrong.", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User authenticated: %v", user)
}
