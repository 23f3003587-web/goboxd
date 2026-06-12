package main

import "github.com/thesouldev/goboxd/internal/app"

func main() {
	if err := app.NewService().Run(); err != nil {
		panic(err)
	}
}
