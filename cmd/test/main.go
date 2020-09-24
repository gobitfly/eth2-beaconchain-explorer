package main

import (
	"log"
	"path/filepath"
)

func main() {
	log.Println("matching paths")

	matches, err := filepath.Glob("templates/login.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("found matches", matches)
}
