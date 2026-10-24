package ws

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SafeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewSafeConn(conn *websocket.Conn) *SafeConn {
	conn.SetReadLimit(4 << 20)
	return &SafeConn{
		conn: conn,
		mu:   sync.Mutex{},
	}
}

func (sc *SafeConn) WriteMessage(messageType int, data []byte) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteMessage(messageType, data)
}

func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteJSON(v)
}

func (sc *SafeConn) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.Close()
}
func (sc *SafeConn) ReadMessage() (int, []byte, error) {
	// sc.mu.Lock()
	// defer sc.mu.Unlock()
	typeID, reader, err := sc.conn.NextReader()
	if err != nil {
		return 0, nil, err
	}
	data, err := io.ReadAll(io.LimitReader(reader, (4<<20)+1))
	if err == nil && len(data) > 4<<20 {
		_ = sc.conn.Close()
		return 0, nil, errors.New("monitor message too large")
	}
	return typeID, data, err
}
func (sc *SafeConn) ReadJSON(v interface{}) error {
	// sc.mu.Lock()
	// defer sc.mu.Unlock()
	_, data, err := sc.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func (sc *SafeConn) SetReadDeadline(t time.Time) error {
	// sc.mu.Lock()
	// defer sc.mu.Unlock()
	return sc.conn.SetReadDeadline(t)
}
func (sc *SafeConn) GetConn() *websocket.Conn {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn
}
