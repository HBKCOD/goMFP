package settings

import (
	"encoding/json"
	"os"
	"sync"
)

// AxisSettings defines configuration for a single device axis (e.g. L0)
type AxisSettings struct {
	Name           string   `json:"name"`
	FriendlyName   string   `json:"friendly_name"`
	FunscriptNames []string `json:"funscript_names"`
	Enabled        bool     `json:"enabled"`
	DefaultValue   float64  `json:"default_value"`
	Min            float64  `json:"min"`
	Max            float64  `json:"max"`
	Offset         float64  `json:"offset"`
	Invert         bool     `json:"invert"`
	Interpolation  string   `json:"interpolation"` // "linear", "pchip", "makima", "step"
	AutoHome       bool     `json:"auto_home"`
	AutoHomeDelay  float64  `json:"auto_home_delay"` // seconds of inactivity before returning home
	MotionType     string   `json:"motion_type"`    // "none", "sine", "triangle", "saw", "square", "random"
	MotionSpeed    float64  `json:"motion_speed"`
	MotionRange    float64  `json:"motion_range"`
	MotionOffset   float64  `json:"motion_offset"`
}

// MediaSourceSettings defines player connection profiles
type MediaSourceSettings struct {
	Type        string `json:"type"` // "internal", "mpchc", "mpv", "vlc", "deovr", "heresphere"
	Endpoint    string `json:"endpoint"`
	Password    string `json:"password"`
	AutoConnect bool   `json:"auto_connect"`
}

// OutputTargetSettings defines device connection profiles
type OutputTargetSettings struct {
	Type        string `json:"type"` // "serial", "udp", "tcp", "websocket", "file"
	Endpoint    string `json:"endpoint"`
	BaudRate    int    `json:"baud_rate"` // For serial
	FilePath    string `json:"file_path"` // For file output
	Enabled     bool   `json:"enabled"`
	AutoConnect bool   `json:"auto_connect"`
}

// AppSettings is the root configuration structure
// AppSettings is the root configuration structure
type AppSettings struct {
	UpdateRate        int                             `json:"update_rate"` // e.g. 50 ticks/sec
	RememberLoc       bool                            `json:"remember_window_location"`
	ActivePlayer      string                          `json:"active_player"`
	Axes              map[string]*AxisSettings        `json:"axes"`
	MediaSources      map[string]*MediaSourceSettings `json:"media_sources"`
	OutputTargets     map[string]*OutputTargetSettings `json:"output_targets"`
	ScriptDirectories []string                        `json:"script_directories"`
	GlobalOffset      float64                         `json:"global_offset"`
	UseGlobalOffset   bool                            `json:"use_global_offset"`
	AllowRemoteAccess bool                            `json:"allow_remote_access"`
	WebPort           string                          `json:"web_port"`
	Language          string                          `json:"language"`
}

// SettingsManager coordinates access to settings
type SettingsManager struct {
	path string
	Mu   sync.RWMutex
	Data AppSettings
}

// NewSettingsManager creates a manager pointing to the settings file
func NewSettingsManager(path string) *SettingsManager {
	return &SettingsManager{
		path: path,
		Data: DefaultSettings(),
	}
}

// DefaultSettings returns settings with standard axes and defaults
func DefaultSettings() AppSettings {
	axes := map[string]*AxisSettings{
		"L0": {Name: "L0", FriendlyName: "Up/Down", FunscriptNames: []string{"*", "stroke", "L0", "up"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"L1": {Name: "L1", FriendlyName: "Forward/Backward", FunscriptNames: []string{"surge", "L1", "forward"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"L2": {Name: "L2", FriendlyName: "Left/Right", FunscriptNames: []string{"sway", "L2", "left"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"R0": {Name: "R0", FriendlyName: "Twist", FunscriptNames: []string{"twist", "R0", "yaw"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"R1": {Name: "R1", FriendlyName: "Roll", FunscriptNames: []string{"roll", "R1"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"R2": {Name: "R2", FriendlyName: "Pitch", FunscriptNames: []string{"pitch", "R2"}, Enabled: true, DefaultValue: 0.5, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: true, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"V0": {Name: "V0", FriendlyName: "Vibrate", FunscriptNames: []string{"vib", "V0"}, Enabled: false, DefaultValue: 0.0, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: false, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"V1": {Name: "V1", FriendlyName: "Pump", FunscriptNames: []string{"pump", "V1"}, Enabled: false, DefaultValue: 0.0, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: false, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"A0": {Name: "A0", FriendlyName: "Valve", FunscriptNames: []string{"valve", "A0"}, Enabled: false, DefaultValue: 0.0, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: false, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"A1": {Name: "A1", FriendlyName: "Suction", FunscriptNames: []string{"suck", "A1"}, Enabled: false, DefaultValue: 0.0, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: false, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
		"A2": {Name: "A2", FriendlyName: "Lube", FunscriptNames: []string{"lube", "A2"}, Enabled: false, DefaultValue: 0.0, Min: 0.0, Max: 1.0, Offset: 0.0, Invert: false, Interpolation: "linear", AutoHome: false, AutoHomeDelay: 5.0, MotionType: "none", MotionSpeed: 1.0, MotionRange: 0.5, MotionOffset: 0.5},
	}

	players := map[string]*MediaSourceSettings{
		"internal":   {Type: "internal", Endpoint: "", Password: "", AutoConnect: false},
		"mpchc":      {Type: "mpchc", Endpoint: "127.0.0.1:13579", Password: "", AutoConnect: false},
		"mpv":        {Type: "mpv", Endpoint: "multifunplayer-mpv", Password: "", AutoConnect: false},
		"vlc":        {Type: "vlc", Endpoint: "127.0.0.1:8080", Password: "test", AutoConnect: false},
		"deovr":      {Type: "deovr", Endpoint: "127.0.0.1:23554", Password: "", AutoConnect: false},
		"heresphere": {Type: "heresphere", Endpoint: "127.0.0.1:23554", Password: "", AutoConnect: false},
	}

	outputs := map[string]*OutputTargetSettings{
		"serial":    {Type: "serial", Endpoint: "COM3", BaudRate: 115200, Enabled: false, AutoConnect: false},
		"udp":       {Type: "udp", Endpoint: "127.0.0.1:8000", Enabled: false, AutoConnect: false},
		"tcp":       {Type: "tcp", Endpoint: "127.0.0.1:8000", Enabled: false, AutoConnect: false},
		"websocket": {Type: "websocket", Endpoint: "ws://127.0.0.1:8000/ws", Enabled: false, AutoConnect: false},
		"file":      {Type: "file", FilePath: "output_record.txt", Enabled: false, AutoConnect: false},
	}

	return AppSettings{
		UpdateRate:        100,
		RememberLoc:       true,
		ActivePlayer:      "internal",
		Axes:              axes,
		MediaSources:      players,
		OutputTargets:     outputs,
		ScriptDirectories: []string{},
		GlobalOffset:      0.0,
		UseGlobalOffset:   true,
		AllowRemoteAccess: false,
		WebPort:           "5000",
		Language:          "en",
	}
}

// Load loads settings from file, or creates default if it doesn't exist
func (sm *SettingsManager) Load() error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	data, err := os.ReadFile(sm.path)
	if err != nil {
		if os.IsNotExist(err) {
			sm.Data = DefaultSettings()
			return sm.saveUnlocked()
		}
		return err
	}

	err = json.Unmarshal(data, &sm.Data)
	if err == nil {
		if sm.Data.Language == "" {
			sm.Data.Language = "en"
		}
	}
	return err
}

// Save saves configuration to settings file
func (sm *SettingsManager) Save() error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	return sm.saveUnlocked()
}

func (sm *SettingsManager) saveUnlocked() error {
	data, err := json.MarshalIndent(sm.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sm.path, data, 0644)
}
