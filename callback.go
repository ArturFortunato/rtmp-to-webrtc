package main

import (
	"log"
	"time"
	
	"github.com/pion/webrtc/v2/pkg/media"
	flvtag "github.com/yutopp/go-flv/tag"
)

func onEventCallback(sub *Sub) func(flv *flvtag.FlvTag, msg []byte) error {

	return func(flv *flvtag.FlvTag, msg []byte) error {

		switch flv.Data.(type) {
			case *flvtag.AudioData:
				err := sub.audioTrack.WriteSample(media.Sample{
					Data:    msg,
					Samples: media.NSamples(20*time.Millisecond, 48000),
				})
				
				if err != nil {
					return err
				}

			case *flvtag.VideoData:
				err := sub.videoTrack.WriteSample(media.Sample{
					Data:    msg,
					Samples: media.NSamples(time.Second/30, 90000),
				})

				if err != nil {
					log.Println("[ARTUR] ERROR WRITING VIDEO")
		 			return err
				}
				
			default:
				log.Println("ENTERED DEFAULT ON CALLBACK")

		}

		return nil
	}
}
