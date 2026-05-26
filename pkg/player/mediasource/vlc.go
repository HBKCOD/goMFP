package mediasource

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type VlcPlayer struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	state    PlayerState
	endpoint string
	password string
	cancel   context.CancelFunc
	client   *http.Client
}

type vlcStatus struct {
	State      string  `json:"state"`
	Time       float64 `json:"time"`      // in seconds
	Length     float64 `json:"length"`    // in seconds
	Rate       float64 `json:"rate"`      // playback speed
	Information struct {
		Category struct {
			Meta struct {
				Filename string `json:"filename"`
			} `json:"meta"`
		} `json:"category"`
	} `json:"information"`
}

func NewVlcPlayer() *VlcPlayer {
	return &VlcPlayer{
		status:   StatusDisconnected,
		password: "", // Default empty or configure via UI
		client:   &http.Client{Timeout: 500 * time.Millisecond},
		state: PlayerState{
			Speed: 1.0,
		},
	}
}

func (p *VlcPlayer) Name() string {
	return "vlc"
}

func (p *VlcPlayer) Status() ConnectionStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *VlcPlayer) State() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *VlcPlayer) Connect(ctx context.Context, endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusConnected {
		return nil
	}

	// VLC endpoint can contain password like "localhost:8080|mypassword"
	parts := strings.Split(endpoint, "|")
	p.endpoint = parts[0]
	if len(parts) > 1 {
		p.password = parts[1]
	} else {
		p.password = ""
	}

	p.status = StatusConnecting

	runCtx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go p.pollLoop(runCtx)
	return nil
}

func (p *VlcPlayer) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusDisconnected {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}
	p.status = StatusDisconnected
	p.state.Playing = false
	return nil
}

func (p *VlcPlayer) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *VlcPlayer) poll() {
	p.mu.RLock()
	endpoint := p.endpoint
	password := p.password
	p.mu.RUnlock()

	urlStr := fmt.Sprintf("http://%s/requests/status.json", endpoint)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		p.setDisconnected()
		return
	}

	// VLC uses basic auth with empty username
	auth := ":" + password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	resp, err := p.client.Do(req)
	if err != nil {
		p.setDisconnected()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.setDisconnected()
		return
	}

	var status vlcStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = StatusConnected
	p.state.LastSeen = time.Now()

	p.state.Playing = (status.State == "playing")
	p.state.Position = status.Time
	p.state.Duration = status.Length
	p.state.Speed = status.Rate
	p.state.Path = status.Information.Category.Meta.Filename
}

func (p *VlcPlayer) setDisconnected() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatusConnected {
		p.status = StatusConnecting
	}
	p.state.Playing = false
}

func (p *VlcPlayer) sendCommand(cmd string, val string) error {
	p.mu.RLock()
	endpoint := p.endpoint
	password := p.password
	p.mu.RUnlock()

	urlStr := fmt.Sprintf("http://%s/requests/status.json?command=%s", endpoint, cmd)
	if val != "" {
		urlStr += "&val=" + url.QueryEscape(val)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	auth := ":" + password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (p *VlcPlayer) SetPlayPause(playing bool) error {
	cmd := "pl_forcepause"
	if playing {
		cmd = "pl_play"
	}
	return p.sendCommand(cmd, "")
}

func (p *VlcPlayer) Seek(position float64) error {
	// VLC seek command takes seconds as val
	posInt := int(position)
	return p.sendCommand("seek", fmt.Sprintf("%d", posInt))
}

func (p *VlcPlayer) SetSpeed(speed float64) error {
	return p.sendCommand("rate", fmt.Sprintf("%.2f", speed))
}
