package main

import (
	"fmt"
	"github.com/arkadiont/lenslocked/controllers"
	"github.com/arkadiont/lenslocked/templates"
	"github.com/arkadiont/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))

	r.Get("/contact", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))

	r.Get("/faq", controllers.FAQ(
		views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

	usersC := controllers.Users{}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	r.Get("/signup", usersC.New)

	//r.Get("/signup", controllers.StaticHandler(
	//	views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	fmt.Println("Starting server on :3000...")
	fmt.Printf("err: %v", http.ListenAndServe(":3000", r))

}

// examples
//func paramExample(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprint(w, fmt.Sprintf("<h1>URLParam is %s</h1>", chi.URLParam(r, "my_key")))
//}
//
//
//func exerciseTemplates(w http.ResponseWriter, r *http.Request) {
//	user := struct {
//		Optional string
//		Name string
//		Age int
//		Money float64
//		Hobbies []string
//		MyMap map[string]int
//		Meta struct{
//			Visit int
//		}
//	}{
//		Optional: "nil",
//		Name:    "Pedro",
//		Age:     25,
//		Money:   100.2,
//		Hobbies: []string{"read", "write"},
//		MyMap: map[string]int{
//			"field1": 1,
//			"field2": 2,
//		},
//		Meta: struct{ Visit int }{Visit: 2},
//	}
//	executeTemplate(w, filepath.Join("templates", "exerciseTemplates.gohtml"), user)
//}
