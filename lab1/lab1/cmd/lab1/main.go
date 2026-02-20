package main

import (
	"lab1/internal/api"
	"log"
)

func main() {
	log.Println("App start")
	api.StartServer()
	log.Print("App terminated")
}
