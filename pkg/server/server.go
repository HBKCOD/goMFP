package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"goMFP/pkg/player"
	"goMFP/pkg/player/funscript"
	"goMFP/pkg/settings"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local app
	},
}

type WebServer struct {
	addr        string
	coord       *player.Coordinator
	sm          *settings.SettingsManager
	clients     map[*websocket.Conn]bool
	clientsMu   sync.Mutex
	tcodeLog    chan string
	stateBroadcast chan interface{}

	server      *http.Server
	serverMu    sync.Mutex
}

func NewWebServer(addr string, coord *player.Coordinator, sm *settings.SettingsManager) *WebServer {
	ws := &WebServer{
		addr:        addr,
		coord:       coord,
		sm:          sm,
		clients:     make(map[*websocket.Conn]bool),
		tcodeLog:    make(chan string, 100),
		stateBroadcast: make(chan interface{}, 10),
	}

	// Link coordinator log callbacks
	coord.StateBroadcastFunc = func(state interface{}) {
		select {
		case ws.stateBroadcast <- state:
		default:
			// Drop state if channel full to prevent blocking tick loop
		}
	}

	coord.TCodeLogFunc = func(tcode string) {
		select {
		case ws.tcodeLog <- tcode:
		default:
			// Drop log if buffer full
		}
	}

	return ws
}

func (ws *WebServer) Start() error {
	// Serve static files from web directory
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/ws", ws.handleWebSocket)

	// API for uploading scripts via HTTP (alternative to WS)
	http.HandleFunc("/api/upload", ws.handleScriptUpload)
	// API for querying simplified script keyframes
	http.HandleFunc("/api/script", ws.handleScriptQuery)
	// API for querying serial ports
	http.HandleFunc("/api/serial-ports", ws.handleSerialPortsQuery)
	// API for folder browsing
	http.HandleFunc("/api/browse-dir", ws.handleBrowseDir)

	// Start background broadcasters
	go ws.broadcastStateLoop()
	go ws.broadcastTCodeLoop()

	ws.sm.Mu.RLock()
	enabled := ws.sm.Data.AllowRemoteAccess
	ws.sm.Mu.RUnlock()

	if enabled {
		return ws.listenAndServe()
	}
	return nil
}

func (ws *WebServer) listenAndServe() error {
	ws.serverMu.Lock()
	
	host := "0.0.0.0"
	
	ws.sm.Mu.RLock()
	port := ws.sm.Data.WebPort
	ws.sm.Mu.RUnlock()
	if port == "" {
		port = "5000"
	}
	bindAddr := host + ":" + port

	ws.server = &http.Server{
		Addr: bindAddr,
	}
	ws.serverMu.Unlock()

	fmt.Printf("Server listening on http://%s\n", bindAddr)
	err := ws.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (ws *WebServer) RebindServer() {
	ws.serverMu.Lock()
	if ws.server != nil {
		fmt.Println("[서버] WebUI 설정 변경으로 인해 서버를 종료합니다...")
		ws.server.Shutdown(context.Background())
		ws.server = nil
	}
	ws.serverMu.Unlock()

	ws.sm.Mu.RLock()
	enabled := ws.sm.Data.AllowRemoteAccess
	ws.sm.Mu.RUnlock()

	if enabled {
		go func() {
			fmt.Println("[서버] WebUI 설정을 반영하여 서버를 기동합니다...")
			if err := ws.listenAndServe(); err != nil {
				fmt.Printf("[서버] 서버 기동 중 오류 발생: %v\n", err)
			}
		}()
	}
}

func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ws.clientsMu.Lock()
	ws.clients[conn] = true
	ws.clientsMu.Unlock()

	defer func() {
		ws.clientsMu.Lock()
		delete(ws.clients, conn)
		ws.clientsMu.Unlock()
	}()

	// Send current settings & state on connect
	ws.sendSettings(conn)

	// Message read loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req map[string]interface{}
		if err := json.Unmarshal(message, &req); err != nil {
			continue
		}

		action, ok := req["action"].(string)
		if !ok {
			continue
		}

		ws.handleWSAction(action, req, conn)
	}
}

func (ws *WebServer) handleWSAction(action string, req map[string]interface{}, conn *websocket.Conn) {
	switch action {
	case "play":
		ws.coord.Play()
	case "pause":
		ws.coord.Pause()
	case "seek":
		if pos, ok := req["position"].(float64); ok {
			ws.coord.Seek(pos)
		}
	case "set_player":
		if name, ok := req["name"].(string); ok {
			ws.coord.SetActivePlayer(name)
			ws.broadcastSettings()
		}
	case "connect_player":
		if name, ok := req["name"].(string); ok {
			ws.coord.ConnectPlayer(name)
		}
	case "disconnect_player":
		if name, ok := req["name"].(string); ok {
			ws.coord.DisconnectPlayer(name)
		}
	case "connect_output":
		if name, ok := req["name"].(string); ok {
			ws.coord.ConnectOutput(name)
		}
	case "disconnect_output":
		if name, ok := req["name"].(string); ok {
			ws.coord.DisconnectOutput(name)
		}
	case "save_axis":
		axisName, ok1 := req["axis"].(string)
		settingsMap, ok2 := req["settings"].(map[string]interface{})
		if ok1 && ok2 {
			data, err := json.Marshal(settingsMap)
			if err == nil {
				var axisSet settings.AxisSettings
				if err := json.Unmarshal(data, &axisSet); err == nil {
					ws.coord.SetAxisSettings(axisName, &axisSet)
					ws.broadcastSettings()
				}
			}
		}
	case "save_player":
		pName, ok1 := req["player"].(string)
		endpoint, ok2 := req["endpoint"].(string)
		password, _ := req["password"].(string)
		autoConnect, ok3 := req["auto_connect"].(bool)
		if ok1 && ok2 && ok3 {
			ws.coord.SetPlayerSettings(pName, endpoint, password, autoConnect)
			ws.broadcastSettings()
		}
	case "save_output":
		oName, ok1 := req["output"].(string)
		endpoint, ok2 := req["endpoint"].(string)
		baud, ok3 := req["baud_rate"].(float64)
		filePath, ok4 := req["file_path"].(string)
		autoConnect, ok5 := req["auto_connect"].(bool)
		if ok1 && ok2 && ok3 && ok4 && ok5 {
			ws.coord.SetOutputSettings(oName, endpoint, int(baud), filePath, autoConnect)
			ws.broadcastSettings()
		}
	case "load_manual_script":
		axis, ok1 := req["axis"].(string)
		path, ok2 := req["path"].(string)
		if ok1 && ok2 {
			ws.coord.LoadManualScript(axis, path)
		}
	case "save_script_directories":
		if dirs, ok := req["directories"].([]interface{}); ok {
			var strDirs []string
			for _, d := range dirs {
				if s, ok := d.(string); ok {
					strDirs = append(strDirs, s)
				}
			}
			ws.sm.Mu.Lock()
			ws.sm.Data.ScriptDirectories = strDirs
			ws.sm.Mu.Unlock()
			ws.sm.Save()
			ws.broadcastSettings()
		}
	case "save_global_offset":
		offset, ok1 := req["global_offset"].(float64)
		useGlobal, ok2 := req["use_global_offset"].(bool)
		if ok1 && ok2 {
			ws.sm.Mu.Lock()
			ws.sm.Data.GlobalOffset = offset
			ws.sm.Data.UseGlobalOffset = useGlobal
			if useGlobal {
				// Copy global offset to all axes settings
				for _, axis := range ws.sm.Data.Axes {
					axis.Offset = offset
				}
			}
			ws.sm.Mu.Unlock()
			ws.sm.Save()
			ws.broadcastSettings()
		}
	case "save_remote_access":
		allowRemote, ok := req["allow_remote_access"].(bool)
		if ok {
			ws.sm.Mu.Lock()
			changed := ws.sm.Data.AllowRemoteAccess != allowRemote
			ws.sm.Data.AllowRemoteAccess = allowRemote
			ws.sm.Mu.Unlock()
			ws.sm.Save()
			ws.broadcastSettings()
			if changed {
				ws.RebindServer()
			}
		}
	case "save_language":
		lang, ok := req["language"].(string)
		if ok {
			ws.sm.Mu.Lock()
			ws.sm.Data.Language = lang
			ws.sm.Mu.Unlock()
			ws.sm.Save()
			ws.broadcastSettings()
		}
	}
}

func (ws *WebServer) handleScriptUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	axis := r.FormValue("axis")
	if axis == "" {
		axis = "L0"
	}

	file, header, err := r.FormFile("script")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse in memory
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine file format
	var script *funscript.Script
	var multi map[string]*funscript.Script

	if strings.HasSuffix(strings.ToLower(header.Filename), ".funscript") {
		script, multi, err = funscript.ParseFunscript(&buf, header.Filename, "")
	} else if strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		script, err = funscript.ParseCSV(&buf, header.Filename, "")
	} else {
		http.Error(w, "Unsupported file format", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Failed to parse: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Apply script to coordinator
	ws.coord.LoadManualScriptData(axis, script, multi)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Loaded script: " + header.Filename))
}

func (ws *WebServer) handleScriptQuery(w http.ResponseWriter, r *http.Request) {
	axis := r.URL.Query().Get("axis")
	if axis == "" {
		axis = "L0"
	}

	kfs := ws.coord.GetScriptKeyframes(axis)
	if kfs == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"keyframes":[]}`))
		return
	}

	// Downsample keyframes to max 500 points for Web rendering
	maxPoints := 500
	var result []funscript.Keyframe
	if len(kfs) <= maxPoints {
		result = kfs
	} else {
		step := len(kfs) / maxPoints
		if step == 0 {
			step = 1
		}
		for i := 0; i < len(kfs); i += step {
			result = append(result, kfs[i])
		}
		// Ensure the last point is always included
		if result[len(result)-1].At != kfs[len(kfs)-1].At {
			result = append(result, kfs[len(kfs)-1])
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keyframes": result,
	})
}

func (ws *WebServer) broadcastStateLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // Throttled state broadcasts (10Hz is plenty)
	defer ticker.Stop()

	var lastState interface{}

	for {
		select {
		case state := <-ws.stateBroadcast:
			lastState = state
		case <-ticker.C:
			if lastState != nil {
				ws.broadcastMessage(map[string]interface{}{
					"type":  "state",
					"value": lastState,
				})
				lastState = nil
			}
		}
	}
}

func (ws *WebServer) broadcastTCodeLoop() {
	for tcode := range ws.tcodeLog {
		ws.broadcastMessage(map[string]interface{}{
			"type":  "tcode",
			"value": tcode,
		})
	}
}

func (ws *WebServer) sendSettings(conn *websocket.Conn) {
	ws.sm.Mu.RLock()
	defer ws.sm.Mu.RUnlock()

	data, err := json.Marshal(map[string]interface{}{
		"type":  "settings",
		"value": ws.sm.Data,
	})
	if err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (ws *WebServer) broadcastSettings() {
	ws.sm.Mu.RLock()
	defer ws.sm.Mu.RUnlock()

	ws.broadcastMessage(map[string]interface{}{
		"type":  "settings",
		"value": ws.sm.Data,
	})
}

func (ws *WebServer) broadcastMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	ws.clientsMu.Lock()
	defer ws.clientsMu.Unlock()

	for client := range ws.clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			client.Close()
			delete(ws.clients, client)
		}
	}
}

func (ws *WebServer) handleSerialPortsQuery(w http.ResponseWriter, r *http.Request) {
	ports, err := serial.GetPortsList()
	if err != nil {
		ports = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ports)
}

type BrowseResponse struct {
	CurrentPath string   `json:"current_path"`
	ParentPath  string   `json:"parent_path"`
	Directories []string `json:"directories"`
	Drives      []string `json:"drives"`
}

func getLogicalDrives() []string {
	var drives []string
	if runtime.GOOS == "windows" {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			path := string(drive) + ":\\"
			if _, err := os.Stat(path); err == nil {
				drives = append(drives, path)
			}
		}
	} else {
		drives = []string{"/"}
	}
	return drives
}

func (ws *WebServer) handleBrowseDir(w http.ResponseWriter, r *http.Request) {
	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			drives := getLogicalDrives()
			if len(drives) > 0 {
				dirPath = drives[0]
			} else {
				dirPath = "/"
			}
		} else {
			dirPath = wd
		}
	}

	entries, err := os.ReadDir(dirPath)
	var dirs []string
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				dirs = append(dirs, entry.Name())
			}
		}
	}

	parent := filepath.Dir(dirPath)
	if parent == dirPath {
		parent = ""
	}

	res := BrowseResponse{
		CurrentPath: dirPath,
		ParentPath:  parent,
		Directories: dirs,
		Drives:      getLogicalDrives(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
