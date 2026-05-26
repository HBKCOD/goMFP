package player

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"goMFP/pkg/player/funscript"
	"goMFP/pkg/player/mediasource"
	"goMFP/pkg/player/motion"
	"goMFP/pkg/player/outputtarget"
	"goMFP/pkg/settings"
)

// RealtimeAxisState holds current running values for a device axis
type RealtimeAxisState struct {
	Name          string  `json:"name"`
	Value         float64 `json:"value"`         // Current raw target value (0.0 - 1.0)
	ActualValue   float64 `json:"actual_value"`  // After range, offset, invert applied
	ScriptName    string  `json:"script_name"`   // Current loaded script name (empty if none)
	ScriptLoaded  bool    `json:"script_loaded"` // Is a script loaded for this axis?
	MotionActive  bool    `json:"motion_active"` // Is motion provider running?
	LastSentValue float64 // Last successfully sent value
	LastActivity  time.Time
	SyncTime      float64 // Remaining seek synchronization interpolation duration (seconds)
	LastRawValue  float64 // Previously computed raw target value (0.0 - 1.0)
}

// Coordinator manages the real-time tick loop, synchronization, and output routing
type Coordinator struct {
	mu           sync.RWMutex
	sm           *settings.SettingsManager
	activePlayer string

	players map[string]mediasource.MediaSource
	outputs map[string]outputtarget.OutputTarget

	// Axis runtime scripts and states
	scripts      map[string]*funscript.Script      // Loaded funscripts per axis
	axisStates   map[string]*RealtimeAxisState     // Real-time state per axis
	motionStates map[string]*motion.MotionState    // Motion provider states

	lastVideoPath          string
	lastTickTime           time.Time
	tickInterval           time.Duration
	cancelLoop             context.CancelFunc
	running                bool
	internalMediaPosition  float64
	lastPlayerRawPosition  float64

	// Callback to broadcast updates to the web clients
	StateBroadcastFunc func(state interface{})
	TCodeLogFunc       func(tcode string)
}

// NewCoordinator creates a coordinator
func NewCoordinator(sm *settings.SettingsManager) *Coordinator {
	// Initialize players
	players := map[string]mediasource.MediaSource{
		"internal":   mediasource.NewInternalPlayer(),
		"mpchc":      mediasource.NewMpcHcPlayer(),
		"vlc":        mediasource.NewVlcPlayer(),
		"deovr":      mediasource.NewDeoVrHereSpherePlayer("deovr"),
		"heresphere": mediasource.NewDeoVrHereSpherePlayer("heresphere"),
		"mpv":        mediasource.NewMpvPlayer(),
	}

	// Initialize outputs
	outputs := map[string]outputtarget.OutputTarget{
		"udp":       outputtarget.NewUdpOutput(),
		"tcp":       outputtarget.NewTcpOutput(),
		"websocket": outputtarget.NewWebSocketOutput(),
		"file":      outputtarget.NewFileOutput(),
		"serial":    outputtarget.NewSerialOutput(sm.Data.OutputTargets["serial"].BaudRate),
	}

	scripts := make(map[string]*funscript.Script)
	axisStates := make(map[string]*RealtimeAxisState)
	motionStates := make(map[string]*motion.MotionState)

	for name, axisSet := range sm.Data.Axes {
		scripts[name] = nil
		axisStates[name] = &RealtimeAxisState{
			Name:          name,
			Value:         axisSet.DefaultValue,
			ActualValue:   axisSet.DefaultValue,
			ScriptLoaded:  false,
			LastSentValue: math.NaN(),
			LastActivity:  time.Now(),
			SyncTime:      0.0,
			LastRawValue:  math.NaN(),
		}
		motionStates[name] = motion.NewMotionState(time.Now().UnixNano())
	}

	return &Coordinator{
		sm:                    sm,
		activePlayer:          sm.Data.ActivePlayer,
		players:               players,
		outputs:               outputs,
		scripts:               scripts,
		axisStates:            axisStates,
		motionStates:          motionStates,
		tickInterval:          time.Duration(1000/sm.Data.UpdateRate) * time.Millisecond,
		internalMediaPosition: 0.0,
		lastPlayerRawPosition: -1.0,
	}
}

// Start launches the update tick loop and attempts auto-connections
func (c *Coordinator) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return
	}

	c.running = true
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelLoop = cancel
	c.lastTickTime = time.Now()

	// 1. Auto connect players and outputs
	go c.autoConnectAll(ctx)

	// 2. Start core 50Hz update loop
	go c.tickLoop(ctx)
}

// Stop terminates the loops and disconnects everything
func (c *Coordinator) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.running = false
	if c.cancelLoop != nil {
		c.cancelLoop()
	}

	// Disconnect all players and outputs
	for _, p := range c.players {
		p.Disconnect()
	}
	for _, o := range c.outputs {
		o.Disconnect()
	}
}

func getConnectEndpoint(name string, pConf *settings.MediaSourceSettings) string {
	if name == "vlc" {
		return pConf.Endpoint + "|" + pConf.Password
	}
	return pConf.Endpoint
}

func (c *Coordinator) autoConnectAll(ctx context.Context) {
	c.mu.Lock()
	activePlayerName := c.activePlayer
	c.mu.Unlock()

	// Connect active player
	if p, ok := c.players[activePlayerName]; ok {
		pConf := c.sm.Data.MediaSources[activePlayerName]
		p.Connect(ctx, getConnectEndpoint(activePlayerName, pConf))
	}

	// Auto-connect other players if marked auto-connect
	for name, pConf := range c.sm.Data.MediaSources {
		if name != activePlayerName && pConf.AutoConnect {
			if p, ok := c.players[name]; ok {
				p.Connect(ctx, getConnectEndpoint(name, pConf))
			}
		}
	}

	// Auto-connect outputs
	for name, oConf := range c.sm.Data.OutputTargets {
		if oConf.AutoConnect || oConf.Enabled {
			if o, ok := c.outputs[name]; ok {
				var endpoint string
				if name == "file" {
					endpoint = oConf.FilePath
				} else {
					endpoint = oConf.Endpoint
				}
				o.Connect(endpoint)
			}
		}
	}
}

func (c *Coordinator) tickLoop(ctx context.Context) {
	ticker := time.NewTicker(c.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.performTick()
		}
	}
}

func (c *Coordinator) performTick() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	deltaTime := now.Sub(c.lastTickTime).Seconds()
	c.lastTickTime = now

	// 1. Fetch current player state
	activePlayer, ok := c.players[c.activePlayer]
	if !ok {
		return
	}

	pState := activePlayer.State()

	// 2. Check if played video changed to auto-load scripts
	if pState.Path != "" && pState.Path != c.lastVideoPath {
		c.lastVideoPath = pState.Path
		c.internalMediaPosition = pState.Position
		c.lastPlayerRawPosition = pState.Position
		go c.autoLoadScripts(pState.Path)
	}

	// 3. Interpolate media position to achieve smooth 1ms accuracy updates between slow player polls (e.g. VLC 200ms)
	if pState.Playing {
		// Accumulate time based on high-resolution local clock
		c.internalMediaPosition += deltaTime * pState.Speed

		// Adjust/sync whenever a new raw position coordinate comes from the player
		if pState.Position != c.lastPlayerRawPosition {
			c.lastPlayerRawPosition = pState.Position

			diff := c.internalMediaPosition - pState.Position
			if math.Abs(diff) > 1.0 {
				// Seek occurred, sync immediately
				c.internalMediaPosition = pState.Position
				// Trigger sync transition (damping) on all loaded axes to prevent jerks
				for _, state := range c.axisStates {
					if state.ScriptLoaded {
						state.SyncTime = 0.5
					}
				}
			} else {
				// Soft sync to absorb drift without sudden hardware jerks
				c.internalMediaPosition -= diff * 0.15
			}
		}
	} else {
		c.internalMediaPosition = pState.Position
		c.lastPlayerRawPosition = pState.Position
	}

	// Override player state position with our interpolated position for evaluation
	pState.Position = c.internalMediaPosition

	// 3. For each axis, compute its new value
	tcodeOutputs := []string{}
	dirtyAxes := []string{}

	// Determine output precision
	precision := 3
	if c.activePlayer == "heresphere" || c.activePlayer == "deovr" {
		precision = 4 // v0.3 standard for VR streaming usually
	}
	outputMaximum := math.Pow(10, float64(precision)) - 1

	for name, axisState := range c.axisStates {
		axisSet := c.sm.Data.Axes[name]
		if !axisSet.Enabled {
			continue
		}

		var targetValue float64
		script := c.scripts[name]

		// Track script evaluation
		scriptActive := false
		if pState.Playing && script != nil && len(script.Keyframes) > 0 && pState.Path != "" {
			// Apply time delay offset (seconds) before evaluating script
			evalPos := pState.Position - axisSet.Offset
			targetValue = script.Evaluate(evalPos, axisSet.Interpolation)
			scriptActive = true
			axisState.ScriptLoaded = true
			axisState.ScriptName = script.Name
			axisState.LastActivity = now
		} else if script != nil && len(script.Keyframes) > 0 && pState.Path != "" {
			// Script is loaded, but player is paused
			axisState.ScriptLoaded = true
			axisState.ScriptName = script.Name
		} else {
			axisState.ScriptLoaded = false
			axisState.ScriptName = ""
		}

		// Track motion provider evaluation
		motionActive := false
		if pState.Playing && !scriptActive && axisSet.MotionType != "none" {
			motionState := c.motionStates[name]
			motionState.Update(deltaTime, axisSet.MotionSpeed)
			// Compute value: maps motion pattern value (0.0-1.0) into configured Range/Offset
			motionVal := motionState.Calculate(axisSet.MotionType, 1, 1.0, 1.0)
			targetValue = motion.Map(motionVal, 0.0, 1.0, axisSet.Min, axisSet.Max) // Scale to min-max limit
			motionActive = true
			axisState.LastActivity = now
		}

		axisState.MotionActive = motionActive

		// Auto-home logic: if inactive for auto_home_delay, return to default value
		if !scriptActive && !motionActive && axisSet.AutoHome {
			inactiveDur := now.Sub(axisState.LastActivity).Seconds()
			if inactiveDur > axisSet.AutoHomeDelay {
				// Gradually smooth back to default value (Auto-home)
				dist := axisSet.DefaultValue - axisState.Value
				if math.Abs(dist) > 0.01 {
					// 1 second duration homing speed
					targetValue = axisState.Value + (dist * deltaTime * 2.0)
				} else {
					targetValue = axisSet.DefaultValue
				}
			} else {
				targetValue = axisState.Value
			}
		} else if !scriptActive && !motionActive {
			// Stay at current value if no auto-home
			targetValue = axisState.Value
		}

		// Apply seek synchronization interpolation (damping)
		if axisState.SyncTime > 0 {
			axisState.SyncTime -= deltaTime
			// Only interpolate if we are not autohoming and targetValue is finite
			if !(!scriptActive && !motionActive && axisSet.AutoHome) {
				duration := 0.5
				t := 1.0 - (axisState.SyncTime / duration)
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				// SmoothStep: t * t * (3.0 - 2.0 * t)
				smoothT := t * t * (3.0 - 2.0*t)

				fromVal := axisState.LastRawValue
				if math.IsNaN(fromVal) {
					fromVal = axisSet.DefaultValue
				}
				targetValue = fromVal + (targetValue-fromVal)*smoothT
			}
		}

		// Save raw computed value
		axisState.Value = targetValue

		// Apply transformations: Invert
		actualVal := targetValue
		if axisSet.Invert {
			actualVal = 1.0 - actualVal
		}

		// Apply Output Target Range (mirrors C# MathUtils.Lerp(Minimum, Maximum, value))
		actualVal = axisSet.Min + (axisSet.Max-axisSet.Min)*actualVal

		// Clamp to ensure value stays within Min/Max limits
		if actualVal < axisSet.Min {
			actualVal = axisSet.Min
		}
		if actualVal > axisSet.Max {
			actualVal = axisSet.Max
		}

		// Update UI values immediately for smooth visualization
		axisState.ActualValue = actualVal
		axisState.LastRawValue = axisState.Value

		// Determine if value is dirty (DeviceAxis.IsValueDirty logic in MultiFunPlayer)
		isDirty := false
		if math.IsNaN(axisState.LastSentValue) {
			isDirty = true
		} else {
			diff := math.Abs(axisState.LastSentValue - actualVal)
			if diff * (outputMaximum + 1) >= 1.0 {
				isDirty = true
			}
		}

		if isDirty {
			tcodeOutputs = append(tcodeOutputs, formatTCode(name, actualVal, precision))
			dirtyAxes = append(dirtyAxes, name)
		}
	}

	// 4. Send outputs if anything dirty (SendDirtyValuesOnly behaviour)
	if len(tcodeOutputs) > 0 {
		tcodeStr := strings.Join(tcodeOutputs, " ") + "\n"

		// Send to connected output targets
		sentAny := false
		for _, o := range c.outputs {
			if o.Status() == outputtarget.StatusConnected {
				if err := o.Send(tcodeStr); err == nil {
					sentAny = true
				}
			}
		}

		// Only update LastSentValue if transmission succeeded
		if sentAny {
			for _, name := range dirtyAxes {
				c.axisStates[name].LastSentValue = c.axisStates[name].ActualValue
			}
		}

		// Log TCode to interface
		if c.TCodeLogFunc != nil {
			c.TCodeLogFunc(tcodeStr)
		}
	}

	// 5. Broadcast real-time state to Web client UI (throttled/batched by webserver)
	if c.StateBroadcastFunc != nil {
		go c.StateBroadcastFunc(c.GetUIState(pState))
	}
}

// formatTCode formats a value to TCode string based on precision (3 = TCode v0.2, 4 = TCode v0.3)
func formatTCode(axis string, val float64, precision int) string {
	maxVal := math.Pow(10, float64(precision)) - 1
	intVal := int(math.Round(val * maxVal))
	if intVal < 0 {
		intVal = 0
	}
	if intVal > int(maxVal) {
		intVal = int(maxVal)
	}

	formatStr := fmt.Sprintf("%%s%%0%dd", precision)
	return fmt.Sprintf(formatStr, axis, intVal)
}

func (c *Coordinator) autoLoadScripts(videoPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	base := filepath.Base(videoPath)
	ext := filepath.Ext(base)
	videoNameNoExt := strings.TrimSuffix(base, ext)

	// Clear current scripts
	for name := range c.scripts {
		c.scripts[name] = nil
	}

	// Determine scan directories: user-configured script library paths first, video directory last
	scanDirs := []string{}
	for _, d := range c.sm.Data.ScriptDirectories {
		if d != "" {
			scanDirs = append(scanDirs, d)
		}
	}
	videoDir := filepath.Dir(videoPath)
	scanDirs = append(scanDirs, videoDir)

	// Scan each directory recursively
	for _, scanDir := range scanDirs {
		filepath.WalkDir(scanDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			fileName := d.Name()
			if !strings.HasPrefix(strings.ToLower(fileName), strings.ToLower(videoNameNoExt)) {
				return nil
			}

			// Match extension
			fileExt := strings.ToLower(filepath.Ext(fileName))
			if fileExt != ".funscript" && fileExt != ".csv" {
				return nil
			}

			// Check axis match
			// e.g. "Movie.funscript" or "Movie.surge.funscript"
			rest := strings.TrimPrefix(strings.ToLower(fileName), strings.ToLower(videoNameNoExt))
			rest = strings.TrimSuffix(rest, fileExt) // e.g. ".surge" or ""

			if rest == "" || rest == "." {
				// Maps to L0 by default
				c.loadScript(path, "L0")
			} else {
				// e.g. rest is ".surge" or ".lube"
				axisName := strings.TrimPrefix(rest, ".") // e.g. "surge"
				// Check if any configured axis has this funscript name mapping
				for axisKey, axisSet := range c.sm.Data.Axes {
					matched := false
					for _, mapName := range axisSet.FunscriptNames {
						if strings.EqualFold(mapName, axisName) {
							matched = true
							break
						}
					}
					if matched {
						c.loadScript(path, axisKey)
					}
				}
			}

			return nil
		})
	}
}

func (c *Coordinator) loadScript(path string, axis string) {
	s, multi, err := funscript.LoadScriptFromFile(path)
	if err != nil {
		return
	}

	if multi != nil && len(multi) > 0 {
		// Distribute multi axis scripts
		for ax, scr := range multi {
			if _, ok := c.scripts[ax]; ok {
				c.scripts[ax] = scr
			}
		}
	} else if s != nil {
		c.scripts[axis] = s
	}
}

// UIState is sent to the Web UI periodically
type UIState struct {
	PlayerName   string                        `json:"player_name"`
	PlayerStatus string                        `json:"player_status"`
	Playing      bool                          `json:"playing"`
	VideoPath    string                        `json:"video_path"`
	VideoTitle   string                        `json:"video_title"`
	Position     float64                       `json:"position"`
	Duration     float64                       `json:"duration"`
	Speed        float64                       `json:"speed"`
	Axes         map[string]*RealtimeAxisState `json:"axes"`
	Outputs      map[string]string             `json:"outputs"` // name -> status
}

func (c *Coordinator) GetUIState(pState mediasource.PlayerState) UIState {
	outputsState := make(map[string]string)
	for name, o := range c.outputs {
		outputsState[name] = string(o.Status())
	}

	axesState := make(map[string]*RealtimeAxisState)
	for name, state := range c.axisStates {
		axesState[name] = &RealtimeAxisState{
			Name:         state.Name,
			Value:        state.Value,
			ActualValue:  state.ActualValue,
			ScriptName:   state.ScriptName,
			ScriptLoaded: state.ScriptLoaded,
			MotionActive: state.MotionActive,
		}
	}

	videoTitle := filepath.Base(pState.Path)
	if pState.Path == "" {
		videoTitle = "No video loaded"
	}

	return UIState{
		PlayerName:   c.activePlayer,
		PlayerStatus: string(c.players[c.activePlayer].Status()),
		Playing:      pState.Playing,
		VideoPath:    pState.Path,
		VideoTitle:   videoTitle,
		Position:     pState.Position,
		Duration:     pState.Duration,
		Speed:        pState.Speed,
		Axes:         axesState,
		Outputs:      outputsState,
	}
}

// APIs exposed for settings changes
func (c *Coordinator) SetActivePlayer(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if p, ok := c.players[c.activePlayer]; ok {
		p.Disconnect()
	}

	c.activePlayer = name
	c.sm.Data.ActivePlayer = name
	c.sm.Save()

	// Trigger connection immediately
	if p, ok := c.players[name]; ok {
		ctx := context.Background()
		pConf := c.sm.Data.MediaSources[name]
		p.Connect(ctx, getConnectEndpoint(name, pConf))
	}
}

func (c *Coordinator) ConnectPlayer(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.players[name]; ok {
		pConf := c.sm.Data.MediaSources[name]
		p.Connect(context.Background(), getConnectEndpoint(name, pConf))
	}
}

func (c *Coordinator) DisconnectPlayer(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.players[name]; ok {
		p.Disconnect()
	}
}

func (c *Coordinator) ConnectOutput(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if o, ok := c.outputs[name]; ok {
		oConf := c.sm.Data.OutputTargets[name]
		var endpoint string
		if name == "file" {
			endpoint = oConf.FilePath
		} else {
			endpoint = oConf.Endpoint
		}
		o.Connect(endpoint)
		// Mark as enabled in settings so the UI ticker reflects connection intent
		oConf.Enabled = true
		c.sm.Save()
	}
}

func (c *Coordinator) DisconnectOutput(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if o, ok := c.outputs[name]; ok {
		o.Disconnect()
		// Mark as disabled in settings so the UI ticker reflects disconnection intent
		if oConf, ok2 := c.sm.Data.OutputTargets[name]; ok2 {
			oConf.Enabled = false
			c.sm.Save()
		}
	}
}

// GetOutputStatuses returns the actual runtime connection status for each output
func (c *Coordinator) GetOutputStatuses() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]string)
	for name, o := range c.outputs {
		res[name] = string(o.Status())
	}
	return res
}

func (c *Coordinator) SetAxisSettings(axisName string, set *settings.AxisSettings) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if oldSet, ok := c.sm.Data.Axes[axisName]; ok {
		if len(set.FunscriptNames) == 0 {
			set.FunscriptNames = oldSet.FunscriptNames
		}
		*oldSet = *set

		// If UseGlobalOffset is enabled, copy this offset to all other axes
		if c.sm.Data.UseGlobalOffset {
			for _, axis := range c.sm.Data.Axes {
				axis.Offset = set.Offset
			}
		}
		c.sm.Save()
	}
}

func (c *Coordinator) SetPlayerSettings(playerName string, endpoint string, password string, autoConnect bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if pConf, ok := c.sm.Data.MediaSources[playerName]; ok {
		pConf.Endpoint = endpoint
		pConf.Password = password
		pConf.AutoConnect = autoConnect
		c.sm.Save()

		// If connected/connecting, reconnect
		if p, ok := c.players[playerName]; ok {
			status := p.Status()
			wasConnected := status == mediasource.StatusConnected || status == mediasource.StatusConnecting
			p.Disconnect()
			if wasConnected {
				p.Connect(context.Background(), getConnectEndpoint(playerName, pConf))
			}
		}
	}
}

func (c *Coordinator) SetOutputSettings(outputName string, endpoint string, baudRate int, filePath string, autoConnect bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if oConf, ok := c.sm.Data.OutputTargets[outputName]; ok {
		oConf.Endpoint = endpoint
		oConf.BaudRate = baudRate
		oConf.FilePath = filePath
		oConf.AutoConnect = autoConnect
		c.sm.Save()

		if o, ok := c.outputs[outputName]; ok {
			o.Disconnect()
			var targetEndpoint string
			if outputName == "file" {
				targetEndpoint = filePath
			} else {
				targetEndpoint = endpoint
			}
			o.Connect(targetEndpoint)
		}
	}
}

// LoadManualScript allows loading a funscript file manually on a specific axis
func (c *Coordinator) LoadManualScript(axis string, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, multi, err := funscript.LoadScriptFromFile(path)
	if err != nil {
		return err
	}

	if multi != nil && len(multi) > 0 {
		for ax, scr := range multi {
			if _, ok := c.scripts[ax]; ok {
				c.scripts[ax] = scr
			}
		}
	} else if s != nil {
		c.scripts[axis] = s
	}
	return nil
}

// MediaControls returns interfaces to control media
func (c *Coordinator) Play() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.players[c.activePlayer]; ok {
		p.SetPlayPause(true)
	}
}

func (c *Coordinator) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.players[c.activePlayer]; ok {
		p.SetPlayPause(false)
	}
}

func (c *Coordinator) Seek(pos float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.players[c.activePlayer]; ok {
		p.Seek(pos)
	}
}

// GetScriptKeyframes returns a copy of the keyframes loaded for an axis
func (c *Coordinator) GetScriptKeyframes(axisName string) []funscript.Keyframe {
	c.mu.RLock()
	defer c.mu.RUnlock()

	s, ok := c.scripts[axisName]
	if !ok || s == nil {
		return nil
	}

	kfs := make([]funscript.Keyframe, len(s.Keyframes))
	copy(kfs, s.Keyframes)
	return kfs
}

// LoadManualScriptData applies raw script data to axis scripts map
func (c *Coordinator) LoadManualScriptData(axis string, s *funscript.Script, multi map[string]*funscript.Script) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if multi != nil && len(multi) > 0 {
		for ax, scr := range multi {
			if _, ok := c.scripts[ax]; ok {
				c.scripts[ax] = scr
			}
		}
	} else if s != nil {
		c.scripts[axis] = s
	}
}

func (c *Coordinator) GetActivePlayerName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activePlayer
}

func (c *Coordinator) GetActivePlayerStatus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if p, ok := c.players[c.activePlayer]; ok {
		return string(p.Status())
	}
	return "disconnected"
}

func (c *Coordinator) GetCurrentVideoTitle() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if p, ok := c.players[c.activePlayer]; ok {
		pState := p.State()
		if pState.Path != "" {
			return filepath.Base(pState.Path)
		}
	}
	return "-"
}

func (c *Coordinator) GetAxisActualValues() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]float64)
	for name, ax := range c.axisStates {
		res[name] = ax.ActualValue
	}
	return res
}

func (c *Coordinator) GetAxisComputedValues() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]float64)
	for name, ax := range c.axisStates {
		res[name] = ax.Value
	}
	return res
}

func (c *Coordinator) GetAxisScriptLoadedStates() map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]bool)
	for name, ax := range c.axisStates {
		res[name] = ax.ScriptLoaded
	}
	return res
}

func (c *Coordinator) GetPlaybackInfo() (float64, float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if p, ok := c.players[c.activePlayer]; ok {
		pState := p.State()
		return c.internalMediaPosition, pState.Duration, pState.Playing
	}
	return 0.0, 0.0, false
}

