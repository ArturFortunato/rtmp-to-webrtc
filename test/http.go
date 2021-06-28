package main

import (
	"log"
	"net/http"
)

func httpServer() {
	httpAddr := "localhost:8080"

	http.HandleFunc("/createPeerConnection", createPeerConnection)

	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
