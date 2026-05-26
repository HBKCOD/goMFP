package mediasource

import (
	"context"
	"sync"
	"time"
)

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusDisconnected ConnectionStatus = "disconnected"
)

// PlayerState represents the current state of a video player
type PlayerState struct {
	Path     string    `json:"path"`
	Duration float64   `json:"duration"` // in seconds
	Position float64   `json:"position"` // in seconds
	Speed    float64   `json:"speed"`
	Playing  bool      `json:"playing"`
	LastSeen time.Time `json:"last_seen"`
}

// MediaSource interface defines all methods needed to interact with a video player
type MediaSource interface {
	Name() string
	Status() ConnectionStatus
	State() PlayerState
	Connect(ctx context.Context, endpoint string) error
	Disconnect() error
	SetPlayPause(playing bool) error
	Seek(position float64) error
	SetSpeed(speed float64) error
}

// InternalPlayer is a local simulated player when no external player is connected
type InternalPlayer struct {
	mu     sync.RWMutex
	status ConnectionStatus
	state  PlayerState
	ticker *time.Ticker
	stop   chan struct{}
}

func NewInternalPlayer() *InternalPlayer {
	return &InternalPlayer{
		status: StatusDisconnected,
		state: PlayerState{
			Path:     "Internal Playlist",
			Duration: 600, // 10 minutes default
			Position: 0,
			Speed:    1.0,
			Playing:  false,
		},
		stop: make(chan struct{}),
	}
}

func (ip *InternalPlayer) Name() string {
	return "internal"
}

func (ip *InternalPlayer) Status() ConnectionStatus {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.status
}

func (ip *InternalPlayer) State() PlayerState {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.state
}

func (ip *InternalPlayer) Connect(ctx context.Context, endpoint string) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.status == StatusConnected {
		return nil
	}

	ip.status = StatusConnected
	ip.state.LastSeen = time.Now()
	ip.stop = make(chan struct{})

	// Start playback simulation ticker
	ip.ticker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ip.ticker.C:
				ip.mu.Lock()
				if ip.state.Playing {
					ip.state.Position += 0.1 * ip.state.Speed
					if ip.state.Position >= ip.state.Duration {
						ip.state.Position = 0 // Loop
					}
					ip.state.LastSeen = time.Now()
				}
				ip.mu.Unlock()
			case <-ip.stop:
				return
			}
		}
	}()

	return nil
}

func (ip *InternalPlayer) Disconnect() error {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.status == StatusDisconnected {
		return nil
	}

	ip.status = StatusDisconnected
	if ip.ticker != nil {
		ip.ticker.Stop()
		close(ip.stop)
	}
	return nil
}

func (ip *InternalPlayer) SetPlayPause(playing bool) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.state.Playing = playing
	ip.state.LastSeen = time.Now()
	return nil
}

func (ip *InternalPlayer) Seek(position float64) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	if position < 0 {
		position = 0
	}
	if position > ip.state.Duration {
		position = ip.state.Duration
	}
	ip.state.Position = position
	ip.state.LastSeen = time.Now()
	return nil
}

func (ip *InternalPlayer) SetSpeed(speed float64) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	if speed <= 0 {
		speed = 1.0
	}
	ip.state.Speed = speed
	ip.state.LastSeen = time.Now()
	return nil
}
