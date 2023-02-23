package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

func main() {
	args := os.Args
	checkArgs(args, 3)
	switch args[1] {
	case "hash":
		hash(args[2])
	case "compare":
		checkArgs(args, 4)
		compare(args[2], args[3])
	default:
		log.Fatalf("Invalid command: %v", args[1])
	}
}

func checkArgs(args []string, expectAtLeast int) {
	if len(args) < expectAtLeast {
		log.Fatalf("unexpected args: %v, need %d args", args, expectAtLeast)
	}
}

func hash(password string) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("err hashing: %v\n", err)
		return
	}
	fmt.Println(string(hashBytes))
}

func compare(password, hash string) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		pwd := fmt.Sprintf(`"%s"`, password)
		fmt.Printf("Password %s is invalid: %v\n", pwd, err)
		return
	}
	fmt.Println("Password is correct")
}
