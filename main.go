package main

import (
	"fmt"
	"github.com/arkadiont/lenslocked/controllers"
	"github.com/arkadiont/lenslocked/models"
	"github.com/arkadiont/lenslocked/templates"
	"github.com/arkadiont/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (cfg config, err error) {
	if err = godotenv.Load(); err != nil {
		return
	}
	cfg.PSQL.Host = os.Getenv("PSQL_HOST")
	cfg.PSQL.Port = os.Getenv("PSQL_PORT")
	cfg.PSQL.Database = os.Getenv("PSQL_DATABASE")
	cfg.PSQL.User = os.Getenv("PSQL_USERNAME")
	cfg.PSQL.Password = os.Getenv("PSQL_PASSWORD")

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	if cfg.CSRF.Secure, err = strconv.ParseBool(os.Getenv("CSRF_SECURE")); err != nil {
		return
	}

	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")

	if cfg.SMTP.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT")); err != nil {
		return
	}
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	cfg.SMTP.User = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Pass = os.Getenv("SMTP_PASSWORD")

	return
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}
	// setup db
	db, err := models.OpenCheckConn(cfg.PSQL)
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
	passSrv := models.NewPasswordResetService(db)
	emailSrv := models.NewEmailService(cfg.SMTP)

	// middlewares
	CSRF := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
	)
	userMiddleware := controllers.UserMiddleware{SessionService: sessionSrv}

	// controllers
	usersC := controllers.Users{
		UserService:     userSrv,
		SessionService:  sessionSrv,
		PasswordService: passSrv,
		EmailService:    emailSrv,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"forgot-pw.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"reset-pw.gohtml", "tailwind.gohtml",
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
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMiddleware.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// run server
	fmt.Printf("Starting server on %s...\n", cfg.Server.Address)
	if err = http.ListenAndServe(cfg.Server.Address, r); err != nil {
		panic(err)
	}
}
