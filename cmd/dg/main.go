package main

import (
	"default-gen/app"
	"log"
)

func main() {
	app := app.New()

	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
