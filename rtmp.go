package main

import (
	// "bytes"
	// "encoding/binary"
	"io"
	"log"
	"net"
	//"time"

	"github.com/pion/webrtc/v2"
	//"github.com/pion/webrtc/v2/pkg/media"
	//"github.com/pkg/errors"
	//flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	//rtmpmsg "github.com/yutopp/go-rtmp/message"
)

var handler Handler 

func startRTMPServer() {
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

func addNewClient(eventId string, audioTrack, videoTrack *webrtc.Track) {
	log.Println("[ARTUR] NEW CLIENT FOR EVENT")
	handler.AddNewClient(eventId, audioTrack, videoTrack)
}

// func (h *Handler) OnAudio(timestamp uint32, payload io.Reader) error {

// 	var audio flvtag.AudioData
// 	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
// 		log.Println("[ARTUR] ON AUDIO ERROR DECODE")

// 		return err
// 	}
	
// 	data := new(bytes.Buffer)
// 	if _, err := io.Copy(data, audio.Data); err != nil {
// 		log.Println("[ARTUR] ON AUDIO ERROR COPY")

// 		return err
// 	}


	// for i := 0; i < len(handler.audioTracks["100"]); i++ {
	// 	err := handler.audioTracks["100"][i].WriteSample(media.Sample{
	// 		Data:    data.Bytes(),
	// 		Samples: media.NSamples(20*time.Millisecond, 48000),
	// 	})

	// 	if err != nil {
	// 		log.Println("[ARTUR] ERROR WRITING AUDIO")
	// 		return err
	// 	}
	// }

// 	return nil

// }

// const headerLengthField = 4

// func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {

// 	var video flvtag.VideoData
// 	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
// 		return err
// 	}

	// data := new(bytes.Buffer)
	// if _, err := io.Copy(data, video.Data); err != nil {
	// 	return err
	// }

// 	outBuf := []byte{}
// 	videoBuffer := data.Bytes()
// 	for offset := 0; offset < len(videoBuffer); {
// 		bufferLength := int(binary.BigEndian.Uint32(videoBuffer[offset : offset+headerLengthField]))
// 		if offset+bufferLength >= len(videoBuffer) {
// 			break
// 		}

// 		offset += headerLengthField
// 		outBuf = append(outBuf, []byte{0x00, 0x00, 0x00, 0x01}...)
// 		outBuf = append(outBuf, videoBuffer[offset:offset+bufferLength]...)

// 		offset += int(bufferLength)
// 	}

// 	for i := 0; i < len(handler.videoTracks["100"]); i++ {
// 		err := handler.videoTracks["100"][i].WriteSample(media.Sample{
// 			Data:    outBuf,
// 			Samples: media.NSamples(time.Second/30, 90000),
// 		})

// 		if err != nil {
// 			log.Println("[ARTUR] ERROR WRITING VIDEO")
// 			return err	
// 		}
// 	}

// 	return nil
// }
