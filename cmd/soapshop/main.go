package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"handmade-soap-shop/internal/fixture"
	shopweb "handmade-soap-shop/internal/web"
)

func main() {
	address := os.Getenv("ADDR")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{
		Addr:              address,
		Handler:           shopweb.NewHandler(fixture.NewService()).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("拾香手工皂正在监听 %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
