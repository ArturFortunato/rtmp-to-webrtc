package main

import (
	"bytes"
	"sync"
	"log"

	"github.com/pion/webrtc/v2"
	flvtag "github.com/yutopp/go-flv/tag"
)

type Pubsub struct {
	srv  *RelayService
	name string

	pub  *Pub
	subs []*Sub

	m sync.Mutex
}

func NewPubsub(srv *RelayService, name string) *Pubsub {
	log.Println("NEW PUBSUB")
	return &Pubsub{
		srv:  srv,
		name: name,

		subs: make([]*Sub, 0),
	}
}

func (pb *Pubsub) Deregister() error {
	log.Println("DEREGISTER")

	pb.m.Lock()
	defer pb.m.Unlock()

	for _, sub := range pb.subs {
		_ = sub.Close()
	}
	log.Println("DEREGISTER END")

	return pb.srv.RemovePubsub(pb.name)
}

func (pb *Pubsub) Pub() *Pub {
	log.Println("NEW PUB INIT")

	pub := &Pub{
		pb: pb,
	}

	pb.pub = pub
	log.Println("NEW PUB END")

	return pub
}

func (pb *Pubsub) Sub(audioTrack, videoTrack *webrtc.Track) *Sub {
	log.Println("NEW SUBSCRIBER: event ", pb.name)

	pb.m.Lock()
	defer pb.m.Unlock()

	sub := &Sub{
		audioTrack: audioTrack,
		videoTrack: videoTrack,
	}

	pb.subs = append(pb.subs, sub)
	log.Println("NEW SUB END")

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

func (p *Pub) Close() error {
	log.Println("CALLED CLOSE")

	return p.pb.Deregister()
}

type Sub struct {
	initialized bool
	closed      bool

	lastTimestamp uint32
	eventCallback func(*flvtag.FlvTag, []byte) error

	audioTrack *webrtc.Track
	videoTrack *webrtc.Track
}

func (s *Sub) onEvent(flv *flvtag.FlvTag, content []byte) error {

	if s.closed {
		log.Println("WTFF????")
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

func cloneView(flv *flvtag.FlvTag) *flvtag.FlvTag {

	v := *flv

	switch flv.Data.(type) {
		case *flvtag.AudioData:
			dCloned := *v.Data.(*flvtag.AudioData)
			v.Data = &dCloned

			dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())

		case *flvtag.VideoData:
			dCloned := *v.Data.(*flvtag.VideoData)
			v.Data = &dCloned

			dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())
		
		case *flvtag.ScriptData:
			dCloned := *v.Data.(*flvtag.ScriptData)
			v.Data = &dCloned

		default:
			panic("unreachable")
	}
	return &v
}