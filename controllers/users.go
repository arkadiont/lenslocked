package controllers

import (
	"fmt"
	"github.com/arkadiont/lenslocked/context"
	"github.com/arkadiont/lenslocked/models"
	"log"
	"net/http"
)

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService    models.UserService
	SessionService models.SessionService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.New.Execute(w, r, data)
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
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Println(err)
		// TODO: long term, we should show a warning about not being able to sign the user in.
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data = struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.SignIn.Execute(w, r, data)
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
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("authenticate user err: %v", err)
		http.Error(w, "Something was wrong.", http.StatusInternalServerError)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.Delete(token)
	if err != nil {
		log.Printf("processSignOut err: %v", err)
		http.Error(w, "Something was wrong.", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

type UserMiddleware struct {
	SessionService models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithUser(r.Context(), user)))
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := context.User(r.Context()); user == nil {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
