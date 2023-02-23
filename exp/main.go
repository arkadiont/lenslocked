package main

import (
	"github.com/arkadiont/lenslocked/models"
	"log"
)

func main() {
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("err closing db %v", err)
		}
	}()
	if err = db.Ping(); err != nil {
		panic(err)
	}
	log.Println("connected")

	us := models.NewUserServicePostgres(db)
	user, err := us.Create("a@a.com", "pass123")
	if err != nil {
		panic(err)
		return
	}
	log.Println(user)
}
