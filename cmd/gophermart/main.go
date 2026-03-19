package main

import (
	"log"

	"github.com/LemuriiL/GopherMart/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
