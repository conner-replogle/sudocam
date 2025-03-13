package webrtc

import (
	"camera/stream"
	"camera/websocket"
	"encoding/json"
	"errors"
	"fmt"
	
	"log/slog"
	pb "messages/msgspb"
	"time"

	"github.com/pion/webrtc/v4"
)

const (
	h264FrameDuration = time.Millisecond * 33
)

type WebRTCManager struct {
	Websocket   *websocket.WebsocketManager
	connections map[string]*webrtc.PeerConnection
	videoTrack  webrtc.TrackLocal
}

func NewWebRTCManager(ws *websocket.WebsocketManager) *WebRTCManager {
	return &WebRTCManager{
		Websocket:   ws,
		connections: make(map[string]*webrtc.PeerConnection),
	}
}

func (manager *WebRTCManager) StartCamera() {
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "sudocam")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	stream.Video(videoTrack)
	manager.videoTrack = videoTrack
}

func (manager *WebRTCManager) CreatePeerConnection(client_uuid string) *webrtc.PeerConnection {
	// Fetch TURN credentials
	creds, err := fetchTURNCredentials(manager.Websocket.ServerUrl.String())
	if err != nil {
		slog.Error("Failed to fetch TURN credentials", "error", err)
		// Continue with default config if TURN fails
		creds = nil
	}

	config := webrtc.Configuration{}
	if creds != nil {
		var iceServers []webrtc.ICEServer
		iceServers = append(iceServers, *creds)
		config.ICEServers = iceServers
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
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

		err = manager.Websocket.SendWebRTCMessage(candidate.ToJSON(), client_uuid)
		slog.Debug("Sent Ice Candidate")

		if err != nil {
			panic(err)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})
	rtpSender, videoTrackErr := peerConnection.AddTrack(manager.videoTrack)
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

func (manager *WebRTCManager) HandleMessage(msg *pb.Webrtc, from string) error {

	if from == "" {
		return errors.New("no from field in message")
	}
	var (
		candidate webrtc.ICECandidateInit
		offer     webrtc.SessionDescription
	)


	switch {
	// Attempt to unmarshal as a SessionDescription. If the SDP field is empty
	// assume it is not one.
	case json.Unmarshal([]byte(msg.Data), &offer) == nil && offer.SDP != "":
		slog.Info("Recieved Offer")
		peerConnection := manager.CreatePeerConnection(from)
		peerConnection.SetRemoteDescription(offer)

		answer, answerErr := peerConnection.CreateAnswer(nil)
		if answerErr != nil {
			return answerErr
		}

		if err := peerConnection.SetLocalDescription(answer); err != nil {
			return err
		}

		if err := manager.Websocket.SendWebRTCMessage(answer, from); err != nil {
			return err
		}
		slog.Debug("Sent Answer")

		manager.connections[from] = peerConnection

	// Attempt to unmarshal as a ICECandidateInit. If the candidate field is empty
	// assume it is not one.
	case json.Unmarshal([]byte(msg.Data), &candidate) == nil && candidate.Candidate != "":
		slog.Debug("Recieved ICE Candidate")
		peerConnection := manager.connections[from]
		if peerConnection == nil {
			panic("Recieved ICE Candicate but no pc")
		}
		if err := peerConnection.AddICECandidate(candidate); err != nil {
			return err
		}
	default:
		slog.Error("Unknown message type", "msg", msg)

	}
	return nil
}
