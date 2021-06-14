package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/pion/webrtc/v2"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	go startRTMPServer()

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/createPeerConnection", createPeerConnection)
	http.HandleFunc("/test", test)
	panic(http.ListenAndServe(":8080", nil))
}

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ABCD"))
}

func createPeerConnection(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	audioTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypePCMA, rand.Uint32(), "audio", "pion")
	if err != nil {
		panic(err)
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		panic(err)
	}

	var offer webrtc.SessionDescription

	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	response, err := json.Marshal(answer)
	if err != nil {
		panic(err)
	}

	if _, err := w.Write(response); err != nil {
		panic(err)
	}

	if len(r.URL.Query()["eventID"]) == 0 {
		print(r.URL.Query())
		panic("erro aqui")
	} else {
		eventID := r.URL.Query()["eventID"][0]
		addNewClient(eventID, audioTrack, videoTrack)
	}
}
