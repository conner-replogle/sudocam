package websocket

import (
	"encoding/json"
	"log/slog"
	pb "messages/msgspb"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)
type ThreadSafeWriter struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func newThreadSafeWriter(conn *websocket.Conn) *ThreadSafeWriter {
	return &ThreadSafeWriter{
		conn: conn,
	}
}

func (w *ThreadSafeWriter) writeMessage(messageType int, data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.conn.WriteMessage(messageType, data)
}









type WebsocketManager struct {
	Writer *ThreadSafeWriter
	conn   *websocket.Conn
	CameraUuid string
	ServerUrl *url.URL
	WSServerURL *url.URL
}


func NewWebsocketManager(u *url.URL, camera_uuid string) *WebsocketManager {

	websocketUrl,_ := u.Parse(u.String())
	websocketUrl.Scheme = "ws"
	if (websocketUrl.Scheme == "https"){
		websocketUrl.Scheme = "wss"
	}
	websocketUrl.Path = "/ws"
	

	slog.Info("connecting to %s", u)

	c, _, err := websocket.DefaultDialer.Dial(websocketUrl.String(), nil)
	if err != nil {
		slog.Error("dial:", err)
		return nil
	}

	writer := newThreadSafeWriter(c)

	SendProtoMessage(writer, &pb.Message{
		From: camera_uuid,
		To:   "server",
		DataType: &pb.Message_Initalization{
			Initalization: &pb.Initalization{
				Id: camera_uuid,
			},
		},
	})

	return &WebsocketManager{
		conn: c,
		Writer: writer,
		CameraUuid: camera_uuid,
		ServerUrl: u,
		WSServerURL: websocketUrl,
	}
}

func (manager *WebsocketManager)Close() {
	manager.conn.Close()
}
func (manager *WebsocketManager)ReadMessage() (*pb.Message, error) {
	_, data,err := manager.conn.ReadMessage()

	if err != nil {
		return nil, err
	}
	msg := &pb.Message{}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
	
}


func (manager *WebsocketManager)SendWebRTCMessage( payload any, to string) error {
	message := &pb.Message{
		From: manager.CameraUuid,
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
	return SendProtoMessage(manager.Writer, message)
}

func SendProtoMessage(writer *ThreadSafeWriter, message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return writer.writeMessage(websocket.BinaryMessage, data)
}