package mediasource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MpcHcPlayer struct {
	mu       sync.RWMutex
	status   ConnectionStatus
	state    PlayerState
	endpoint string
	cancel   context.CancelFunc
	client   *http.Client
}

func NewMpcHcPlayer() *MpcHcPlayer {
	return &MpcHcPlayer{
		status: StatusDisconnected,
		client: &http.Client{Timeout: 500 * time.Millisecond},
		state: PlayerState{
			Speed: 1.0,
		},
	}
}

func (p *MpcHcPlayer) Name() string {
	return "mpchc"
}

func (p *MpcHcPlayer) Status() ConnectionStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *MpcHcPlayer) State() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *MpcHcPlayer) Connect(ctx context.Context, endpoint string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusConnected {
		return nil
	}

	p.endpoint = endpoint
	p.status = StatusConnecting

	runCtx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go p.pollLoop(runCtx)
	return nil
}

func (p *MpcHcPlayer) Disconnect() error {
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

func (p *MpcHcPlayer) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Regex for variables.html parsing
	// e.g. <p id="filepath">C:\Video.mp4</p>
	regex := regexp.MustCompile(`<p id="(?P<name>[^"]+)">(?P<value>.*?)</p>`)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(regex)
		}
	}
}

func (p *MpcHcPlayer) poll(regex *regexp.Regexp) {
	urlStr := fmt.Sprintf("http://%s/variables.html", p.endpoint)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		p.setDisconnected()
		return
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = StatusConnected
	p.state.LastSeen = time.Now()

	matches := regex.FindAllStringSubmatch(string(body), -1)
	vars := make(map[string]string)
	for _, m := range matches {
		if len(m) >= 3 {
			vars[m[1]] = m[2]
		}
	}

	// State mapping: 0 = stopped, 1 = paused, 2 = playing
	if stateStr, ok := vars["state"]; ok {
		if stateVal, err := strconv.Atoi(stateStr); err == nil {
			p.state.Playing = (stateVal == 2)
		}
	}

	if filepath, ok := vars["filepath"]; ok && filepath != "" {
		p.state.Path = filepath
	} else {
		// No file loaded
		p.state.Path = ""
		p.state.Playing = false
		p.state.Position = 0
		p.state.Duration = 0
	}

	if durationStr, ok := vars["duration"]; ok {
		if durationVal, err := strconv.ParseFloat(durationStr, 64); err == nil {
			p.state.Duration = durationVal / 1000.0 // ms to seconds
		}
	}

	if positionStr, ok := vars["position"]; ok {
		if positionVal, err := strconv.ParseFloat(positionStr, 64); err == nil {
			p.state.Position = positionVal / 1000.0 // ms to seconds
		}
	}

	if speedStr, ok := vars["playbackrate"]; ok {
		speedStr = strings.ReplaceAll(speedStr, ",", ".")
		if speedVal, err := strconv.ParseFloat(speedStr, 64); err == nil && speedVal > 0 {
			p.state.Speed = speedVal
		}
	}
}

func (p *MpcHcPlayer) setDisconnected() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatusConnected {
		p.status = StatusConnecting
	}
	p.state.Playing = false
}

func (p *MpcHcPlayer) sendCommand(commandID int, extraParams url.Values) error {
	p.mu.RLock()
	endpoint := p.endpoint
	p.mu.RUnlock()

	params := url.Values{}
	params.Set("wm_command", strconv.Itoa(commandID))
	for k, v := range extraParams {
		for _, val := range v {
			params.Add(k, val)
		}
	}

	urlStr := fmt.Sprintf("http://%s/command.html?%s", endpoint, params.Encode())
	resp, err := p.client.Get(urlStr)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (p *MpcHcPlayer) SetPlayPause(playing bool) error {
	cmd := 888 // Pause
	if playing {
		cmd = 887 // Play
	}
	return p.sendCommand(cmd, nil)
}

func (p *MpcHcPlayer) Seek(position float64) error {
	// position in seconds => convert to HH:MM:SS
	t := time.Duration(position * float64(time.Second))
	h := t / time.Hour
	t -= h * time.Hour
	m := t / time.Minute
	t -= m * time.Minute
	s := t / time.Second

	timeStr := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	params := url.Values{}
	params.Set("position", timeStr)

	// wm_command = -1 for seek in MPC-HC web API
	return p.sendCommand(-1, params)
}

func (p *MpcHcPlayer) SetSpeed(speed float64) error {
	// MPC-HC doesn't support setting arbitrary speed via web API WM_COMMAND directly easily,
	// but let's ignore or log it. It's not a dealbreaker.
	return nil
}
