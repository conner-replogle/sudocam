// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// play-from-disk demonstrates how to send video and/or audio to your browser from files saved to disk.
package main

import (
	"camera/stream"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	pb "messages/msgspb"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
	"google.golang.org/protobuf/proto"
)

const (
	oggPageDuration   = time.Millisecond * 20
	h264FrameDuration = time.Millisecond * 33
)

var addr = flag.String("addr", "100.117.177.44:8080", "http service address")

const ClientUuid = "camera"

type threadSafeWriter struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func newThreadSafeWriter(conn *websocket.Conn) *threadSafeWriter {
	return &threadSafeWriter{
		conn: conn,
	}
}

func (w *threadSafeWriter) writeMessage(messageType int, data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

func main() { //nolint
	// Assert that we have an audio or video file
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	writer := newThreadSafeWriter(c)

	sendProtoMessage(writer, &pb.Message{
		From: ClientUuid,
		To:   "server",
		DataType: &pb.Message_Initalization{
			Initalization: &pb.Initalization{
				Id: ClientUuid,
			},
		},
	})

	// Create a video track
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	go func() {
		// Open a H264 file and start reading using our IVFReader
		r, w := io.Pipe()
		stream.Video(stream.CameraOptions{
			Width:        1080,
			Height:       1920,
			Fps:          30,
			UseLibcamera: true,
			AutoFocus:    true,
		}, w)

		h264, h264Err := h264reader.NewReader(r)
		if h264Err != nil {
			panic(h264Err)
		}

		// Wait for connection established

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		ticker := time.NewTicker(h264FrameDuration)
		fmt.Println("Reading h264")
		for ; true; <-ticker.C {
			nal, h264Err := h264.NextNAL()
			if errors.Is(h264Err, io.EOF) {
				fmt.Printf("All video frames parsed and sent")
				os.Exit(0)
			}
			if h264Err != nil {
				panic(h264Err)
			}

			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: h264FrameDuration}); h264Err != nil {
				panic(h264Err)
			}
		}

	}()

	connections := make(map[string]*webrtc.PeerConnection)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		newMessage := &pb.Message{}
		if err := proto.Unmarshal(data, newMessage); err != nil {
			panic(err)
		}

		webRtc := newMessage.GetWebrtc()
		var (
			candidate webrtc.ICECandidateInit
			offer     webrtc.SessionDescription
		)
		switch {
		// Attempt to unmarshal as a SessionDescription. If the SDP field is empty
		// assume it is not one.
		case json.Unmarshal([]byte(webRtc.Data), &offer) == nil && offer.SDP != "":
			slog.Info("Recieved Offer")
			peerConnection := createPeerConnection(videoTrack, writer, newMessage.From)
			peerConnection.SetRemoteDescription(offer)

			answer, answerErr := peerConnection.CreateAnswer(nil)
			if answerErr != nil {
				panic(answerErr)
			}

			if err = peerConnection.SetLocalDescription(answer); err != nil {
				panic(err)
			}

			if err = sendWebRTCMessage(writer, answer, newMessage.From); err != nil {
				panic(err)
			}
			slog.Debug("Set PeerConnection")
			connections[newMessage.From] = peerConnection

		// Attempt to unmarshal as a ICECandidateInit. If the candidate field is empty
		// assume it is not one.
		case json.Unmarshal([]byte(webRtc.Data), &candidate) == nil && candidate.Candidate != "":
			peerConnection := connections[newMessage.From]
			if peerConnection == nil {
				panic("Recieved ICE Candicate but no pc")
			}
			if err = peerConnection.AddICECandidate(candidate); err != nil {
				panic(err)
			}
		default:
			panic("Unknown message")
		}
	}
}
func createPeerConnection(videoTrack webrtc.TrackLocal, writer *threadSafeWriter, to string) *webrtc.PeerConnection {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	// When Pion gathers a new ICE Candidate send it to the client. This is how
	// ice trickle is implemented. Everytime we have a new candidate available we send
	// it as soon as it is ready. We don't wait to emit a Offer/Answer until they are
	// all available
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		sendWebRTCMessage(writer, candidate.ToJSON(), to)
		if err != nil {
			panic(err)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})
	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	return peerConnection
}

func sendWebRTCMessage(writer *threadSafeWriter, payload any, to string) error {
	message := &pb.Message{
		From: ClientUuid,
		To:   to,
		DataType: &pb.Message_Webrtc{
			Webrtc: &pb.Webrtc{
				Data: func() string {
					data, err := json.Marshal(payload)
					if err != nil {
						panic(err)
					}
					return string(data)
				}(),
			},
		},
	}
	return sendProtoMessage(writer, message)
}

func sendProtoMessage(writer *threadSafeWriter, message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return writer.writeMessage(websocket.BinaryMessage, data)
}
