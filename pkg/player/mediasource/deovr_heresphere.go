package mediasource

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type DeoVrHereSpherePlayer struct {
	mu         sync.RWMutex
	status     ConnectionStatus
	state      PlayerState
	endpoint   string
	playerName string // "deovr" or "heresphere"
	conn       net.Conn
	cancel     context.CancelFunc
}

type tcpJsonState struct {
	Path          string   `json:"path,omitempty"`
	Resource      string   `json:"resource,omitempty"` // Used by HereSphere
	CurrentTime   *float64 `json:"currentTime,omitempty"`
	PlaybackSpeed *float64 `json:"playbackSpeed,omitempty"`
	PlayerState   *int     `json:"playerState,omitempty"` // 0 = playing, 1 = paused
	Duration      *float64 `json:"duration,omitempty"`
}

func NewDeoVrHereSpherePlayer(name string) *DeoVrHereSpherePlayer {
	return &DeoVrHereSpherePlayer{
		status:     StatusDisconnected,
		playerName: name,
		state: PlayerState{
			Speed: 1.0,
		},
	}
}

func (p *DeoVrHereSpherePlayer) Name() string {
	return p.playerName
}

func (p *DeoVrHereSpherePlayer) Status() ConnectionStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *DeoVrHereSpherePlayer) State() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *DeoVrHereSpherePlayer) Connect(ctx context.Context, endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusConnected {
		return nil
	}

	p.endpoint = endpoint
	p.status = StatusConnecting

	runCtx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go p.connectionLoop(runCtx)
	return nil
}

func (p *DeoVrHereSpherePlayer) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusDisconnected {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	p.status = StatusDisconnected
	p.state.Playing = false
	return nil
}

func (p *DeoVrHereSpherePlayer) connectionLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		p.mu.RLock()
		endpoint := p.endpoint
		p.mu.RUnlock()

		d := net.Dialer{Timeout: 1 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", endpoint)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		p.mu.Lock()
		p.conn = conn
		p.status = StatusConnected
		p.state.LastSeen = time.Now()
		p.mu.Unlock()

		// Start read and write loops
		errChan := make(chan error, 2)
		readCtx, readCancel := context.WithCancel(ctx)

		go p.readLoop(readCtx, conn, errChan)
		go p.writeLoop(readCtx, conn, errChan)

		// Wait for error in connection
		select {
		case <-ctx.Done():
			readCancel()
			conn.Close()
			return
		case <-errChan:
			readCancel()
			conn.Close()
			p.mu.Lock()
			p.status = StatusConnecting
			p.state.Playing = false
			p.mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}
}

func (p *DeoVrHereSpherePlayer) readLoop(ctx context.Context, conn net.Conn, errChan chan error) {
	lenBuf := make([]byte, 4)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read 4-byte length prefix
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			errChan <- err
			return
		}

		length := binary.LittleEndian.Uint32(lenBuf)
		if length == 0 {
			// Keep alive or empty packet, skip
			continue
		}

		if length > 65536 {
			// Protection against corrupted length prefix
			errChan <- errors.New("packet size exceeds limit")
			return
		}

		dataBuf := make([]byte, length)
		_, err = io.ReadFull(conn, dataBuf)
		if err != nil {
			errChan <- err
			return
		}

		var payload tcpJsonState
		if err := json.Unmarshal(dataBuf, &payload); err != nil {
			continue // Non-fatal, just continue
		}

		p.mu.Lock()
		p.state.LastSeen = time.Now()

		if payload.Path != "" {
			p.state.Path = payload.Path
		} else if payload.Resource != "" {
			p.state.Path = payload.Resource
		}

		if payload.CurrentTime != nil {
			p.state.Position = *payload.CurrentTime
		}
		if payload.Duration != nil {
			p.state.Duration = *payload.Duration
		}
		if payload.PlaybackSpeed != nil {
			p.state.Speed = *payload.PlaybackSpeed
		}
		if payload.PlayerState != nil {
			p.state.Playing = (*payload.PlayerState == 0) // 0 is playing in DeoVR/HereSphere
		}
		p.mu.Unlock()
	}
}

func (p *DeoVrHereSpherePlayer) writeLoop(ctx context.Context, conn net.Conn, errChan chan error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	keepAlive := make([]byte, 4) // 4 bytes of 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send keep-alive
			_, err := conn.Write(keepAlive)
			if err != nil {
				errChan <- err
				return
			}
		}
	}
}

func (p *DeoVrHereSpherePlayer) sendState(state tcpJsonState) error {
	p.mu.Lock()
	conn := p.conn
	p.mu.Unlock()

	if conn == nil {
		return errors.New("not connected")
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	buf := make([]byte, 4+length)
	binary.LittleEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], data)

	_, err = conn.Write(buf)
	return err
}

func (p *DeoVrHereSpherePlayer) SetPlayPause(playing bool) error {
	stateVal := 1 // paused
	if playing {
		stateVal = 0 // playing
	}
	return p.sendState(tcpJsonState{
		PlayerState: &stateVal,
	})
}

func (p *DeoVrHereSpherePlayer) Seek(position float64) error {
	return p.sendState(tcpJsonState{
		CurrentTime: &position,
	})
}

func (p *DeoVrHereSpherePlayer) SetSpeed(speed float64) error {
	return p.sendState(tcpJsonState{
		PlaybackSpeed: &speed,
	})
}
