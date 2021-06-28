package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/GRVYDEV/lightspeed-webrtc/ws"
	"github.com/gorilla/websocket"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

var (
	addr     = flag.String("addr", "localhost", "http service address")
	ip       = flag.String("ip", "none", "IP address for webrtc")
	httpPort = flag.Int("ws-port", 8080, "Port for websocket")
	rtpPort  = flag.Int("rtp-port", 65535, "Port for RTP")
	ports    = flag.String("ports", "20000-20500", "Port range for webrtc")
	sslCert  = flag.String("ssl-cert", "", "Ssl cert for websocket (optional)")
	sslKey   = flag.String("ssl-key", "", "Ssl key for websocket (optional)")
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	hub *ws.Hub
)

//var pubsubs []*Pubsub
var pubsub *Pubsub

func addNewClient(eventID string) {
	//pubsub := nil

	// Search pubsub from event with eventID
	/*for index, pb := range pubsubs {
		if (pb->name == eventID) {
			pubsub = pb
			break
		}
	}

	// No pubsub found
	if (pubsub == nil) {
		pubsub = NewPubsub{

		}
	}*/

	pubsub.addNewClient()
}

func createPeerConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
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

	eventID := r.URL.Query()["eventID"][0]

	addNewClient(eventID)
}

func waitForRTPPackets() {
	// Open a UDP Listener for RTP Packets on port 65535
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(*addr), Port: *rtpPort})
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Waiting for RTP Packets")

	inboundRTPPacket := make([]byte, 4096) // UDP MTU

	// Read RTP packets forever and send them to the WebRTC Client
	for {

		n, _, err := listener.ReadFrom(inboundRTPPacket)

		if err != nil {
			fmt.Printf("error during read: %s", err)
			panic(err)
		}

		packet := &rtp.Packet{}
		if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
		}

		// TODO: send to pubsub subs
		if packet.Header.PayloadType == 96 {
			for _, sub := range pubsub.subs {
				if _, writeErr := sub.videoTrack.Write(inboundRTPPacket[:n]); writeErr != nil {
					panic(writeErr)
				}
			}
		} else if packet.Header.PayloadType == 97 {
			for _, sub := range pubsub.subs {
				if _, writeErr := sub.audioTrack.Write(inboundRTPPacket[:n]); writeErr != nil {
					panic(writeErr)
				}
			}
		}

	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	// hub = ws.NewHub()
	// go hub.Run()

	// start HTTP server
	go httpServer()

	go waitForRTPPackets()

}
