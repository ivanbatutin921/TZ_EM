package main

import (
	"root/core/app"

	_ "github.com/lib/pq"
	_"root/module/song/controller"

)

// @title Song Library API
// @version 1.0
// @description API for managing a song library.
// @contact.name Developer Support
// @contact.email developer@example.com
// @host localhost:3000
// @BasePath /
func main() {
	a := app.NewApp()
	err := a.Run()
	if err != nil {
		panic(err)
	}

}


