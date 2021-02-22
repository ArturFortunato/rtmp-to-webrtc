package main

import (
	//"bytes"
	//"encoding/binary"
	//"context"
	"log"
	//"io"
	"time"
	
	"github.com/pion/webrtc/v2/pkg/media"
	flvtag "github.com/yutopp/go-flv/tag"
	//"github.com/yutopp/go-rtmp"
	//rtmpmsg "github.com/yutopp/go-rtmp/message"
)

func onEventCallback(sub *Sub) func(flv *flvtag.FlvTag, msg []byte) error {

	return func(flv *flvtag.FlvTag, msg []byte) error {

		switch flv.Data.(type) {
			case *flvtag.AudioData:
				// A MENSAGEM TEM DE SER COPIADA ANTES DE VIR PARA AQUI
				/*d := flv.Data.(*flvtag.AudioData)

				data := new(bytes.Buffer)

				if _, err := io.Copy(data, d.Data); err != nil {
					return err
				}*/

				err := sub.audioTrack.WriteSample(media.Sample{
					Data:    msg,
					Samples: media.NSamples(20*time.Millisecond, 48000),
				})
				
				if err != nil {
					return err
				}

			case *flvtag.VideoData:
				/*d := flv.Data.(*flvtag.VideoData)

				data := new(bytes.Buffer)
				if _, err := io.Copy(data, d.Data); err != nil {
					return err
				}*/

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


// func onEventCallback(conn *rtmp.Conn, streamID string) func(flv *flvtag.FlvTag) error {
// 	return func(flv *flvtag.FlvTag) error {
// 		buf := new(bytes.Buffer)

// 		switch flv.Data.(type) {
// 		case *flvtag.AudioData:
// 			d := flv.Data.(*flvtag.AudioData)

// 			// Consume flv payloads (d)
// 			if err := flvtag.EncodeAudioData(buf, d); err != nil {
// 				return err
// 			}

// 			// TODO: Fix these values
// 			ctx := context.Background()

// 			chunkStreamID := 5
// 			return nil
// 			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
// 				StreamID: streamID,
// 				Message: &rtmpmsg.AudioMessage{
// 					Payload: buf,
// 				},
// 			})

// 		case *flvtag.VideoData:
// 			d := flv.Data.(*flvtag.VideoData)

// 			// Consume flv payloads (d)
// 			if err := flvtag.EncodeVideoData(buf, d); err != nil {
// 				return err
// 			}

// 			//TODO: Fix these values
// 			ctx := context.Background()
// 			chunkStreamID := 6
// 			return nil
// 			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
// 				StreamID: streamID,
// 				Message: &rtmpmsg.VideoMessage{
// 					Payload: buf,
// 				},
// 			})

// 		case *flvtag.ScriptData:
// 			d := flv.Data.(*flvtag.ScriptData)

// 			// Consume flv payloads (d)
// 			if err := flvtag.EncodeScriptData(buf, d); err != nil {
// 				return err
// 			}

// 			// TODO: hide these implementation
// 			amdBuf := new(bytes.Buffer)
// 			amfEnc := rtmpmsg.NewAMFEncoder(amdBuf, rtmpmsg.EncodingTypeAMF0)
// 			if err := rtmpmsg.EncodeBodyAnyValues(amfEnc, &rtmpmsg.NetStreamSetDataFrame{
// 				Payload: buf.Bytes(),
// 			}); err != nil {
// 				return err
// 			}

// 			// TODO: Fix these values
// 			ctx := context.Background()
// 			chunkStreamID := 8
// 			return nil
// 			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
// 				StreamID: streamID,
// 				Message: &rtmpmsg.DataMessage{
// 					Name:     "@setDataFrame", // TODO: fix
// 					Encoding: rtmpmsg.EncodingTypeAMF0,
// 					Body:     amdBuf,
// 				},
// 			})

// 		default:
// 			panic("unreachable")
// 		}
// 	}
// }
