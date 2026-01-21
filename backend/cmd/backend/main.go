package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Rudraksh121a/BookStore/internal/config"
)

func main() {
	// fmt.Println("Hello, Backend!")
	//load config

	cfg := config.MustLoad()

	//db setup

	//setup router
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	//start server

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	fmt.Println("server started")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("unable to start a server")
	}

}
