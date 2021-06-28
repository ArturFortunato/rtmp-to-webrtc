package main

import (
	"log"
	"sync"

	"github.com/pion/webrtc/v3"
	flvtag "github.com/yutopp/go-flv/tag"
)

type Pubsub struct {
	//srv  *RelayService
	name string

	pub  *Pub
	subs []*Sub

	m sync.Mutex
}

func NewPubsub(name string) *Pubsub {
	return &Pubsub{
		name: name,

		subs: make([]*Sub, 0),
	}
}

func (pb *Pubsub) addNewClient() {
	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion")
	if err != nil {
		panic(err)
	}

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "video", "pion")
	if err != nil {
		panic(err)
	}

	pb.Sub(audioTrack, videoTrack)
}

// func (pb *Pubsub) Deregister() error {
// 	pb.m.Lock()
// 	defer pb.m.Unlock()

// 	for _, sub := range pb.subs {
// 		_ = sub.Close()
// 	}

// 	return pb.srv.RemovePubsub(pb.name)
// }

func (pb *Pubsub) GetPub() *Pub {
	return pb.pub
}

func (pb *Pubsub) Pub() *Pub {

	pub := &Pub{
		pb: pb,
	}

	pb.pub = pub

	return pub
}

func (pb *Pubsub) Sub(audioTrack, videoTrack *webrtc.TrackLocalStaticRTP) *Sub {
	log.Println("NEW SUBSCRIBER: event ", pb.name)

	pb.m.Lock()
	defer pb.m.Unlock()

	sub := &Sub{
		audioTrack: audioTrack,
		videoTrack: videoTrack,
	}

	pb.subs = append(pb.subs, sub)

	return sub
}

type Pub struct {
	pb *Pubsub

	avcSeqHeader *flvtag.FlvTag
	lastKeyFrame *flvtag.FlvTag
}

func (p *Pub) Publish(flv *flvtag.FlvTag, content []byte) error {
	switch flv.Data.(type) {
	case *flvtag.AudioData:
		for _, sub := range p.pb.subs {
			data := content

			_ = sub.onEvent(flv, data)
		}

	case *flvtag.VideoData:
		for _, sub := range p.pb.subs {
			data := content

			_ = sub.onEvent(flv, data)
		}

	case *flvtag.ScriptData:
		log.Println("Received ScriptData")

	default:
		panic("unexpected")
	}

	return nil
}

// func (p *Pub) Close() error {
// 	log.Println("CALLED CLOSE")

// 	return p.pb.Deregister()
// }

type Sub struct {
	initialized bool
	closed      bool

	lastTimestamp uint32
	eventCallback func(*flvtag.FlvTag, []byte) error

	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP
}

func (s *Sub) onEvent(flv *flvtag.FlvTag, content []byte) error {

	if s.closed {
		return nil
	}

	return s.eventCallback(flv, content)
}

func (s *Sub) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true

	return nil
}
