package main

import (
	"fmt"
	"github.com/arkadiont/lenslocked/controllers"
	"github.com/arkadiont/lenslocked/models"
	"github.com/arkadiont/lenslocked/templates"
	"github.com/arkadiont/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"log"
	"net/http"
)

func main() {
	// setup db
	cfg := models.DefaultPostgresConfig()
	db, err := models.OpenCheckConn(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("err closing db %v", err)
		}
	}()

	// services
	userSrv := models.NewUserServicePostgres(db)
	sessionSrv := models.NewSessionServicePostgres(db)

	// middlewares
	CSRF := csrf.Protect(
		[]byte("ASDFGHJKLZXCVBNMQWERTUIOP1234567"),
		csrf.Secure(false),
	)
	userMiddleware := controllers.UserMiddleware{SessionService: sessionSrv}

	// controllers
	usersC := controllers.Users{
		UserService:    userSrv,
		SessionService: sessionSrv,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
	))

	// build router
	r := chi.NewRouter()
	r.Use(
		CSRF,
		userMiddleware.SetUser,
	)
	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(
		views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMiddleware.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// run server
	fmt.Println("Starting server on :3000...")
	fmt.Printf("err: %v", http.ListenAndServe(":3000", r))
}
