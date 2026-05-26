package mediasource

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
)

type MpvPlayer struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	state    PlayerState
	endpoint string // Pipe name (e.g. "multifunplayer-mpv")
	conn     net.Conn
	cancel   context.CancelFunc
}

type mpvEvent struct {
	Event string      `json:"event"`
	Name  string      `json:"name"`
	Data  interface{} `json:"data"`
	Id    int         `json:"id"`
}

type mpvCommand struct {
	Command []interface{} `json:"command"`
}

func NewMpvPlayer() *MpvPlayer {
	return &MpvPlayer{
		status: StatusDisconnected,
		state: PlayerState{
			Speed: 1.0,
		},
	}
}

func (p *MpvPlayer) Name() string {
	return "mpv"
}

func (p *MpvPlayer) Status() ConnectionStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *MpvPlayer) State() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *MpvPlayer) Connect(ctx context.Context, endpoint string) error {
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

func (p *MpvPlayer) Disconnect() error {
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

func (p *MpvPlayer) connectionLoop(ctx context.Context) {
	pipePath := ""

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		p.mu.RLock()
		endpoint := p.endpoint
		p.mu.RUnlock()

		// Build named pipe path on Windows
		if !strings.HasPrefix(endpoint, `\\`) {
			pipePath = `\\.\pipe\` + endpoint
		} else {
			pipePath = endpoint
		}

		// Dial Windows named pipe
		conn, err := winio.DialPipe(pipePath, nil)
		if err != nil {
			// Wait and retry
			time.Sleep(1 * time.Second)
			continue
		}

		p.mu.Lock()
		p.conn = conn
		p.status = StatusConnected
		p.state.LastSeen = time.Now()
		p.mu.Unlock()

		// Initialize property observation
		err = p.observeProperties()
		if err != nil {
			conn.Close()
			p.mu.Lock()
			p.status = StatusConnecting
			p.mu.Unlock()
			continue
		}

		// Read responses
		errChan := make(chan error, 1)
		go p.readLoop(ctx, conn, errChan)

		select {
		case <-ctx.Done():
			conn.Close()
			return
		case <-errChan:
			conn.Close()
			p.mu.Lock()
			p.status = StatusConnecting
			p.state.Playing = false
			p.mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}
}

func (p *MpvPlayer) observeProperties() error {
	commands := []mpvCommand{
		{Command: []interface{}{"observe_property", 1, "pause"}},
		{Command: []interface{}{"observe_property", 2, "duration"}},
		{Command: []interface{}{"observe_property", 3, "time-pos"}},
		{Command: []interface{}{"observe_property", 4, "path"}},
		{Command: []interface{}{"observe_property", 5, "speed"}},
	}

	for _, cmd := range commands {
		data, err := json.Marshal(cmd)
		if err != nil {
			return err
		}
		p.mu.Lock()
		conn := p.conn
		p.mu.Unlock()
		if conn == nil {
			return errors.New("conn is nil")
		}
		_, err = conn.Write(append(data, '\n'))
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *MpvPlayer) readLoop(ctx context.Context, conn net.Conn, errChan chan error) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev mpvEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}

		if ev.Event == "property-change" {
			p.mu.Lock()
			p.state.LastSeen = time.Now()

			switch ev.Name {
			case "pause":
				if val, ok := ev.Data.(bool); ok {
					p.state.Playing = !val
				} else if valStr, ok := ev.Data.(string); ok {
					p.state.Playing = (valStr == "no")
				}
			case "duration":
				if val, ok := ev.Data.(float64); ok {
					p.state.Duration = val
				}
			case "time-pos":
				if val, ok := ev.Data.(float64); ok {
					p.state.Position = val
				}
			case "path":
				if val, ok := ev.Data.(string); ok {
					p.state.Path = val
				}
			case "speed":
				if val, ok := ev.Data.(float64); ok {
					p.state.Speed = val
				} else if valStr, ok := ev.Data.(string); ok {
					if f, err := strconv.ParseFloat(valStr, 64); err == nil {
						p.state.Speed = f
					}
				}
			}
			p.mu.Unlock()
		}
	}

	if err := scanner.Err(); err != nil {
		errChan <- err
	} else {
		errChan <- io.EOF
	}
}

func (p *MpvPlayer) writeCommand(cmd []interface{}) error {
	p.mu.Lock()
	conn := p.conn
	p.mu.Unlock()

	if conn == nil {
		return errors.New("not connected")
	}

	payload := mpvCommand{Command: cmd}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = conn.Write(append(data, '\n'))
	return err
}

func (p *MpvPlayer) SetPlayPause(playing bool) error {
	val := "yes"
	if playing {
		val = "no"
	}
	return p.writeCommand([]interface{}{"set_property", "pause", val})
}

func (p *MpvPlayer) Seek(position float64) error {
	return p.writeCommand([]interface{}{"set_property", "time-pos", position})
}

func (p *MpvPlayer) SetSpeed(speed float64) error {
	return p.writeCommand([]interface{}{"set_property", "speed", speed})
}
