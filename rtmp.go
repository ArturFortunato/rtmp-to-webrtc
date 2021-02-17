package main

import (
	"bytes"
	//"encoding/binary"
	"io"
	"log"
	"net"
	//"time"

	"github.com/pion/webrtc/v2"
	//"github.com/pion/webrtc/v2/pkg/media"
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

	relayService := NewRelayService()

	handler = Handler{
		//peerConnection: peerConnection,
		//videoTrack:     videoTrack,
		//audioTrack:     audioTrack,
		videoTracks:    make(map[string][]*webrtc.Track, 0),
		audioTracks:    make(map[string][]*webrtc.Track, 0),
		publishingName: "",
		
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

func addNewClient(eventID string,  videoTrack, audioTrack *webrtc.Track) {
	log.Println("[ARTUR] NEW CLIENT FOR EVENT")
	log.Println(eventID)
	
	if _, ok := handler.videoTracks[eventID]; !ok {
		handler.videoTracks[eventID] = make([]*webrtc.Track, 0)	
	}

	if _, ok := handler.audioTracks[eventID]; !ok {
		handler.audioTracks[eventID] = make([]*webrtc.Track, 0)	
	}

	handler.videoTracks[eventID] = append(handler.videoTracks[eventID], videoTrack)
	handler.audioTracks[eventID] = append(handler.audioTracks[eventID], audioTrack)

}

// Handler implementation
type Handler struct {
	rtmp.DefaultHandler
	//peerConnection         *webrtc.PeerConnection
	//videoTrack *webrtc.Track
	//audioTrack *webrtc.Track
	videoTracks map[string][]*webrtc.Track
	audioTracks map[string][]*webrtc.Track
	publishingName string

	//TO TEST
	relayService *RelayService

	//
	conn *rtmp.Conn

	//
	pub *Pub
	sub *Sub
}

func (h *Handler) OnServe(conn *rtmp.Conn) {
	log.Println("[ARTUR] STARTED SERVING VIDEO")
	h.conn = conn
}

func (h *Handler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Println("[ARTUR] ON CONNECT")

	// TODO: check app name to distinguish stream names per apps
	// cmd.Command.App

	log.Printf("OnConnect: %#v", cmd)
	return nil
}

func (h *Handler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Println("[ARTUR] ON CREATE STREAM")

	log.Printf("OnCreateStream: %#v", cmd)
	return nil
}

func (h *Handler) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Println("[ARTUR] ON PUBLISH")
	
	if h.sub != nil {
		return errors.New("Cannot publish to this stream")
	}

	if cmd.PublishingName == "" {
		return errors.New("PublishingName is empty")
	}

	pubsub, err := h.relayService.NewPubsub(cmd.PublishingName)
	if err != nil {
		return errors.Wrap(err, "Failed to create pubsub")
	}

	pub := pubsub.Pub()

	h.pub = pub

	return nil
}

func (h *Handler) OnPlay(ctx *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPlay) error {
	if h.sub != nil {
		return errors.New("Cannot play on this stream")
	}

	pubsub, err := h.relayService.GetPubsub(cmd.StreamName)
	if err != nil {
		return errors.Wrap(err, "Failed to get pubsub")
	}

	sub := pubsub.Sub()
	sub.eventCallback = onEventCallback(h.conn, ctx.StreamID)

	h.sub = sub

	return nil
}

func (h *Handler) OnSetDataFrame(timestamp uint32, data *rtmpmsg.NetStreamSetDataFrame) error {
	log.Println("[ARTUR] I think this should never run")
	r := bytes.NewReader(data.Payload)

	var script flvtag.ScriptData
	if err := flvtag.DecodeScriptData(r, &script); err != nil {
		log.Printf("Failed to decode script data: Err = %+v", err)
		return nil // ignore
	}

	log.Printf("SetDataFrame: Script = %#v", script)

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeScriptData,
		Timestamp: timestamp,
		Data:      &script,
	})

	return nil
}

func (h *Handler) OnAudio(timestamp uint32, payload io.Reader) error {

	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		log.Println("[ARTUR] ON AUDIO ERROR DECODE")

		return err
	}	
	
	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		log.Println("[ARTUR] ON AUDIO ERROR COPY")

		return err
	}
	audio.Data = data

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeAudio,
		Timestamp: timestamp,
		Data:      &audio,
	})

	return nil

}

const headerLengthField = 4

/*func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {

	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, video.Data); err != nil {
		return err
	}
	video.Data = data

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
}*/

func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {
	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	// Need deep copy because payload will be recycled
	flvBody := new(bytes.Buffer)
	if _, err := io.Copy(flvBody, video.Data); err != nil {
		return err
	}
	video.Data = flvBody

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeVideo,
		Timestamp: timestamp,
		Data:      &video,
	})

	return nil
}

func (h *Handler) OnClose() {
	log.Printf("OnClose")

	if h.pub != nil {
		_ = h.pub.Close()
	}

	if h.sub != nil {
		_ = h.sub.Close()
	}
}
