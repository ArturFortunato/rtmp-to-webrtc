package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"strconv"

	"github.com/ArturFortunato/go-rtmp"
	rtmpmsg "github.com/ArturFortunato/go-rtmp/message"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
	flvtag "github.com/yutopp/go-flv/tag"
)

var _ rtmp.Handler = (*Handler)(nil)

type Handler struct {
	rtmp.DefaultHandler
	relayService *RelayService

	conn *rtmp.Conn

	pub *Pub
	sub *Sub
}

func (h *Handler) AddNewClient(streamID string, audioTrack, videoTrack *webrtc.Track) error {
	pubsub, err := h.relayService.GetPubsub(streamID)
	if err != nil {
		return errors.Wrap(err, "Failed to get pubsub")
	}

	sub := pubsub.Sub(audioTrack, videoTrack)
	sub.eventCallback = onEventCallback(sub)

	h.sub = sub

	return nil
}

func (h *Handler) OnServe(conn *rtmp.Conn) {
	log.Println("STARTED ONSERVE")
	h.conn = conn
}

func (h *Handler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Printf("OnConnect: %#v", cmd)

	return nil
}

func (h *Handler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Printf("OnCreateStream: %#v", cmd)
	return nil
}

func (h *Handler) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Printf("OnPublish: %#v", cmd)

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

func (h *Handler) OnSetDataFrame(timestamp uint32, data *rtmpmsg.NetStreamSetDataFrame) error {
	r := bytes.NewReader(data.Payload)

	var script flvtag.ScriptData
	if err := flvtag.DecodeScriptData(r, &script); err != nil {
		log.Printf("Failed to decode script data: Err = %+v", err)
		return nil
	}

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeScriptData,
		Timestamp: timestamp,
		Data:      &script,
	}, nil)

	return nil
}

func (h *Handler) OnAudio(timestamp uint32, payload io.Reader, streamID uint32) error {

	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		return err
	}

	eventID := strconv.FormatUint(uint64(streamID), 10)
	pubsub, _ := h.relayService.GetPubsub(eventID)

	_ = pubsub.GetPub().Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeAudio,
		Timestamp: timestamp,
		Data:      &audio,
	}, data.Bytes())

	return nil
}

const headerLengthField = 4

func (h *Handler) OnVideo(timestamp uint32, payload io.Reader, streamID uint32) error {

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

	video.Data = data

	eventID := strconv.FormatUint(uint64(streamID), 10)
	pubsub, _ := h.relayService.GetPubsub(eventID)

	_ = pubsub.GetPub().Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeVideo,
		Timestamp: timestamp,
		Data:      &video,
	},
		outBuf,
	)

	return nil
}

func (h *Handler) OnCloseByController(eventID string) {
	log.Printf("OnCloseByController")

	pubsub, _ := h.relayService.GetPubsub(eventID)

	pubsub.GetPub().Close()

	for _, sub := range pubsub.subs {
		sub.Close()
	}
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
