package main

import (
	"fmt"
	"github.com/arkadiont/lenslocked/models"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	s := models.NewEmailService(models.SMTPConfig{
		Host: host,
		Port: port,
		User: username,
		Pass: password,
	})
	if err = s.ForgotPassword("arkadiont@gamil.com", "https://lenslockerd.com/reset-pw?token=123"); err != nil {
		panic(err)
	}
	fmt.Println("email sent")
}
