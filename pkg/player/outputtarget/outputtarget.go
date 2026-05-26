package outputtarget

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusDisconnected ConnectionStatus = "disconnected"
)

// OutputTarget defines methods for sending TCode commands to devices
type OutputTarget interface {
	Name() string
	Status() ConnectionStatus
	Connect(endpoint string) error
	Disconnect() error
	Send(tcode string) error
}

// ==========================================
// UDP Output
// ==========================================
type UdpOutput struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	endpoint string
	conn     net.Conn
}

func NewUdpOutput() *UdpOutput {
	return &UdpOutput{status: StatusDisconnected}
}

func (o *UdpOutput) Name() string { return "udp" }
func (o *UdpOutput) Status() ConnectionStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *UdpOutput) Connect(endpoint string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	conn, err := net.Dial("udp", endpoint)
	if err != nil {
		return err
	}
	o.conn = conn
	o.endpoint = endpoint
	o.status = StatusConnected
	return nil
}

func (o *UdpOutput) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.conn != nil {
		o.conn.Close()
		o.conn = nil
	}
	o.status = StatusDisconnected
	return nil
}

func (o *UdpOutput) Send(tcode string) error {
	o.mu.Lock()
	conn := o.conn
	o.mu.Unlock()

	if conn == nil {
		return errors.New("udp output not connected")
	}
	_, err := conn.Write([]byte(tcode))
	return err
}

// ==========================================
// TCP Output
// ==========================================
type TcpOutput struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	endpoint string
	conn     net.Conn
	cancel   context.CancelFunc
}

func NewTcpOutput() *TcpOutput {
	return &TcpOutput{status: StatusDisconnected}
}

func (o *TcpOutput) Name() string { return "tcp" }
func (o *TcpOutput) Status() ConnectionStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *TcpOutput) Connect(endpoint string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.endpoint = endpoint
	o.status = StatusConnecting

	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := net.DialTimeout("tcp", endpoint, 2*time.Second)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			o.mu.Lock()
			o.conn = conn
			o.status = StatusConnected
			o.mu.Unlock()

			// Check connection status by reading or waiting
			buf := make([]byte, 1)
			for {
				_, err := conn.Read(buf)
				if err != nil {
					// Disconnected
					break
				}
			}

			o.mu.Lock()
			o.status = StatusConnecting
			o.conn = nil
			o.mu.Unlock()
			conn.Close()
		}
	}()

	return nil
}

func (o *TcpOutput) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel != nil {
		o.cancel()
	}
	if o.conn != nil {
		o.conn.Close()
		o.conn = nil
	}
	o.status = StatusDisconnected
	return nil
}

func (o *TcpOutput) Send(tcode string) error {
	o.mu.Lock()
	conn := o.conn
	o.mu.Unlock()

	if conn == nil {
		return errors.New("tcp output not connected")
	}
	_, err := conn.Write([]byte(tcode))
	return err
}

// ==========================================
// WebSocket Output
// ==========================================
type WebSocketOutput struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	endpoint string
	conn     *websocket.Conn
	cancel   context.CancelFunc
}

func NewWebSocketOutput() *WebSocketOutput {
	return &WebSocketOutput{status: StatusDisconnected}
}

func (o *WebSocketOutput) Name() string { return "websocket" }
func (o *WebSocketOutput) Status() ConnectionStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *WebSocketOutput) Connect(endpoint string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.endpoint = endpoint
	o.status = StatusConnecting

	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			dialer := websocket.DefaultDialer
			dialer.HandshakeTimeout = 2 * time.Second
			conn, _, err := dialer.Dial(endpoint, nil)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			o.mu.Lock()
			o.conn = conn
			o.status = StatusConnected
			o.mu.Unlock()

			// Listen for connection close
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}

			o.mu.Lock()
			o.status = StatusConnecting
			o.conn = nil
			o.mu.Unlock()
			conn.Close()
		}
	}()

	return nil
}

func (o *WebSocketOutput) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel != nil {
		o.cancel()
	}
	if o.conn != nil {
		o.conn.Close()
		o.conn = nil
	}
	o.status = StatusDisconnected
	return nil
}

func (o *WebSocketOutput) Send(tcode string) error {
	o.mu.Lock()
	conn := o.conn
	o.mu.Unlock()

	if conn == nil {
		return errors.New("websocket output not connected")
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(tcode))
}

// ==========================================
// File Output
// ==========================================
type FileOutput struct {
	mu     sync.RWMutex
	status ConnectionStatus
	file   *os.File
}

func NewFileOutput() *FileOutput {
	return &FileOutput{status: StatusDisconnected}
}

func (o *FileOutput) Name() string { return "file" }
func (o *FileOutput) Status() ConnectionStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *FileOutput) Connect(filePath string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	o.file = file
	o.status = StatusConnected
	return nil
}

func (o *FileOutput) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.file != nil {
		o.file.Close()
		o.file = nil
	}
	o.status = StatusDisconnected
	return nil
}

func (o *FileOutput) Send(tcode string) error {
	o.mu.Lock()
	file := o.file
	o.mu.Unlock()

	if file == nil {
		return errors.New("file output not open")
	}
	_, err := file.WriteString(tcode)
	return err
}

// ==========================================
// Serial Output
// ==========================================
type SerialOutput struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	endpoint string // COM Port e.g. "COM3"
	baudRate int
	port     serial.Port
	cancel   context.CancelFunc
}

func NewSerialOutput(baudRate int) *SerialOutput {
	if baudRate <= 0 {
		baudRate = 115200
	}
	return &SerialOutput{
		status:   StatusDisconnected,
		baudRate: baudRate,
	}
}

func (o *SerialOutput) Name() string { return "serial" }
func (o *SerialOutput) Status() ConnectionStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.status
}

func (o *SerialOutput) Connect(endpoint string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.endpoint = endpoint
	o.status = StatusConnecting

	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			mode := &serial.Mode{
				BaudRate: o.baudRate,
			}
			port, err := serial.Open(o.endpoint, mode)
			if err != nil {
				// Retry connection
				time.Sleep(1 * time.Second)
				continue
			}

			// Enable DTR & RTS to support ESP32/Arduino microcontrollers (standard in MultiFunPlayer)
			_ = port.SetDTR(true)
			_ = port.SetRTS(true)

			o.mu.Lock()
			o.port = port
			o.status = StatusConnected
			o.mu.Unlock()

			// Block and wait until disconnected or context cancelled.
			// Eliminates background Read concurrency stalls on Windows.
			<-ctx.Done()

			o.mu.Lock()
			o.port = nil
			o.status = StatusDisconnected
			o.mu.Unlock()
			port.Close()
			return
		}
	}()

	return nil
}

func (o *SerialOutput) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel != nil {
		o.cancel()
	}
	o.status = StatusDisconnected
	return nil
}

func (o *SerialOutput) Send(tcode string) error {
	o.mu.Lock()
	port := o.port
	o.mu.Unlock()

	if port == nil {
		return fmt.Errorf("serial port %s not connected", o.endpoint)
	}

	_, err := port.Write([]byte(tcode))
	if err != nil {
		// Trigger active disconnect on Write failure
		go o.Disconnect()
	}
	return err
}
