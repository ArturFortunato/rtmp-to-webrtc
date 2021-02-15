package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pkg/errors"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

var handler Handler 

//func startRTMPServer(peerConnection *webrtc.PeerConnection, videoTrack, audioTrack *webrtc.Track) {
func startRTMPServer() {
	log.Println("Starting RTMP Server")
	log.Println("[ARTUR] ON START RTMP SERVER")

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Println("[ARTUR] ON ERROR RESOLVING TCP")

		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Println("[ARTUR] ON ERROR LISTENING TCP")

		log.Panicf("Failed: %+v", err)
	}

	handler = Handler{
		//peerConnection: peerConnection,
		//videoTrack:     videoTrack,
		//audioTrack:     audioTrack,
		videoTracks:    make(map[string][]*webrtc.Track, 0),
		audioTracks:    make(map[string][]*webrtc.Track, 0),
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

func addNewClient(eventId string,  videoTrack, audioTrack *webrtc.Track) {
	log.Println("[ARTUR] NEW CLIENT FOR EVENT")
	log.Println(eventId)
	
	if _, ok := handler.videoTracks[eventId]; !ok {
		handler.videoTracks[eventId] = make([]*webrtc.Track, 0)	
	}

	if _, ok := handler.audioTracks[eventId]; !ok {
		handler.audioTracks[eventId] = make([]*webrtc.Track, 0)	
	}

	handler.videoTracks[eventId] = append(handler.videoTracks[eventId], videoTrack)
	handler.audioTracks[eventId] = append(handler.audioTracks[eventId], audioTrack)

}

type Handler struct {
	rtmp.DefaultHandler
	//peerConnection         *webrtc.PeerConnection
	videoTrack *webrtc.Track
	audioTrack *webrtc.Track
	videoTracks map[string][]*webrtc.Track
	audioTracks map[string][]*webrtc.Track
}

func (h *Handler) OnServe(conn *rtmp.Conn) {
	log.Println("[ARTUR] STARTED SERVING VIDEO")

}

func (h *Handler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Println("[ARTUR] ON CONNECT")

	log.Printf("OnConnect: %#v", cmd)
	return nil
}

func (h *Handler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Println("[ARTUR] ON CREATE STREAM")

	log.Printf("OnCreateStream: %#v", cmd)
	return nil
}

func (h *Handler) OnPublish(timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Println("[ARTUR] ON PUBLISH")

	if cmd.PublishingName == "" {
		return errors.New("PublishingName is empty")
	}
	return nil
}

func (h *Handler) OnAudio(timestamp uint32, payload io.Reader) error {

	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		log.Println("[ARTUR] ON AUDIO ERROR DECODE")

		return err
	}
	
	//log.Println(audio.Data)
	//log.Println(timestamp)
	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		log.Println("[ARTUR] ON AUDIO ERROR COPY")

		return err
	}


	for i := 0; i < len(handler.audioTracks["100"]); i++ {
		err := handler.audioTracks["100"][i].WriteSample(media.Sample{
			Data:    data.Bytes(),
			Samples: media.NSamples(20*time.Millisecond, 48000),
		})

		if err != nil {
			log.Println("[ARTUR] ERROR WRITING AUDIO")
			return err
		}
	}

	return nil

	
}

const headerLengthField = 4

func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {

	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, video.Data); err != nil {
		return err
	}

	outBuf := []byte{}
	videoBuffer := data.Bytes()
	for offset := 0; offset < len(videoBuffer); {
		bufferLength := int(binary.BigEndian.Uint32(videoBuffer[offset : offset+headerLengthField]))
		if offset+bufferLength >= len(videoBuffer) {
			break
		}

		offset += headerLengthField
		outBuf = append(outBuf, []byte{0x00, 0x00, 0x00, 0x01}...)
		outBuf = append(outBuf, videoBuffer[offset:offset+bufferLength]...)

		offset += int(bufferLength)
	}

	for i := 0; i < len(handler.videoTracks["100"]); i++ {
		err := handler.videoTracks["100"][i].WriteSample(media.Sample{
			Data:    outBuf,
			Samples: media.NSamples(time.Second/30, 90000),
		})

		if err != nil {
			log.Println("[ARTUR] ERROR WRITING VIDEO")
			return err	
		}
	}

	return nil
}

func (h *Handler) OnClose() {
	log.Println("[ARTUR] ON CLOSE")
	log.Printf("OnClose")
}
