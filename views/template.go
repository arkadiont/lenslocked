package views

import (
	"bytes"
	"fmt"
	"github.com/arkadiont/lenslocked/context"
	"github.com/arkadiont/lenslocked/models"
	"github.com/gorilla/csrf"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
)

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	tpl := template.New(pattern[0])
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("currentUser not implemented")
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, pattern...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing fs template: %w", err)
	}

	return Template{
		htmlTpl: tpl,
	}, err
}

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("err cloning template %v", err)
		http.Error(w, "There was an error executing template", http.StatusInternalServerError)
		return
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
		},
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buff bytes.Buffer
	if err = tpl.Execute(&buff, data); err != nil {
		log.Printf("err executing template %v", err)
		http.Error(w, "There was an error executing template", http.StatusInternalServerError)
		return
	}
	_, _ = io.Copy(w, &buff)
}
