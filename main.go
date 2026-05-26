package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"net"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"goMFP/pkg/player"
	"goMFP/pkg/player/funscript"
	"goMFP/pkg/server"
	"goMFP/pkg/settings"

	"go.bug.st/serial"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var langDict = map[string]map[string]string{
	"ko": {
		"title": "goMFP - Go MultiFunPlayer 제어 센터",
		"tab_media": "미디어 & 출력",
		"tab_axis": "축 상세 조정",
		"tab_tcode": "TCode 모니터",
		"tab_sys": "시스템 설정",
		"hdr_player": "플레이어 연동",
		"hdr_script": "스크립트 라이브러리 경로 설정",
		"hdr_output": "출력 채널 관리",
		"hdr_autohome": "오토 홈 (Auto Home)",
		"hdr_motion": "자체 모션 생성기 (Motion Provider)",
		"hdr_sys": "시스템 제어 설정",
		"hdr_playback": "플레이백 상태",
		
		"lbl_player_select": "플레이어 선택:",
		"lbl_address_pipe": "주소 / 파이프:",
		"lbl_vlc_password": "비밀번호 (VLC):",
		"chk_autoconnect": "시작 시 자동 연결",
		"btn_connect": "연결",
		"ph_script_dir": "예: C:\\funscripts",
		"btn_add": "추가",
		"chk_axis_enabled": "활성화 (Enabled)",
		"lbl_axis_minmax": "동작 한계 설정 (Min/Max Limit):",
		"lbl_axis_offset": "시간 차이 오프셋 (Offset Delay):",
		"chk_global_offset": "전역 오프셋 사용",
		"lbl_axis_interpolation": "보간 방식 (Interpolation):",
		"chk_axis_invert": "출력 반전 (Invert)",
		"chk_autohome_enabled": "오토 홈 활성화",
		"lbl_delay_sec": "딜레이(초):",
		"lbl_motion_type": "모션 타입:",
		"lbl_motion_speed": "속도 배율:",
		"lbl_motion_range": "진폭 (Range):",
		"lbl_motion_offset": "기준 (Offset):",
		"btn_manual_script": "수동 .funscript/.csv 로드",
		"btn_save_axis": "축 설정 저장",
		"btn_clear_log": "로그 지우기",
		"chk_remote_access": "WebUI 접속 허용 (웹서버 켜기)",
		"lbl_web_port": "웹 서버 포트 번호(Port):",
		"lbl_language": "언어 설정 (Language):",
		"btn_save_restart": "설정 저장 및 웹서버 재기동",
		"btn_open_web": "수동으로 브라우저(WebUI) 열기",
		
		"status_player_prefix": "플레이어: ",
		"status_offline": "오프라인",
		"status_connected": "연결됨",
		"status_connecting": "연결 중",
		"video_none": "재생 중인 비디오: 없음",
		"video_prefix": "비디오: ",
		
		"web_disabled": "웹 서버 상태: 비활성화됨 (접속 불가)",
		"web_address_prefix": "웹 서버 주소: http://%s:%s",
		
		"script_loaded_yes": "로드된 스크립트: 활성화됨",
		"script_loaded_no": "로드된 스크립트: 없음",
		
		"save_complete": "저장 완료",
		"save_axis_success": "[%s] 축 설정이 정상적으로 저장되었습니다.",
		"player_save_success": "플레이어 설정이 정상적으로 저장되었습니다.",
		"sys_save_success": "포트 및 원격접속 설정이 저장되었으며 웹서버가 재바인딩되었습니다.",
		
		"dialog_channel_settings": "%s 채널 설정",
		"dialog_btn_save": "저장",
		"dialog_btn_cancel": "취소",
		"dialog_lbl_filepath": "저장 파일 경로:",
		"dialog_lbl_serial_port": "시리얼 포트:",
		"dialog_lbl_baudrate": "Baud Rate:",
		"dialog_lbl_address": "연결 주소 (IP:PORT):",
		"dialog_lbl_autoconnect": "자동연결:",
		"axis_desc_format": "축 설명: %s",
		"hdr_lang": "언어 설정 (Language Settings)",
	},
	"en": {
		"title": "goMFP - Go MultiFunPlayer Control Center",
		"tab_media": "Media & Output",
		"tab_axis": "Axis Control & Tuning",
		"tab_tcode": "TCode Monitor",
		"tab_sys": "System Settings",
		"hdr_player": "Player Integration",
		"hdr_script": "Script Library Paths",
		"hdr_output": "Output Channels Manager",
		"hdr_autohome": "Auto Home Control",
		"hdr_motion": "Motion Generator (Motion Provider)",
		"hdr_sys": "System Control Settings",
		"hdr_playback": "Playback Status",
		
		"lbl_player_select": "Select Player:",
		"lbl_address_pipe": "Address / Pipe:",
		"lbl_vlc_password": "Password (VLC):",
		"chk_autoconnect": "Auto Connect on Start",
		"btn_connect": "Connect",
		"ph_script_dir": "e.g. C:\\funscripts",
		"btn_add": "Add",
		"chk_axis_enabled": "Enabled",
		"lbl_axis_minmax": "Limit (Min/Max):",
		"lbl_axis_offset": "Offset Delay:",
		"chk_global_offset": "Use Global Offset",
		"lbl_axis_interpolation": "Interpolation:",
		"chk_axis_invert": "Invert Output",
		"chk_autohome_enabled": "Enable Auto Home",
		"lbl_delay_sec": "Delay (sec):",
		"lbl_motion_type": "Motion Type:",
		"lbl_motion_speed": "Speed:",
		"lbl_motion_range": "Range:",
		"lbl_motion_offset": "Offset:",
		"btn_manual_script": "Load Script Manually",
		"btn_save_axis": "Save Axis Settings",
		"btn_clear_log": "Clear Logs",
		"chk_remote_access": "Allow WebUI Access (Start Web Server)",
		"lbl_web_port": "Web Server Port:",
		"lbl_language": "Language:",
		"btn_save_restart": "Save Settings & Restart Server",
		"btn_open_web": "Open Browser (WebUI) Manually",
		
		"status_player_prefix": "Player: ",
		"status_offline": "Offline",
		"status_connected": "Connected",
		"status_connecting": "Connecting",
		"video_none": "Video: None",
		"video_prefix": "Video: ",
		
		"web_disabled": "Web Server: Disabled (Access Blocked)",
		"web_address_prefix": "Web Server Address: http://%s:%s",
		
		"script_loaded_yes": "Loaded Script: Active",
		"script_loaded_no": "Loaded Script: None",
		
		"save_complete": "Save Complete",
		"save_axis_success": "[%s] Axis settings saved successfully.",
		"player_save_success": "Player settings saved successfully.",
		"sys_save_success": "Settings saved, web server re-bound.",
		
		"dialog_channel_settings": "%s Channel Settings",
		"dialog_btn_save": "Save",
		"dialog_btn_cancel": "Cancel",
		"dialog_lbl_filepath": "Save File Path:",
		"dialog_lbl_serial_port": "Serial Port:",
		"dialog_lbl_baudrate": "Baud Rate:",
		"dialog_lbl_address": "Connection Address (IP:PORT):",
		"dialog_lbl_autoconnect": "Auto Connect:",
		"axis_desc_format": "Axis Description: %s",
		"hdr_lang": "Language Settings",
	},
}

type ClickableBox struct {
	widget.BaseWidget
	OnTapped func(pos fyne.Position)
}

func NewClickableBox(onTapped func(pos fyne.Position)) *ClickableBox {
	box := &ClickableBox{OnTapped: onTapped}
	box.ExtendBaseWidget(box)
	return box
}

type ClickableBoxRenderer struct {
	rect *canvas.Rectangle
}

func (r *ClickableBoxRenderer) Destroy() {}
func (r *ClickableBoxRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
}
func (r *ClickableBoxRenderer) MinSize() fyne.Size {
	return fyne.NewSize(1, 1)
}
func (r *ClickableBoxRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}
func (r *ClickableBoxRenderer) Refresh() {}

func (b *ClickableBox) CreateRenderer() fyne.WidgetRenderer {
	return &ClickableBoxRenderer{rect: canvas.NewRectangle(color.Transparent)}
}

func (b *ClickableBox) Tapped(pe *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped(pe.Position)
	}
}

func main() {
	var applyFyneLanguage func(string)
	
	// Parse command line arguments to check for logging option
	enableLog := false
	for _, arg := range os.Args {
		if arg == "--log" || arg == "-log" || arg == "-d" || arg == "--debug" {
			enableLog = true
			break
		}
	}

	var logFile *os.File
	var logErr error
	var logWrite func(string)

	if enableLog {
		logFile, logErr = os.OpenFile("goMFP_startup.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if logErr == nil {
			logWrite = func(msg string) {
				logFile.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05.000"), msg))
				logFile.Sync()
			}
		} else {
			logWrite = func(msg string) {}
		}
	} else {
		logWrite = func(msg string) {}
	}

	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	logWrite("프로그램 시작됨 (main() 진입)")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			errStr := fmt.Sprintf("프로그램이 예기치 못한 패닉 오류로 종료되었습니다:\n%v\n\n로그 파일(goMFP_startup.log)을 확인하십시오.", r)
			logWrite(fmt.Sprintf("PANIC OCCURRED: %v\nSTACK TRACE:\n%s", r, string(stack)))
			showNativeMessageBox("goMFP 패닉 에러", errStr)
		}
	}()

	// 1. Settings Initialization
	logWrite("1. 설정 파일 로드 시작")
	settingsPath := "settings.json"
	sm := settings.NewSettingsManager(settingsPath)
	if err := sm.Load(); err != nil {
		logWrite("설정 파일 로드 실패: " + err.Error())
		showNativeMessageBox("goMFP 에러", "설정 파일을 로드하는 중 오류가 발생했습니다: "+err.Error())
		os.Exit(1)
	}

	sm.Mu.Lock()
	if sm.Data.WebPort == "" {
		sm.Data.WebPort = "5000"
	}
	sm.Mu.Unlock()
	sm.Save()
	logWrite("설정 파일 로드 완료")

	// 2. Start Player Coordinator
	logWrite("2. 플레이어 코디네이터 시작")
	coord := player.NewCoordinator(sm)
	coord.Start()
	logWrite("플레이어 코디네이터 시작 완료")

	// 3. Start Web Server
	logWrite("3. 웹 서버 시작")
	ws := server.NewWebServer("127.0.0.1:5000", coord, sm)
	go func() {
		if err := ws.Start(); err != nil {
			logWrite("웹 서버 시작 실패: " + err.Error())
			showNativeMessageBox("goMFP 에러", "웹 서버를 시작하는 중 오류가 발생했습니다.\n5000번 포트가 이미 사용 중이거나 바인딩에 실패했습니다.\n\n에러 메시지: "+err.Error())
			coord.Stop()
			os.Exit(1)
		}
	}()
	logWrite("웹 서버 백그라운드 구동 시작")

	// 4. Initialize Fyne Application
	logWrite("4. Fyne 앱 인스턴스 생성 시작")
	myApp := app.NewWithID("com.gomfp.app")
	logWrite("Fyne app.NewWithID 완료")
	myWindow := myApp.NewWindow("goMFP - Go MultiFunPlayer 제어 센터")
	logWrite("Fyne NewWindow 완료")
	myWindow.Resize(fyne.NewSize(900, 680))

	// 5. GUI State & Bindings Setup
	logWrite("5. GUI State & Bindings Setup 시작")
	playerStatusBind := binding.NewString()
	playerStatusBind.Set("플레이어: -")

	videoTitleBind := binding.NewString()
	videoTitleBind.Set("재생 중인 비디오: 없음")

	videoTimeBind := binding.NewString()
	videoTimeBind.Set("00:00 / 00:00")

	videoProgressBind := binding.NewFloat()
	videoProgressBind.Set(0.0)

	serverUrlBind := binding.NewString()
	updateServerURL := func() {
		sm.Mu.RLock()
		port := sm.Data.WebPort
		remote := sm.Data.AllowRemoteAccess
		lang := sm.Data.Language
		sm.Mu.RUnlock()

		dict := langDict[lang]
		if dict == nil {
			dict = langDict["ko"]
		}

		if !remote {
			serverUrlBind.Set(dict["web_disabled"])
			return
		}

		host := getLocalIP()
		serverUrlBind.Set(fmt.Sprintf(dict["web_address_prefix"], host, port))
	}
	updateServerURL()
	logWrite("5.1. updateServerURL 완료")

	// 11개 축의 미니 프로그레스 바 & 라벨 바인딩 리스트 생성
	axesKeys := []string{"L0", "L1", "L2", "R0", "R1", "R2", "V0", "V1", "A0", "A1", "A2"}
	axisProgressBinds := make(map[string]binding.Float)
	axisLabelBinds := make(map[string]binding.String)
	axisScriptStatusBinds := make(map[string]binding.String)

	for _, key := range axesKeys {
		axisProgressBinds[key] = binding.NewFloat()
		axisProgressBinds[key].Set(0.5)

		axisLabelBinds[key] = binding.NewString()
		axisLabelBinds[key].Set("0.50")

		axisScriptStatusBinds[key] = binding.NewString()
		axisScriptStatusBinds[key].Set("off")
	}
	logWrite("5.2. Axes bindings 완료")

	// TCode 콘솔 로그용 바인딩 및 버퍼링
	tcodeConsoleBind := binding.NewString()
	tcodeConsoleBind.Set("")

	var tcodeLogBuffer []string
	var tcodeLogMu sync.Mutex

	originalTCodeLog := coord.TCodeLogFunc
	coord.TCodeLogFunc = func(tcode string) {
		if originalTCodeLog != nil {
			originalTCodeLog(tcode)
		}
		tcodeLogMu.Lock()
		tcodeLogBuffer = append(tcodeLogBuffer, tcode)
		if len(tcodeLogBuffer) > 25 { // 터미널 렉 방지 및 25줄 최대 표시 유지
			tcodeLogBuffer = tcodeLogBuffer[len(tcodeLogBuffer)-25:]
		}
		tcodeLogMu.Unlock()
	}

	// ==============================================================
	// [탭 1: 미디어 & 출력 장치 제어]
	// ==============================================================

	// 1. 플레이어 연동 카드
	playerSelectLabel := widget.NewLabel("플레이어 선택:")
	endpointLabel := widget.NewLabel("주소 / 파이프:")
	passwordLabel := widget.NewLabel("비밀번호 (VLC):")

	playerSelect := widget.NewSelect([]string{"internal", "mpchc", "mpv", "vlc", "deovr", "heresphere"}, func(selected string) {
		coord.SetActivePlayer(selected)
	})
	sm.Mu.RLock()
	activePlayer := sm.Data.ActivePlayer
	sm.Mu.RUnlock()
	playerSelect.SetSelected(activePlayer)

	endpointEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()
	autoConnectCheck := widget.NewCheck("시작 시 자동 연결", nil)

	updatePlayerInputs := func(playerName string) {
		sm.Mu.RLock()
		conf, ok := sm.Data.MediaSources[playerName]
		sm.Mu.RUnlock()
		if ok {
			endpointEntry.SetText(conf.Endpoint)
			passwordEntry.SetText(conf.Password)
			autoConnectCheck.Checked = conf.AutoConnect
			autoConnectCheck.Refresh()
		}
		if playerName == "vlc" {
			passwordEntry.Enable()
		} else {
			passwordEntry.Disable()
		}
	}
	updatePlayerInputs(playerSelect.Selected)

	playerSelect.OnChanged = func(selected string) {
		coord.SetActivePlayer(selected)
		updatePlayerInputs(selected)
	}

	playerConnectBtn := widget.NewButton("연결", func() {
		coord.SetPlayerSettings(playerSelect.Selected, endpointEntry.Text, passwordEntry.Text, autoConnectCheck.Checked)
		coord.ConnectPlayer(playerSelect.Selected)
	})

	playerSaveBtn := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		coord.SetPlayerSettings(playerSelect.Selected, endpointEntry.Text, passwordEntry.Text, autoConnectCheck.Checked)
		sm.Mu.RLock()
		lang := sm.Data.Language
		sm.Mu.RUnlock()
		dict := langDict[lang]
		if dict == nil {
			dict = langDict["ko"]
		}
		dialog.ShowInformation(dict["save_complete"], dict["player_save_success"], myWindow)
	})

	playerHeaderLabel := widget.NewLabelWithStyle("플레이어 연동", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	playerControls := container.NewVBox(
		playerHeaderLabel,
		container.NewGridWithColumns(2, playerSelectLabel, playerSelect),
		container.NewGridWithColumns(2, endpointLabel, endpointEntry),
		container.NewGridWithColumns(2, passwordLabel, passwordEntry),
		container.NewBorder(nil, nil, autoConnectCheck, playerSaveBtn, playerConnectBtn),
	)
	logWrite("5.3. playerControls 완료")

	// 2. 스크립트 라이브러리 목록 카드 (Cgo 없이 동작하는 widget.NewList로 직관적 설계)
	var scriptDirsList *widget.List
	scriptDirsList = widget.NewList(
		func() int {
			sm.Mu.RLock()
			defer sm.Mu.RUnlock()
			return len(sm.Data.ScriptDirectories)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				widget.NewLabel("template_path"),
			)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			sm.Mu.RLock()
			pathText := ""
			if id < len(sm.Data.ScriptDirectories) {
				pathText = sm.Data.ScriptDirectories[id]
			}
			sm.Mu.RUnlock()

			border := o.(*fyne.Container)
			lbl := border.Objects[0].(*widget.Label)
			lbl.SetText(pathText)

			btn := border.Objects[1].(*widget.Button)
			btn.OnTapped = func() {
				sm.Mu.Lock()
				dirs := sm.Data.ScriptDirectories
				if id >= 0 && id < len(dirs) {
					dirs = append(dirs[:id], dirs[id+1:]...)
					sm.Data.ScriptDirectories = dirs
				}
				sm.Mu.Unlock()
				sm.Save()
				scriptDirsList.Refresh()
			}
		},
	)
	logWrite("5.4. scriptDirsList 완료")

	addPathEntry := widget.NewEntry()
	addPathEntry.SetPlaceHolder("예: C:\\funscripts")

	browseFolderBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			addPathEntry.SetText(list.Path())
		}, myWindow)
	})

	addPathBtn := widget.NewButton("추가", func() {
		newPath := addPathEntry.Text
		if newPath == "" {
			return
		}
		sm.Mu.Lock()
		// 중복 체크
		exists := false
		for _, d := range sm.Data.ScriptDirectories {
			if strings.EqualFold(d, newPath) {
				exists = true
				break
			}
		}
		if !exists {
			sm.Data.ScriptDirectories = append(sm.Data.ScriptDirectories, newPath)
		}
		sm.Mu.Unlock()
		sm.Save()
		scriptDirsList.Refresh()
		addPathEntry.SetText("")
	})

	scriptDirsHeaderLabel := widget.NewLabelWithStyle("스크립트 라이브러리 경로 설정", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	scriptDirsCard := container.NewBorder(
		scriptDirsHeaderLabel,
		container.NewBorder(nil, nil, nil, addPathBtn, container.NewBorder(nil, nil, nil, browseFolderBtn, addPathEntry)),
		nil, nil,
		scriptDirsList,
	)
	logWrite("5.5. scriptDirsCard 완료")

	// 3. 미디어 재생 정보 및 히트맵 카드
	playBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() { coord.Play() })
	pauseBtn := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() { coord.Pause() })

	playbackSlider := widget.NewSlider(0, 100)
	playbackSlider.Step = 1.0
	playbackSlider.OnChanged = func(val float64) {
		coord.Seek(val)
	}

	// 히트맵 그리기용 Raster Canvas
	var axisHeatmaps = make(map[string][]funscript.Keyframe)
	var axisColors = map[string]color.RGBA{
		"L0": {R: 0, G: 242, B: 254, A: 255}, // Cyan (L0은 불투명)
		"L1": {R: 76, G: 175, B: 80, A: 75},   // 30% transparency
		"L2": {R: 205, G: 220, B: 57, A: 75},
		"R0": {R: 244, G: 67, B: 54, A: 75},
		"R1": {R: 255, G: 152, B: 0, A: 75},
		"R2": {R: 233, G: 30, B: 99, A: 75},
		"V0": {R: 156, G: 39, B: 176, A: 75},
		"V1": {R: 63, G: 81, B: 181, A: 75},
		"A0": {R: 0, G: 150, B: 136, A: 75},
		"A1": {R: 255, G: 193, B: 7, A: 75},
		"A2": {R: 255, G: 87, B: 34, A: 75},
	}
	var heatmapBgImage *image.RGBA
	var heatmapBgMu sync.Mutex
	var mediaDuration float64 = 0.0
	var mediaPosition float64 = 0.0

	heatmapRaster := canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))

		heatmapBgMu.Lock()
		if heatmapBgImage == nil || heatmapBgImage.Bounds().Dx() != w || heatmapBgImage.Bounds().Dy() != h {
			heatmapBgImage = image.NewRGBA(image.Rect(0, 0, w, h))
			// Background fill
			backgroundColor := color.RGBA{R: 16, G: 20, B: 35, A: 255}
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					heatmapBgImage.Set(x, y, backgroundColor)
				}
			}

			hasAnyKfs := false
			for _, kfs := range axisHeatmaps {
				if len(kfs) > 0 {
					hasAnyKfs = true
					break
				}
			}

			if !hasAnyKfs || mediaDuration <= 0 {
				// No script indicator text (draw simple grid)
				gridColor := color.RGBA{R: 255, G: 255, B: 255, A: 20}
				for x := 0; x < w; x += 20 {
					for y := 0; y < h; y++ {
						if x%100 == 0 || y%20 == 0 {
							heatmapBgImage.Set(x, y, gridColor)
						}
					}
				}
			} else {
				// Draw L0 Keyframes curve first (background, prominent)
				if kfs, ok := axisHeatmaps["L0"]; ok && len(kfs) > 0 {
					col := axisColors["L0"]
					for i := 0; i < len(kfs)-1; i++ {
						p0 := kfs[i]
						p1 := kfs[i+1]

						x0 := int((float64(p0.At) / mediaDuration) * float64(w))
						y0 := h - int(p0.Pos*float64(h))

						x1 := int((float64(p1.At) / mediaDuration) * float64(w))
						y1 := h - int(p1.Pos*float64(h))

						drawLineAA(heatmapBgImage, x0, y0, x1, y1, col)
					}
				}

				// Draw non-L0 Keyframes curves last (foreground, transparent)
				for key, kfs := range axisHeatmaps {
					if key == "L0" || len(kfs) == 0 {
						continue
					}
					col, ok := axisColors[key]
					if !ok {
						col = color.RGBA{R: 255, G: 255, B: 255, A: 75}
					}
					for i := 0; i < len(kfs)-1; i++ {
						p0 := kfs[i]
						p1 := kfs[i+1]

						x0 := int((float64(p0.At) / mediaDuration) * float64(w))
						y0 := h - int(p0.Pos*float64(h))

						x1 := int((float64(p1.At) / mediaDuration) * float64(w))
						y1 := h - int(p1.Pos*float64(h))

						drawLineAA(heatmapBgImage, x0, y0, x1, y1, col)
					}
				}
			}
		}

		// Fast copy of the cached background
		copy(img.Pix, heatmapBgImage.Pix)
		heatmapBgMu.Unlock()

		// Draw current position line
		if mediaPosition >= 0 && mediaPosition <= mediaDuration {
			currentX := int((mediaPosition / mediaDuration) * float64(w))
			cursorColor := color.RGBA{R: 255, G: 23, B: 68, A: 255}
			for y := 0; y < h; y++ {
				drawPixelBlend(img, currentX, y, cursorColor, 1.0)
				if currentX > 0 {
					drawPixelBlend(img, currentX-1, y, cursorColor, 0.5)
				}
				if currentX < w-1 {
					drawPixelBlend(img, currentX+1, y, cursorColor, 0.5)
				}
			}
		}

		// Draw glow circle at current point on L0
		if kfs, ok := axisHeatmaps["L0"]; ok && len(kfs) > 0 && mediaPosition >= 0 && mediaPosition <= mediaDuration {
			currentX := int((mediaPosition / mediaDuration) * float64(w))
			curVal := interpolateValue(mediaPosition, kfs)
			currentY := h - int(curVal*float64(h))

			cursorColor := color.RGBA{R: 255, G: 23, B: 68, A: 255}
			drawGlowCircle(img, currentX, currentY, 4, cursorColor)
		}

		return img
	})
	heatmapRaster.SetMinSize(fyne.NewSize(350, 60))

	var clickableBox *ClickableBox
	clickableBox = NewClickableBox(func(pos fyne.Position) {
		if mediaDuration <= 0 {
			return
		}
		w := clickableBox.Size().Width
		if w <= 0 {
			return
		}
		pct := float64(pos.X) / float64(w)
		if pct < 0 {
			pct = 0
		}
		if pct > 1 {
			pct = 1
		}
		coord.Seek(pct * mediaDuration)
	})

	clickableHeatmap := container.NewStack(
		heatmapRaster,
		clickableBox,
	)

	playbackInfoCard := widget.NewCard("플레이백 상태", "", container.NewVBox(
		widget.NewLabelWithData(videoTitleBind),
		container.NewBorder(nil, nil, nil, widget.NewLabelWithData(videoTimeBind), playbackSlider),
		clickableHeatmap,
		container.NewCenter(container.NewHBox(playBtn, pauseBtn)),
	))
	logWrite("5.6. playbackInfoCard 완료")

	// 4. 출력 채널 관리 리스트
	outputsHeaderLabel := widget.NewLabelWithStyle("출력 채널 관리", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	outputsContainer := container.NewVBox(outputsHeaderLabel)
	outputCheckboxes := make(map[string]*widget.Check)
	outputLabels := make(map[string]*widget.Label)

	for _, outKey := range []string{"udp", "tcp", "websocket", "serial", "file"} {
		key := outKey // closure binding
		sm.Mu.RLock()
		conf := sm.Data.OutputTargets[key]
		sm.Mu.RUnlock()

		addrText := conf.Endpoint
		if key == "file" {
			addrText = conf.FilePath
		} else if key == "serial" {
			addrText = fmt.Sprintf("%s (%dbps)", conf.Endpoint, conf.BaudRate)
		}

		chk := widget.NewCheck(strings.ToUpper(key), func(checked bool) {
			if checked {
				coord.ConnectOutput(key)
			} else {
				coord.DisconnectOutput(key)
			}
		})
		chk.Checked = conf.Enabled

		lbl := widget.NewLabel(addrText)

		configBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			// 출력 상세 설정 Form Dialog 오픈
			sm.Mu.RLock()
			currentConf := sm.Data.OutputTargets[key]
			currentLang := sm.Data.Language
			sm.Mu.RUnlock()

			dict := langDict[currentLang]
			if dict == nil {
				dict = langDict["ko"]
			}

			addrEntry := widget.NewEntry()
			addrEntry.SetText(currentConf.Endpoint)
			baudEntry := widget.NewEntry()
			baudEntry.SetText(strconv.Itoa(currentConf.BaudRate))
			fileEntry := widget.NewEntry()
			fileEntry.SetText(currentConf.FilePath)
			autoChk := widget.NewCheck(dict["chk_autoconnect"], nil)
			autoChk.Checked = currentConf.AutoConnect

			var serialPortEntry *widget.SelectEntry

			formItems := []*widget.FormItem{}
			if key == "file" {
				formItems = append(formItems, widget.NewFormItem(dict["dialog_lbl_filepath"], fileEntry))
			} else if key == "serial" {
				ports, _ := serial.GetPortsList()
				serialPortEntry = widget.NewSelectEntry(ports)
				serialPortEntry.SetText(currentConf.Endpoint)
				formItems = append(formItems, widget.NewFormItem(dict["dialog_lbl_serial_port"], serialPortEntry))
				formItems = append(formItems, widget.NewFormItem(dict["dialog_lbl_baudrate"], baudEntry))
			} else {
				formItems = append(formItems, widget.NewFormItem(dict["dialog_lbl_address"], addrEntry))
			}
			formItems = append(formItems, widget.NewFormItem(dict["dialog_lbl_autoconnect"], autoChk))

			dialog.ShowForm(fmt.Sprintf(dict["dialog_channel_settings"], strings.ToUpper(key)), dict["dialog_btn_save"], dict["dialog_btn_cancel"], formItems, func(ok bool) {
				if ok {
					bRate, _ := strconv.Atoi(baudEntry.Text)
					endpointVal := addrEntry.Text
					if key == "serial" && serialPortEntry != nil {
						endpointVal = serialPortEntry.Text
					}
					coord.SetOutputSettings(key, endpointVal, bRate, fileEntry.Text, autoChk.Checked)
					
					// 레이블 업데이트
					sm.Mu.RLock()
					newConf := sm.Data.OutputTargets[key]
					sm.Mu.RUnlock()
					newText := newConf.Endpoint
					if key == "file" {
						newText = newConf.FilePath
					} else if key == "serial" {
						newText = fmt.Sprintf("%s (%dbps)", newConf.Endpoint, newConf.BaudRate)
					}
					lbl.SetText(newText)
				}
			}, myWindow)
		})

		outputCheckboxes[key] = chk
		outputLabels[key] = lbl

		outputsContainer.Add(container.NewBorder(nil, nil, chk, configBtn, lbl))
	}
	logWrite("5.7. Output channels 완료")

	leftSplit := container.NewVSplit(playerControls, scriptDirsCard)
	leftSplit.Offset = 0.45
	rightSplit := container.NewVSplit(playbackInfoCard, outputsContainer)
	rightSplit.Offset = 0.50

	mainSplit := container.NewHSplit(leftSplit, rightSplit)
	mainSplit.Offset = 0.50
	logWrite("5.8. mainSplit 완료")

	// ==============================================================
	// [탭 2: 축 상세 조정 (Axis Settings)]
	// ==============================================================

	currentAxis := "L0"

	// 축 상세 폼 위젯 정의
	axisFriendlyNameLabel := widget.NewLabel("Friendly Name: -")
	axisEnabledCheck := widget.NewCheck("활성화 (Enabled)", nil)
	axisMinSlider := widget.NewSlider(0, 1)
	axisMaxSlider := widget.NewSlider(0, 1)
	axisMinSlider.Step = 0.01
	axisMaxSlider.Step = 0.01
	axisMinLabel := widget.NewLabel("Min: 0.00")
	axisMaxLabel := widget.NewLabel("Max: 1.00")

	axisOffsetSlider := widget.NewSlider(-1, 1)
	axisOffsetSlider.Step = 0.01
	axisOffsetLabel := widget.NewLabel("Offset: 0.00s")
	axisUseGlobalOffsetCheck := widget.NewCheck("전역 오프셋 사용", nil)

	axisInterpolationSelect := widget.NewSelect([]string{"linear", "pchip", "makima", "step"}, nil)
	axisInvertCheck := widget.NewCheck("출력 반전 (Invert)", nil)

	axisAutoHomeCheck := widget.NewCheck("오토 홈 활성화", nil)
	axisAutoHomeDelayEntry := widget.NewEntry()

	axisMotionTypeSelect := widget.NewSelect([]string{"none", "sine", "triangle", "saw", "square", "double_bounce", "sharp_bounce", "random"}, nil)
	axisMotionSpeedEntry := widget.NewEntry()
	axisMotionRangeSlider := widget.NewSlider(0, 1)
	axisMotionRangeSlider.Step = 0.01
	axisMotionOffsetSlider := widget.NewSlider(0, 1)
	axisMotionOffsetSlider.Step = 0.01

	axisScriptInfoLabel := widget.NewLabel("로드된 스크립트: 없음")
	logWrite("5.9. Axis detail form widgets 완료")

	// Drag/Drop 수동 스크립트 로드 버튼
	loadScriptManualBtn := widget.NewButton("수동 .funscript/.csv 로드", func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			coord.LoadManualScript(currentAxis, uri.URI().Path())
			
			sm.Mu.RLock()
			lang := sm.Data.Language
			sm.Mu.RUnlock()
			dict := langDict[lang]
			if dict == nil {
				dict = langDict["ko"]
			}
			dialog.ShowInformation(dict["save_complete"], dict["script_load_msg"], myWindow)
		}, myWindow)
	})

	updateAxisFormUI := func(axisKey string) {
		sm.Mu.RLock()
		set, ok := sm.Data.Axes[axisKey]
		useGlobalOffset := sm.Data.UseGlobalOffset
		lang := sm.Data.Language
		sm.Mu.RUnlock()

		if !ok {
			return
		}

		dict := langDict[lang]
		if dict == nil {
			dict = langDict["ko"]
		}

		axisFriendlyNameLabel.SetText(fmt.Sprintf(dict["axis_desc_format"], set.FriendlyName))
		axisEnabledCheck.Checked = set.Enabled
		axisEnabledCheck.Refresh()

		axisMinSlider.SetValue(set.Min)
		axisMaxSlider.SetValue(set.Max)
		axisMinLabel.SetText(fmt.Sprintf("Min: %.2f", set.Min))
		axisMaxLabel.SetText(fmt.Sprintf("Max: %.2f", set.Max))

		axisOffsetSlider.SetValue(set.Offset)
		axisOffsetLabel.SetText(fmt.Sprintf("Offset: %.2fs", set.Offset))
		axisUseGlobalOffsetCheck.Checked = useGlobalOffset
		axisUseGlobalOffsetCheck.Refresh()

		axisInterpolationSelect.SetSelected(set.Interpolation)
		axisInvertCheck.Checked = set.Invert
		axisInvertCheck.Refresh()

		axisAutoHomeCheck.Checked = set.AutoHome
		axisAutoHomeCheck.Refresh()
		axisAutoHomeDelayEntry.SetText(fmt.Sprintf("%.1f", set.AutoHomeDelay))

		axisMotionTypeSelect.SetSelected(set.MotionType)
		axisMotionSpeedEntry.SetText(fmt.Sprintf("%.2f", set.MotionSpeed))
		axisMotionRangeSlider.SetValue(set.MotionRange)
		axisMotionOffsetSlider.SetValue(set.MotionOffset)
	}

	axisMinSlider.OnChanged = func(val float64) {
		axisMinLabel.SetText(fmt.Sprintf("Min: %.2f", val))
		if val > axisMaxSlider.Value {
			axisMaxSlider.SetValue(val)
		}
	}
	axisMaxSlider.OnChanged = func(val float64) {
		axisMaxLabel.SetText(fmt.Sprintf("Max: %.2f", val))
		if val < axisMinSlider.Value {
			axisMinSlider.SetValue(val)
		}
	}
	axisOffsetSlider.OnChanged = func(val float64) {
		axisOffsetLabel.SetText(fmt.Sprintf("Offset: %.2fs", val))
	}

	axisSaveBtn := widget.NewButton("축 설정 저장", func() {
		ahDelay, _ := strconv.ParseFloat(axisAutoHomeDelayEntry.Text, 64)
		mSpeed, _ := strconv.ParseFloat(axisMotionSpeedEntry.Text, 64)

		sm.Mu.RLock()
		oldSet := sm.Data.Axes[currentAxis]
		sm.Mu.RUnlock()

		settingsMap := &settings.AxisSettings{
			Name:           currentAxis,
			FriendlyName:   oldSet.FriendlyName,
			FunscriptNames: oldSet.FunscriptNames,
			Enabled:        axisEnabledCheck.Checked,
			DefaultValue:   oldSet.DefaultValue,
			Min:            axisMinSlider.Value,
			Max:            axisMaxSlider.Value,
			Offset:         axisOffsetSlider.Value,
			Invert:         axisInvertCheck.Checked,
			Interpolation:  axisInterpolationSelect.Selected,
			AutoHome:       axisAutoHomeCheck.Checked,
			AutoHomeDelay:  ahDelay,
			MotionType:     axisMotionTypeSelect.Selected,
			MotionSpeed:    mSpeed,
			MotionRange:    axisMotionRangeSlider.Value,
			MotionOffset:   axisMotionOffsetSlider.Value,
		}

		coord.SetAxisSettings(currentAxis, settingsMap)

		// 전역 오프셋 체크박스 정보도 동기화
		sm.Mu.Lock()
		sm.Data.UseGlobalOffset = axisUseGlobalOffsetCheck.Checked
		if sm.Data.UseGlobalOffset {
			sm.Data.GlobalOffset = axisOffsetSlider.Value
			for _, ax := range sm.Data.Axes {
				ax.Offset = axisOffsetSlider.Value
			}
		}
		sm.Mu.Unlock()
		sm.Save()

		sm.Mu.RLock()
		lang := sm.Data.Language
		sm.Mu.RUnlock()
		dict := langDict[lang]
		if dict == nil {
			dict = langDict["ko"]
		}
		dialog.ShowInformation(dict["save_complete"], fmt.Sprintf(dict["save_axis_success"], currentAxis), myWindow)
		updateAxisFormUI(currentAxis)
	})

	// 축 선택용 세그먼트/버튼 컨테이너
	axesButtonsContainer := container.NewGridWithColumns(6)
	for _, key := range axesKeys {
		btnKey := key
		// 미니 진행률 바와 라벨이 들어간 개별 컴포넌트 생성
		progressBar := widget.NewProgressBarWithData(axisProgressBinds[btnKey])
		progressBar.Min = 0.0
		progressBar.Max = 1.0

		lbl := widget.NewLabelWithData(axisLabelBinds[btnKey])
		lbl.Alignment = fyne.TextAlignCenter

		btn := widget.NewButton(btnKey, func() {
			currentAxis = btnKey
			updateAxisFormUI(btnKey)
		})

		axesButtonsContainer.Add(container.NewVBox(btn, progressBar, lbl))
	}

	autohomeHeaderLabel := widget.NewLabelWithStyle("오토 홈 (Auto Home)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	motionHeaderLabel := widget.NewLabelWithStyle("자체 모션 생성기 (Motion Provider)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	axisMinMaxLabel := widget.NewLabel("동작 한계 설정 (Min/Max Limit):")
	axisOffsetDelayLabel := widget.NewLabel("시간 차이 오프셋 (Offset Delay):")
	axisInterpolationLabel := widget.NewLabel("보간 방식 (Interpolation):")
	axisDelayLabel := widget.NewLabel("딜레이(초):")
	axisMotionTypeLabel := widget.NewLabel("모션 타입:  ")
	axisMotionSpeedLabel := widget.NewLabel("속도 배율:  ")
	axisMotionRangeLabel := widget.NewLabel("진폭 (Range):")
	axisMotionOffsetLabel := widget.NewLabel("기준 (Offset):")

	axisEditorForm := container.NewVBox(
		axisFriendlyNameLabel,
		widget.NewSeparator(),
		axisEnabledCheck,
		container.NewGridWithColumns(2,
			container.NewVBox(axisMinMaxLabel, axisMinSlider, axisMinLabel),
			container.NewVBox(widget.NewLabel(""), axisMaxSlider, axisMaxLabel),
		),
		container.NewGridWithColumns(2,
			container.NewVBox(axisOffsetDelayLabel, axisOffsetSlider, axisOffsetLabel, axisUseGlobalOffsetCheck),
			container.NewVBox(axisInterpolationLabel, axisInterpolationSelect, axisInvertCheck),
		),
		widget.NewSeparator(),
		autohomeHeaderLabel,
		container.NewGridWithColumns(2,
			axisAutoHomeCheck,
			container.NewBorder(nil, nil, axisDelayLabel, nil, axisAutoHomeDelayEntry),
		),
		widget.NewSeparator(),
		motionHeaderLabel,
		container.NewGridWithColumns(2,
			container.NewBorder(nil, nil, axisMotionTypeLabel, nil, axisMotionTypeSelect),
			container.NewBorder(nil, nil, axisMotionSpeedLabel, nil, axisMotionSpeedEntry),
		),
		container.NewGridWithColumns(2,
			container.NewBorder(nil, nil, axisMotionRangeLabel, nil, axisMotionRangeSlider),
			container.NewBorder(nil, nil, axisMotionOffsetLabel, nil, axisMotionOffsetSlider),
		),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, axisScriptInfoLabel, loadScriptManualBtn, axisSaveBtn),
	)

	// 초기 축 데이터 바인딩 로드
	updateAxisFormUI(currentAxis)
	logWrite("5.10. updateAxisFormUI 완료")

	axesTabsScroll := container.NewVScroll(axisEditorForm)
	axesSplit := container.NewVSplit(axesButtonsContainer, axesTabsScroll)
	axesSplit.Offset = 0.22
	logWrite("5.11. axesSplit 완료")

	// ==============================================================
	// [탭 3: 실시간 TCode 모니터 (TCode Terminal)]
	// ==============================================================
	tcodeConsoleEntry := widget.NewMultiLineEntry()
	tcodeConsoleEntry.Bind(tcodeConsoleBind)
	tcodeConsoleEntry.TextStyle = fyne.TextStyle{Monospace: true}

	clearLogBtn := widget.NewButton("로그 지우기", func() {
		tcodeLogMu.Lock()
		tcodeLogBuffer = nil
		tcodeLogMu.Unlock()
		tcodeConsoleBind.Set("")
	})
	tcodeTerminalLayout := container.NewBorder(nil, clearLogBtn, nil, nil, tcodeConsoleEntry)
	logWrite("5.12. tcodeTerminalLayout 완료")

	// ==============================================================
	// [탭 4: 시스템 환경설정 (System Settings)]
	// ==============================================================
	sysPortEntry := widget.NewEntry()
	sm.Mu.RLock()
	sysPortEntry.SetText(sm.Data.WebPort)
	allowRemoteVal := sm.Data.AllowRemoteAccess
	sm.Mu.RUnlock()

	var sysRemoteCheck *widget.Check
	sysRemoteCheck = widget.NewCheck("WebUI 접속 허용 (웹서버 켜기)", func(checked bool) {
		sm.Mu.Lock()
		sm.Data.AllowRemoteAccess = checked
		sm.Mu.Unlock()
		sm.Save()
		
		updateServerURL()
		ws.RebindServer()
	})
	sysRemoteCheck.Checked = allowRemoteVal

	sysSaveBtn := widget.NewButton("설정 저장 및 웹서버 재기동", func() {
		newPort := sysPortEntry.Text
		if newPort == "" {
			newPort = "5000"
		}

		sm.Mu.Lock()
		sm.Data.WebPort = newPort
		sm.Data.AllowRemoteAccess = sysRemoteCheck.Checked
		sm.Mu.Unlock()

		sm.Mu.RLock()
		lang := sm.Data.Language
		sm.Mu.RUnlock()
		dict := langDict[lang]
		if dict == nil {
			dict = langDict["ko"]
		}

		if err := sm.Save(); err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation(dict["save_complete"], dict["sys_save_success"], myWindow)
		}

		updateServerURL()
		ws.RebindServer()
	})

	sysOpenWebBtn := widget.NewButton("수동으로 브라우저(WebUI) 열기", func() {
		sm.Mu.RLock()
		port := sm.Data.WebPort
		sm.Mu.RUnlock()
		openBrowser("http://127.0.0.1:" + port)
	})

	langOptions := []string{"한국어 (Korean)", "English"}
	langCodeMap := map[string]string{
		"한국어 (Korean)": "ko",
		"English":          "en",
	}
	langDisplayMap := map[string]string{
		"ko": "한국어 (Korean)",
		"en": "English",
	}

	sysLangSelect := widget.NewSelect(langOptions, func(selected string) {
		code := langCodeMap[selected]
		sm.Mu.Lock()
		sm.Data.Language = code
		sm.Mu.Unlock()
		sm.Save()
		if applyFyneLanguage != nil {
			applyFyneLanguage(code)
		}
		ws.RebindServer()
	})
	sm.Mu.RLock()
	initialLangCode := sm.Data.Language
	sm.Mu.RUnlock()
	sysLangSelect.SetSelected(langDisplayMap[initialLangCode])

	sysPortLabel := widget.NewLabel("웹 서버 포트 번호(Port):")
	sysLangLabel := widget.NewLabel("언어 설정 (Language):")
	
	systemForm := container.NewVBox(
		sysRemoteCheck,
		widget.NewLabelWithData(serverUrlBind),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			sysPortLabel,
			sysPortEntry,
		),
		sysSaveBtn,
		sysOpenWebBtn,
	)
	systemCard := widget.NewCard("시스템 제어 설정", "", systemForm)

	langForm := container.NewVBox(
		container.NewGridWithColumns(2,
			sysLangLabel,
			sysLangSelect,
		),
	)
	langCard := widget.NewCard("언어 설정 (Language Settings)", "", langForm)

	systemSettingsContainer := container.NewVBox(
		systemCard,
		langCard,
	)

	// 탭 레이아웃 생성
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("미디어 & 출력", theme.MediaPlayIcon(), mainSplit),
		container.NewTabItemWithIcon("축 상세 조정", theme.SettingsIcon(), axesSplit),
		container.NewTabItemWithIcon("TCode 모니터", theme.DocumentIcon(), tcodeTerminalLayout),
		container.NewTabItemWithIcon("시스템 설정", theme.ComputerIcon(), systemSettingsContainer),
	)

	applyFyneLanguage = func(lang string) {
		dict, ok := langDict[lang]
		if !ok {
			dict = langDict["ko"]
		}
		myWindow.SetTitle(dict["title"])
		tabs.Items[0].Text = dict["tab_media"]
		tabs.Items[1].Text = dict["tab_axis"]
		tabs.Items[2].Text = dict["tab_tcode"]
		tabs.Items[3].Text = dict["tab_sys"]
		tabs.Refresh()

		playerHeaderLabel.SetText(dict["hdr_player"])
		scriptDirsHeaderLabel.SetText(dict["hdr_script"])
		outputsHeaderLabel.SetText(dict["hdr_output"])
		autohomeHeaderLabel.SetText(dict["hdr_autohome"])
		motionHeaderLabel.SetText(dict["hdr_motion"])
		systemCard.Title = dict["hdr_sys"]
		systemCard.Refresh()
		langCard.Title = dict["hdr_lang"]
		langCard.Refresh()

		playbackInfoCard.Title = dict["hdr_playback"]
		playbackInfoCard.Refresh()

		// Update sub-widgets texts
		playerSelectLabel.SetText(dict["lbl_player_select"])
		endpointLabel.SetText(dict["lbl_address_pipe"])
		passwordLabel.SetText(dict["lbl_vlc_password"])
		autoConnectCheck.SetText(dict["chk_autoconnect"])
		playerConnectBtn.SetText(dict["btn_connect"])
		
		addPathEntry.SetPlaceHolder(dict["ph_script_dir"])
		addPathBtn.SetText(dict["btn_add"])
		
		axisEnabledCheck.SetText(dict["chk_axis_enabled"])
		axisMinMaxLabel.SetText(dict["lbl_axis_minmax"])
		axisOffsetDelayLabel.SetText(dict["lbl_axis_offset"])
		axisUseGlobalOffsetCheck.SetText(dict["chk_global_offset"])
		axisInterpolationLabel.SetText(dict["lbl_axis_interpolation"])
		axisInvertCheck.SetText(dict["chk_axis_invert"])
		axisAutoHomeCheck.SetText(dict["chk_autohome_enabled"])
		axisDelayLabel.SetText(dict["lbl_delay_sec"])
		axisMotionTypeLabel.SetText(dict["lbl_motion_type"])
		axisMotionSpeedLabel.SetText(dict["lbl_motion_speed"])
		axisMotionRangeLabel.SetText(dict["lbl_motion_range"])
		axisMotionOffsetLabel.SetText(dict["lbl_motion_offset"])
		loadScriptManualBtn.SetText(dict["btn_manual_script"])
		axisSaveBtn.SetText(dict["btn_save_axis"])
		
		clearLogBtn.SetText(dict["btn_clear_log"])
		
		sysRemoteCheck.SetText(dict["chk_remote_access"])
		sysPortLabel.SetText(dict["lbl_web_port"])
		sysLangLabel.SetText(dict["lbl_language"])
		sysSaveBtn.SetText(dict["btn_save_restart"])
		sysOpenWebBtn.SetText(dict["btn_open_web"])

		// Also update current active axis description
		updateAxisFormUI(currentAxis)
		// Force server URL bind text refresh
		updateServerURL()
	}

	sm.Mu.RLock()
	initialLang := sm.Data.Language
	sm.Mu.RUnlock()
	applyFyneLanguage(initialLang)

	myWindow.SetContent(tabs)
	logWrite("5.15. myWindow.SetContent 완료")

	// GUI State Polling Ticker (0.1s intervals) - 데이터 바인딩을 이용해 스레드 안전하게 리프레시
	stopChan := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logWrite(fmt.Sprintf("TICKER LOOP PANIC: %v\nSTACK TRACE:\n%s", r, string(debug.Stack())))
			}
		}()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastVideo string
		var lastLoadedStates = make(map[string]bool)

		for {
			select {
			case <-ticker.C:
				// 1. Playback Info
				pos, dur, playing := coord.GetPlaybackInfo()
				mediaDuration = dur
				mediaPosition = pos

				// Format time label
				videoTimeBind.Set(fmt.Sprintf("%s / %s", formatTime(pos), formatTime(dur)))

				// Manage keyframes cache update safely
				currVideo := coord.GetCurrentVideoTitle()
				currLoadedScripts := coord.GetAxisScriptLoadedStates()

				stateChanged := currVideo != lastVideo
				if !stateChanged {
					for _, key := range axesKeys {
						if currLoadedScripts[key] != lastLoadedStates[key] {
							stateChanged = true
							break
						}
					}
				}

				if stateChanged {
					lastVideo = currVideo
					lastLoadedStates = make(map[string]bool)
					for k, v := range currLoadedScripts {
						lastLoadedStates[k] = v
					}

					fyne.Do(func() {
						axisHeatmaps = make(map[string][]funscript.Keyframe)
						for _, key := range axesKeys {
							if currLoadedScripts[key] {
								axisHeatmaps[key] = coord.GetScriptKeyframes(key)
							}
						}
						// Reset background image cache to trigger redraw on state change
						heatmapBgMu.Lock()
						heatmapBgImage = nil
						heatmapBgMu.Unlock()
					})
				}

				fyne.Do(func() {
					// Set slider value
					if dur > 0 {
						playbackSlider.Max = dur
						if pos <= dur {
							playbackSlider.SetValue(pos)
						}
					} else {
						playbackSlider.Max = 100
						playbackSlider.SetValue(0)
					}

					// UX: 재생 상태에 따라 재생/정지 버튼 활성화/비활성화 처리
					if playing {
						playBtn.Disable()
						pauseBtn.Enable()
					} else {
						playBtn.Enable()
						pauseBtn.Disable()
					}

					heatmapRaster.Refresh()
				})

				sm.Mu.RLock()
				lang := sm.Data.Language
				sm.Mu.RUnlock()
				dict := langDict[lang]
				if dict == nil {
					dict = langDict["ko"]
				}

				// 2. Playback state card (Status / Video Name)
				pName := coord.GetActivePlayerName()
				status := coord.GetActivePlayerStatus()
				video := coord.GetCurrentVideoTitle()

				statusText := dict["status_offline"]
				if status == "connected" {
					statusText = dict["status_connected"]
				} else if status == "connecting" {
					statusText = dict["status_connecting"]
				}
				playerStatusBind.Set(fmt.Sprintf(dict["status_player_prefix"]+"%s (%s)", strings.ToUpper(pName), statusText))

				if video == "" || video == "No video loaded" {
					videoTitleBind.Set(dict["video_none"])
				} else {
					videoTitleBind.Set(dict["video_prefix"] + video)
				}

				// 3. Mini Progress Bars & Labels update
				actualValues := coord.GetAxisActualValues()
				loadedScripts := coord.GetAxisScriptLoadedStates()

				for _, key := range axesKeys {
					val, ok := actualValues[key]
					if ok {
						axisProgressBinds[key].Set(val)
						axisLabelBinds[key].Set(fmt.Sprintf("%.2f", val))
					}
				}

				// Current Selected Axis Script Info 갱신
				currentScriptLoaded, _ := loadedScripts[currentAxis]
				fyne.Do(func() {
					if currentScriptLoaded {
						axisScriptInfoLabel.SetText(dict["script_loaded_yes"])
					} else {
						axisScriptInfoLabel.SetText(dict["script_loaded_no"])
					}
				})

				// 4. Outputs switch checkbox sync
				sm.Mu.RLock()
				type chkState struct {
					key     string
					enabled bool
				}
				var chkStates []chkState
				for _, k := range []string{"udp", "tcp", "websocket", "serial", "file"} {
					chkStates = append(chkStates, chkState{key: k, enabled: sm.Data.OutputTargets[k].Enabled})
				}
				sm.Mu.RUnlock()

				fyne.Do(func() {
					for _, state := range chkStates {
						chk := outputCheckboxes[state.key]
						if chk != nil && chk.Checked != state.enabled {
							chk.Checked = state.enabled
							chk.Refresh()
						}
					}
				})

				// 5. TCode Terminal batch logs render
				tcodeLogMu.Lock()
				var logLines string
				hasLogs := len(tcodeLogBuffer) > 0
				if hasLogs {
					logLines = strings.Join(tcodeLogBuffer, "")
				}
				tcodeLogMu.Unlock()

				if hasLogs {
					tcodeConsoleBind.Set(logLines)
				}

			case <-stopChan:
				return
			}
		}
	}()

	// 6. Block Window loop until closed
	logWrite("6. myWindow.ShowAndRun() 호출 직전")
	myWindow.ShowAndRun()
	logWrite("myWindow.ShowAndRun() 리턴 (창 닫힘)")

	// Clean up resources on exit
	close(stopChan)
	logWrite("프로그램 종료 처리 시작 (리소스 정리)")
	coord.Stop()
	logWrite("종료 완료 (main() 정상 리턴)")
}

func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	var ip192, ip10, ipOther []string

	for _, iface := range interfaces {
		nameLower := strings.ToLower(iface.Name)
		// Exclude interfaces related to NordLynx, VPN, virtual, loopback, etc.
		if strings.Contains(nameLower, "vpn") || strings.Contains(nameLower, "nordlynx") || 
		   strings.Contains(nameLower, "virtual") || strings.Contains(nameLower, "pseudo") ||
		   strings.Contains(nameLower, "loopback") || (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}

			ipStr := ip.String()
			if strings.HasPrefix(ipStr, "192.168.") {
				ip192 = append(ip192, ipStr)
			} else if strings.HasPrefix(ipStr, "10.") {
				ip10 = append(ip10, ipStr)
			} else {
				ipOther = append(ipOther, ipStr)
			}
		}
	}

	// Prioritize: 192.168.x.x first, then 10.x.x.x, then other IPv4s
	if len(ip192) > 0 {
		return ip192[0]
	}
	if len(ip10) > 0 {
		return ip10[0]
	}
	if len(ipOther) > 0 {
		return ipOther[0]
	}
	return "127.0.0.1"
}

func formatTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	m := int(seconds / 60)
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// Bresenham's line algorithm for draw raster
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawLineAA(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0

	absDx := abs(dx)
	absDy := abs(dy)

	if absDx == 0 && absDy == 0 {
		drawPixelBlend(img, x0, y0, col, 1.0)
		return
	}

	if absDx > absDy {
		if x0 > x1 {
			x0, x1 = x1, x0
			y0, y1 = y1, y0
		}
		gradient := float64(dy) / float64(dx)
		y := float64(y0)
		for x := x0; x <= x1; x++ {
			drawPixelBlend(img, x, int(y), col, 1.0-(y-float64(int(y))))
			drawPixelBlend(img, x, int(y)+1, col, y-float64(int(y)))
			y += gradient
		}
	} else {
		if y0 > y1 {
			x0, x1 = x1, x0
			y0, y1 = y1, y0
		}
		gradient := float64(dx) / float64(dy)
		x := float64(x0)
		for y := y0; y <= y1; y++ {
			drawPixelBlend(img, int(x), y, col, 1.0-(x-float64(int(x))))
			drawPixelBlend(img, int(x)+1, y, col, x-float64(int(x)))
			x += gradient
		}
	}
}

func drawPixelBlend(img *image.RGBA, x, y int, col color.RGBA, alphaFactor float64) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}

	cur := img.RGBAAt(x, y)
	alpha := (float64(col.A) / 255.0) * alphaFactor

	r := uint8(float64(col.R)*alpha + float64(cur.R)*(1.0-alpha))
	g := uint8(float64(col.G)*alpha + float64(cur.G)*(1.0-alpha))
	b := uint8(float64(col.B)*alpha + float64(cur.B)*(1.0-alpha))

	img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
}

func drawCircle(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	for y := cy - r - 2; y <= cy+r+2; y++ {
		for x := cx - r - 2; x <= cx+r+2; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(r) {
				drawPixelBlend(img, x, y, col, 1.0)
			} else if dist < float64(r)+1.0 {
				drawPixelBlend(img, x, y, col, 1.0-(dist-float64(r)))
			}
		}
	}
}

func drawGlowCircle(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	drawCircle(img, cx, cy, r, col)

	glowCol := col
	glowCol.A = 80
	for y := cy - r - 6; y <= cy+r+6; y++ {
		for x := cx - r - 6; x <= cx+r+6; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > float64(r) && dist < float64(r)+6.0 {
				factor := 1.0 - ((dist - float64(r)) / 6.0)
				drawPixelBlend(img, x, y, glowCol, factor)
			}
		}
	}
}

func interpolateValue(timeVal float64, keyframes []funscript.Keyframe) float64 {
	if len(keyframes) == 0 {
		return 0.5
	}
	timeMs := timeVal * 1000.0

	if timeMs <= keyframes[0].At {
		return keyframes[0].Pos
	}
	if timeMs >= keyframes[len(keyframes)-1].At {
		return keyframes[len(keyframes)-1].Pos
	}

	for i := 0; i < len(keyframes)-1; i++ {
		p0 := keyframes[i]
		p1 := keyframes[i+1]
		if timeMs >= p0.At && timeMs <= p1.At {
			t := (timeMs - p0.At) / (p1.At - p0.At)
			return p0.Pos + (p1.Pos-p0.Pos)*t
		}
	}
	return 0.5
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		if err != nil {
			exec.Command("cmd", "/c", "start", url).Start()
		}
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Printf("[UI] 브라우저 자동 실행 실패 (수동 접속 필요): %s\n", url)
	}
}

func showNativeMessageBox(title, message string) {
	if runtime.GOOS != "windows" {
		fmt.Printf("[%s] %s\n", title, message)
		return
	}
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	messageBoxW.Call(0, uintptr(unsafe.Pointer(messagePtr)), uintptr(unsafe.Pointer(titlePtr)), 0x10)
}
