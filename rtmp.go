package main

import (
	"io"
	"log"
	"net"

	"github.com/ArturFortunato/go-rtmp"
	"github.com/pion/webrtc/v2"
)

var handler Handler

func startRTMPServer() {
	log.Println("Starting RTMP server...")
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	relayService := NewRelayService()

	handler = Handler{
		relayService: relayService,
	}

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			log.Println("[ARTUR] ON NEW STREAM RECEIVED")

			return conn, &rtmp.ConnConfig{
				Handler: &handler,

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},
			}
		},
	})
	if err := srv.Serve(listener); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}

func addNewClient(eventID string, audioTrack, videoTrack *webrtc.Track) {
	handler.AddNewClient(eventID, audioTrack, videoTrack)
}

func closeConnection(streamID string) {
	handler.OnCloseByController(streamID)
}
