// Global State
let ws = null;
let appSettings = null;
let currentAxis = "L0";
let activeState = null;
let scriptDataCache = {}; // axis -> keyframes list

const translations = {
    ko: {
        header_remote: "원격접속 허용",
        player_offline: "Player: Offline",
        output_offline: "Output: Offline",
        panel_media_title: "미디어 및 스크립트",
        player_card_title: "플레이어 연동",
        player_select_internal: "내부 재생기 (Internal)",
        player_select_mpchc: "MPC-HC",
        player_select_mpv: "MPV (Named Pipe)",
        player_select_vlc: "VLC",
        player_select_deovr: "DeoVR",
        player_select_heresphere: "HereSphere",
        connect: "연결",
        endpoint_label: "주소 / 파이프 이름",
        password_label: "비밀번호 (VLC 전용)",
        autoconnect_label: "시작 시 자동 연결",
        save_tooltip: "저장",
        script_library_title: "스크립트 라이브러리 경로 설정",
        script_library_placeholder: "예: C:\\funscripts",
        add: "추가",
        no_registered_paths: "등록된 경로 없음",
        delete: "삭제",
        no_video: "재생 중인 비디오 없음",
        no_script: "스크립트 없음",
        panel_axes_title: "축 제어 및 조정 (Device Axes)",
        axis_enabled: "활성화",
        axis_limit_label: "최소 / 최대 범위 (Limit)",
        axis_offset_label: "오프셋 (Offset)",
        use_global_offset: "전역 오프셋 사용",
        axis_interpolation_label: "보간 방식",
        axis_interpolation_linear: "Linear (선형)",
        axis_interpolation_pchip: "Pchip (부드러운 곡선)",
        axis_interpolation_makima: "Makima (정밀 곡선)",
        axis_interpolation_step: "Step (계단형)",
        axis_invert_label: "출력 반전 (Invert)",
        auto_home_title: "오토 홈 (Auto Home)",
        auto_home_enabled: "오토 홈 활성화",
        auto_home_delay_label: "복귀 딜레이 (초)",
        motion_provider_title: "자체 모션 생성기 (Motion Provider)",
        motion_type_label: "모션 타입",
        motion_type_none: "사용 안함 (None)",
        motion_speed_label: "속도배율 (Speed)",
        motion_range_label: "진폭 (Range)",
        motion_offset_label: "기준점 (Offset)",
        axis_save_btn: "축 설정 저장",
        panel_outputs_title: "출력 장치 연동 및 로그",
        output_list_title: "출력 채널 관리",
        terminal_title: "TCode 실시간 출력 모니터",
        clear: "지우기",
        config: "설정",
        loaded_script_prefix: "로드된 스크립트: ",
        loaded_script_none: "로드된 스크립트: 없음",
        loading_script: "업로드 중...",
        upload_failed: "업로드 실패",
        already_registered: "이미 등록된 경로입니다.",
        modal_output_config_title: "출력 채널 설정",
        modal_endpoint_address: "연결 주소 (IP:PORT)",
        modal_endpoint_serial: "시리얼 포트 직접 입력 (e.g. COM3)",
        modal_baudrate_label: "보오레이트 (Baud Rate)",
        modal_filepath_label: "저장 파일 경로",
        modal_cancel: "취소",
        modal_save: "저장",
        folder_modal_title: "폴더 찾아보기",
        folder_modal_drive: "드라이브 선택",
        folder_modal_parent: "상위",
        folder_modal_select: "이 폴더 선택",
        folder_modal_no_sub: "하위 폴더 없음",
        folder_modal_loading: "로딩 중...",
        folder_modal_fail: "경로를 읽어오는 데 실패했습니다.",
        serial_scanned_label: "시스템 장치 목록",
        serial_scanning: "포트를 조회하는 중...",
        serial_scan_select: "포트 선택 (스캔 결과)",
        serial_scan_none: "사용 가능한 포트 없음",
        serial_scan_fail: "포트 조회 실패",
        dropzone_desc: "여기에 .funscript 또는 .csv 파일을 드래그 앤 드롭 하거나 클릭하여 로드하세요."
    },
    en: {
        header_remote: "Allow Remote Connection",
        player_offline: "Player: Offline",
        output_offline: "Output: Offline",
        panel_media_title: "Media & Scripts",
        player_card_title: "Player Connection",
        player_select_internal: "Internal Player",
        player_select_mpchc: "MPC-HC",
        player_select_mpv: "MPV (Named Pipe)",
        player_select_vlc: "VLC",
        player_select_deovr: "DeoVR",
        player_select_heresphere: "HereSphere",
        connect: "Connect",
        endpoint_label: "Address / Pipe Name",
        password_label: "Password (VLC Only)",
        autoconnect_label: "Auto Connect on Start",
        save_tooltip: "Save",
        script_library_title: "Script Library Paths",
        script_library_placeholder: "e.g. C:\\funscripts",
        add: "Add",
        no_registered_paths: "No registered paths",
        delete: "Delete",
        no_video: "No video playing",
        no_script: "No script loaded",
        panel_axes_title: "Axis Control & Tuning",
        axis_enabled: "Enabled",
        axis_limit_label: "Min / Max Limits",
        axis_offset_label: "Offset",
        use_global_offset: "Use Global Offset",
        axis_interpolation_label: "Interpolation Type",
        axis_interpolation_linear: "Linear",
        axis_interpolation_pchip: "Pchip",
        axis_interpolation_makima: "Makima",
        axis_interpolation_step: "Step",
        axis_invert_label: "Invert Output",
        auto_home_title: "Auto Home",
        auto_home_enabled: "Enable Auto Home",
        auto_home_delay_label: "Home Delay (sec)",
        motion_provider_title: "Motion Provider",
        motion_type_label: "Motion Type",
        motion_type_none: "None",
        motion_speed_label: "Speed Multiplier",
        motion_range_label: "Range",
        motion_offset_label: "Center Offset",
        axis_save_btn: "Save Axis Settings",
        panel_outputs_title: "Output Devices & Logs",
        output_list_title: "Output Channels",
        terminal_title: "Live TCode Console",
        clear: "Clear",
        config: "Config",
        loaded_script_prefix: "Loaded script: ",
        loaded_script_none: "Loaded script: None",
        loading_script: "Uploading...",
        upload_failed: "Upload failed",
        already_registered: "Path already registered.",
        modal_output_config_title: "Configure Output Channel",
        modal_endpoint_address: "Connection Address (IP:PORT)",
        modal_endpoint_serial: "Serial Port Input (e.g. COM3)",
        modal_baudrate_label: "Baud Rate",
        modal_filepath_label: "Save File Path",
        modal_cancel: "Cancel",
        modal_save: "Save",
        folder_modal_title: "Browse Folders",
        folder_modal_drive: "Select Drive",
        folder_modal_parent: "Up",
        folder_modal_select: "Select This Folder",
        folder_modal_no_sub: "No subdirectories",
        folder_modal_loading: "Loading...",
        folder_modal_fail: "Failed to read path.",
        serial_scanned_label: "System Device List",
        serial_scanning: "Scanning ports...",
        serial_scan_select: "Select Port (Scan Result)",
        serial_scan_none: "No ports found",
        serial_scan_fail: "Scan failed",
        dropzone_desc: "Drag & drop a .funscript or .csv file here, or click to browse."
    }
};

// Default axis names if not loaded from settings yet
const fallbackAxes = ["L0", "L1", "L2", "R0", "R1", "R2", "V0", "V1", "A0", "A1", "A2"];

// Initialize UI Elements
document.addEventListener("DOMContentLoaded", () => {
    initWebSocket();
    setupEventListeners();
    renderAxisTabs(fallbackAxes);
});

// Setup WebSocket Connection
function initWebSocket() {
    const wsUri = `ws://${window.location.host}/ws`;
    ws = new WebSocket(wsUri);

    ws.onopen = () => {
        console.log("WebSocket connected");
        document.getElementById("player-indicator").querySelector(".status-dot").className = "status-dot connecting";
        document.getElementById("output-indicator").querySelector(".status-dot").className = "status-dot connecting";
    };

    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.type === "settings") {
            appSettings = msg.value;
            updateSettingsUI();
        } else if (msg.type === "state") {
            activeState = msg.value;
            updateStateUI();
        } else if (msg.type === "tcode") {
            appendTCode(msg.value);
        }
    };

    ws.onclose = () => {
        console.log("WebSocket disconnected, retrying...");
        document.getElementById("player-indicator").querySelector(".status-dot").className = "status-dot disconnected";
        document.getElementById("output-indicator").querySelector(".status-dot").className = "status-dot disconnected";
        setTimeout(initWebSocket, 2000);
    };
}

// Render Axis Tab buttons
function renderAxisTabs(axesKeys) {
    const list = document.getElementById("axis-tab-list");
    list.innerHTML = "";

    axesKeys.forEach(key => {
        const tab = document.createElement("div");
        tab.className = `axis-tab ${key === currentAxis ? 'active' : ''}`;
        tab.id = `tab-${key}`;
        tab.innerHTML = `
            <span class="axis-tab-name">${key}</span>
            <div class="axis-tab-progress-container">
                <div class="axis-tab-progress-bar" id="tab-bar-${key}"></div>
            </div>
            <span class="axis-tab-lbl" id="tab-lbl-${key}">off</span>
            <span class="axis-tab-status"></span>
        `;
        tab.addEventListener("click", () => selectAxis(key));
        list.appendChild(tab);
    });
}

// Select an axis to edit in settings card
function selectAxis(key) {
    currentAxis = key;
    document.querySelectorAll(".axis-tab").forEach(el => el.classList.remove("active"));
    document.getElementById(`tab-${key}`).classList.add("active");

    if (!appSettings || !appSettings.axes[key]) return;

    const axisSet = appSettings.axes[key];
    document.getElementById("axis-editor-badge").innerText = key;
    document.getElementById("axis-editor-name").innerText = axisSet.friendly_name;
    document.getElementById("axis-enabled").checked = axisSet.enabled;
    document.getElementById("axis-min").value = axisSet.min;
    document.getElementById("axis-max").value = axisSet.max;
    document.getElementById("axis-min-lbl").innerText = axisSet.min.toFixed(2);
    document.getElementById("axis-max-lbl").innerText = axisSet.max.toFixed(2);
    document.getElementById("axis-offset").value = axisSet.offset;
    document.getElementById("axis-offset-lbl").innerText = axisSet.offset.toFixed(2);

    const axisOffsetSlider = document.getElementById("axis-offset");
    axisOffsetSlider.disabled = false;
    axisOffsetSlider.style.opacity = "1.0";

    // Update global offset checkbox state in current tab
    document.getElementById("use-global-offset").checked = appSettings.use_global_offset;

    document.getElementById("axis-interpolation").value = axisSet.interpolation;
    document.getElementById("axis-invert").checked = axisSet.invert;

    document.getElementById("axis-autohome").checked = axisSet.auto_home;
    document.getElementById("axis-autohomedelay").value = axisSet.auto_home_delay;

    document.getElementById("axis-motion-type").value = axisSet.motion_type;
    document.getElementById("axis-motion-speed").value = axisSet.motion_speed;
    document.getElementById("axis-motion-range").value = axisSet.motion_range;
    document.getElementById("axis-motion-offset").value = axisSet.motion_offset;

    // Update loaded file text
    const dropzone = document.getElementById("script-dropzone");
    const scriptInfo = document.getElementById("loaded-script-info");
    const lang = appSettings ? appSettings.language || "ko" : "ko";
    const dict = translations[lang];
    if (activeState && activeState.axes[key] && activeState.axes[key].script_loaded) {
        scriptInfo.innerText = `${dict.loaded_script_prefix}${activeState.axes[key].script_name}`;
        scriptInfo.style.color = "var(--success)";
    } else {
        scriptInfo.innerText = dict.loaded_script_none;
        scriptInfo.style.color = "var(--text-secondary)";
    }
}

// Update UI configuration based on received Settings
function updateSettingsUI() {
    if (!appSettings) return;

    // Render tabs list if count changes
    const order = ["L0", "L1", "L2", "R0", "R1", "R2", "V0", "V1", "A0", "A1", "A2"];
    const axesKeys = Object.keys(appSettings.axes).sort((a, b) => {
        const idxA = order.indexOf(a);
        const idxB = order.indexOf(b);
        return (idxA === -1 ? 99 : idxA) - (idxB === -1 ? 99 : idxB);
    });
    renderAxisTabs(axesKeys);

    // Refresh active axis editor values
    selectAxis(currentAxis);

    // Update language selection
    const lang = appSettings.language || "ko";
    if (document.getElementById("header-lang-select")) {
        document.getElementById("header-lang-select").value = lang;
    }
    applyLanguage(lang);

    // Update global offset settings toggle
    document.getElementById("use-global-offset").checked = appSettings.use_global_offset;

    // Update player settings
    document.getElementById("player-select").value = appSettings.active_player;
    const playerConf = appSettings.media_sources[appSettings.active_player];
    if (playerConf) {
        document.getElementById("player-endpoint").value = playerConf.endpoint;
        document.getElementById("player-password").value = playerConf.password || "";
        document.getElementById("player-autoconnect").checked = playerConf.auto_connect;
    }

    // Toggle endpoint input visibility for internal player
    if (appSettings.active_player === "internal") {
        document.getElementById("player-endpoint-group").style.display = "none";
        document.getElementById("player-password-group").style.display = "none";
    } else if (appSettings.active_player === "vlc") {
        document.getElementById("player-endpoint-group").style.display = "flex";
        document.getElementById("player-password-group").style.display = "flex";
    } else {
        document.getElementById("player-endpoint-group").style.display = "flex";
        document.getElementById("player-password-group").style.display = "none";
    }

    // Update outputs settings list
    updateOutputsListUI();



    // Update script directories list
    updateScriptDirectoriesUI();
}

function updateOutputsListUI() {
    if (!appSettings) return;
    const outputs = appSettings.output_targets;
    const container = document.getElementById("outputs-list-container");
    container.innerHTML = "";

    const lang = appSettings.language || "ko";
    const dict = translations[lang];

    Object.keys(outputs).forEach(key => {
        const out = outputs[key];
        let address = out.endpoint;
        if (key === "file") {
            address = out.file_path;
        } else if (key === "serial") {
            address = `${out.endpoint} (${out.baud_rate}bps)`;
        }

        const row = document.createElement("div");
        row.className = "output-row";
        row.innerHTML = `
            <div class="output-label">
                <h4>${key.toUpperCase()}</h4>
                <span>${address}</span>
            </div>
            <div class="toggle-row">
                <button class="btn btn-sm btn-outline" onclick="openOutputConfig('${key}')">${dict.config}</button>
                <input type="checkbox" id="out-${key}-status" class="toggle-switch" ${out.enabled ? 'checked' : ''} onchange="toggleOutput('${key}', this.checked)">
            </div>
        `;
        container.appendChild(row);
    });
}

function updateScriptDirectoriesUI() {
    if (!appSettings) return;
    const container = document.getElementById("script-dirs-list-container");
    container.innerHTML = "";

    const lang = appSettings.language || "ko";
    const dict = translations[lang];

    const dirs = appSettings.script_directories || [];
    if (dirs.length === 0) {
        container.innerHTML = `<span style="font-size: 13px; color: var(--text-secondary); text-align: center; display: block; padding: 10px 0;">${dict.no_registered_paths}</span>`;
        return;
    }

    dirs.forEach((dir, idx) => {
        const row = document.createElement("div");
        row.className = "dir-row";
        row.style = "display: flex; justify-content: space-between; align-items: center; background: rgba(255,255,255,0.05); padding: 6px 10px; border-radius: 6px; font-size: 13px; gap: 8px;";
        row.innerHTML = `
            <span style="word-break: break-all; color: var(--text-primary); text-align: left; flex: 1; min-width: 0;">${dir}</span>
            <button class="btn btn-xs btn-outline" style="color: var(--danger); border-color: var(--danger); flex-shrink: 0;" onclick="removeScriptDirectory(${idx})">${dict.delete}</button>
        `;
        container.appendChild(row);
    });
}

window.removeScriptDirectory = function(idx) {
    if (!appSettings) return;
    const dirs = [...(appSettings.script_directories || [])];
    dirs.splice(idx, 1);
    ws.send(JSON.stringify({
        action: "save_script_directories",
        directories: dirs
    }));
};

// Update state UI based on received real-time variables
let lastVideoPath = "";
function updateStateUI() {
    if (!activeState) return;

    const lang = appSettings ? appSettings.language || "ko" : "ko";
    const dict = translations[lang];

    // Update indicators
    const playerDot = document.getElementById("player-indicator").querySelector(".status-dot");
    const playerLbl = document.getElementById("player-indicator").querySelector(".status-label");
    playerDot.className = `status-dot ${activeState.player_status}`;
    
    const statusText = activeState.player_status === "connected" 
        ? (lang === "ko" ? "연결됨" : "Connected")
        : (activeState.player_status === "connecting" ? (lang === "ko" ? "연결 중" : "Connecting") : (lang === "ko" ? "오프라인" : "Offline"));
    playerLbl.innerText = `Player: ${activeState.player_name.toUpperCase()} (${statusText})`;

    // Check if outputs changed connected status
    let outputConnected = false;
    Object.keys(activeState.outputs).forEach(key => {
        const status = activeState.outputs[key];
        if (status === "connected") {
            outputConnected = true;
        }
    });

    const outputDot = document.getElementById("output-indicator").querySelector(".status-dot");
    const outputLbl = document.getElementById("output-indicator").querySelector(".status-label");
    if (outputConnected) {
        outputDot.className = "status-dot connected";
        outputLbl.innerText = lang === "ko" ? "출력 장치: 연결됨" : "Output: Connected";
    } else {
        outputDot.className = "status-dot disconnected";
        outputLbl.innerText = dict.output_offline;
    }

    // Playback card
    document.getElementById("video-title").innerText = activeState.video_path ? activeState.video_title : dict.no_video;
    document.getElementById("video-path").innerText = activeState.video_path || "-";

    const pos = activeState.position;
    const dur = activeState.duration;
    document.getElementById("heatmap-time").innerText = `${formatTime(pos)} / ${formatTime(dur)}`;

    const progressSlider = document.getElementById("progress-slider");
    if (dur > 0) {
        progressSlider.max = dur;
        progressSlider.value = pos;
    } else {
        progressSlider.max = 100;
        progressSlider.value = 0;
    }

    // Active axis editor preview values
    if (activeState.axes[currentAxis]) {
        const curAx = activeState.axes[currentAxis];
        document.getElementById("axis-bar-fill").style.width = `${curAx.actual_value * 100}%`;
        document.getElementById("axis-bar-marker").style.left = `${curAx.value * 100}%`;
        document.getElementById("axis-val-computed").innerText = curAx.value.toFixed(3);
        document.getElementById("axis-val-actual").innerText = curAx.actual_value.toFixed(3);

        // Update loaded script file text dynamically
        const scriptInfo = document.getElementById("loaded-script-info");
        if (scriptInfo) {
            const lang = appSettings ? appSettings.language || "ko" : "ko";
            const dict = translations[lang];
            if (curAx.script_loaded) {
                scriptInfo.innerText = `${dict.loaded_script_prefix}${curAx.script_name}`;
                scriptInfo.style.color = "var(--success)";
            } else {
                scriptInfo.innerText = dict.loaded_script_none;
                scriptInfo.style.color = "var(--text-secondary)";
            }
        }
    }

    // Update tab labels
    Object.keys(activeState.axes).forEach(key => {
        const ax = activeState.axes[key];
        const lbl = document.getElementById(`tab-lbl-${key}`);
        if (lbl) {
            lbl.innerText = ax.actual_value.toFixed(2);
        }

        const bar = document.getElementById(`tab-bar-${key}`);
        if (bar) {
            bar.style.width = `${ax.actual_value * 100}%`;
        }

        const tab = document.getElementById(`tab-${key}`);
        if (tab) {
            tab.classList.toggle("has-script", ax.script_loaded);
            tab.classList.toggle("has-motion", ax.motion_active);
        }
    });

    // Auto-fetch script keyframes for heatmap if video changed or script loaded
    if (activeState.video_path !== lastVideoPath) {
        lastVideoPath = activeState.video_path;
        scriptDataCache = {};
    }
    fetchHeatmapData();

    drawHeatmap();
}

// Fetch simplified script data for active axes (heatmap)
function fetchHeatmapData() {
    if (!activeState || activeState.video_path === "") {
        scriptDataCache = {};
        return;
    }

    const loadedAxes = Object.keys(activeState.axes).filter(key => activeState.axes[key].script_loaded);

    // Remove axes that are no longer loaded
    Object.keys(scriptDataCache).forEach(key => {
        if (!loadedAxes.includes(key)) {
            delete scriptDataCache[key];
        }
    });

    loadedAxes.forEach(axis => {
        if (scriptDataCache[axis]) return;

        // Prevent duplicate simultaneous fetches by assigning an empty array
        scriptDataCache[axis] = [];

        fetch(`/api/script?axis=${axis}`)
            .then(res => {
                if (res.ok) return res.json();
                throw new Error();
            })
            .then(data => {
                scriptDataCache[axis] = data.keyframes || [];
            })
            .catch(() => {
                scriptDataCache[axis] = [];
            });
    });
}

// Draw funscript heatmap on HTML5 Canvas
function drawHeatmap() {
    const canvas = document.getElementById("heatmap-canvas");
    if (!canvas) return;
    const ctx = canvas.getContext("2d");

    // Fit canvas to element size
    const rect = canvas.getBoundingClientRect();
    if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
    }

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const hasAnyKfs = Object.values(scriptDataCache).some(kfs => kfs && kfs.length > 0);

    if (!hasAnyKfs || !activeState || activeState.duration <= 0) {
        // Draw empty indicator
        ctx.fillStyle = "rgba(255,255,255,0.05)";
        ctx.font = "12px sans-serif";
        ctx.textAlign = "center";
        const lang = appSettings ? appSettings.language || "ko" : "ko";
        const dict = translations[lang];
        ctx.fillText(dict ? dict.no_script : "스크립트 없음", canvas.width / 2, canvas.height / 2 + 4);
        return;
    }

    const dur = activeState.duration;

    const axisColors = {
        "L0": { color: "rgba(0, 242, 254, 1.0)", width: 2.0 },
        "L1": { color: "rgba(76, 175, 80, 0.3)", width: 1.5 },
        "L2": { color: "rgba(205, 220, 57, 0.3)", width: 1.5 },
        "R0": { color: "rgba(244, 67, 54, 0.3)", width: 1.5 },
        "R1": { color: "rgba(255, 152, 0, 0.3)", width: 1.5 },
        "R2": { color: "rgba(233, 30, 99, 0.3)", width: 1.5 },
        "V0": { color: "rgba(156, 39, 176, 0.3)", width: 1.5 },
        "V1": { color: "rgba(63, 81, 181, 0.3)", width: 1.5 },
        "A0": { color: "rgba(0, 150, 136, 0.3)", width: 1.5 },
        "A1": { color: "rgba(255, 193, 7, 0.3)", width: 1.5 },
        "A2": { color: "rgba(255, 87, 34, 0.3)", width: 1.5 },
    };

    // 1. Draw L0 curve first (background, prominent)
    const l0Kfs = scriptDataCache["L0"];
    if (l0Kfs && l0Kfs.length > 0) {
        const style = axisColors["L0"];
        ctx.strokeStyle = style.color;
        ctx.lineWidth = style.width;
        ctx.beginPath();

        l0Kfs.forEach((kf, idx) => {
            const x = (kf.At / dur) * canvas.width;
            const y = canvas.height - (kf.Pos * canvas.height);
            if (idx === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        ctx.stroke();
    }

    // 2. Draw non-L0 curves last (foreground, transparent)
    Object.keys(scriptDataCache).forEach(axis => {
        if (axis === "L0") return;
        const kfs = scriptDataCache[axis];
        if (!kfs || kfs.length === 0) return;

        const style = axisColors[axis] || { color: "rgba(255, 255, 255, 0.3)", width: 1.5 };
        ctx.strokeStyle = style.color;
        ctx.lineWidth = style.width;
        ctx.beginPath();

        kfs.forEach((kf, idx) => {
            const x = (kf.At / dur) * canvas.width;
            const y = canvas.height - (kf.Pos * canvas.height);
            if (idx === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        ctx.stroke();
    });

    // Draw current position vertical lines
    const currentX = (activeState.position / dur) * canvas.width;
    ctx.strokeStyle = "#ff1744";
    ctx.lineWidth = 1.5;
    ctx.beginPath();
    ctx.moveTo(currentX, 0);
    ctx.lineTo(currentX, canvas.height);
    ctx.stroke();

    // Draw glow circle at current point for L0
    if (l0Kfs && l0Kfs.length > 0) {
        const curVal = interpolateValue(activeState.position, l0Kfs);
        const curY = canvas.height - (curVal * canvas.height);

        ctx.fillStyle = "#ff1744";
        ctx.shadowColor = "#ff1744";
        ctx.shadowBlur = 8;
        ctx.beginPath();
        ctx.arc(currentX, curY, 4, 0, Math.PI * 2);
        ctx.fill();
        ctx.shadowBlur = 0; // reset
    }
}

// Simple client-side linear interpolation for cursor position marker
function interpolateValue(time, keyframes) {
    if (keyframes.length === 0) return 0.5;
    if (time <= keyframes[0].At) return keyframes[0].Pos;
    if (time >= keyframes[keyframes.length-1].At) return keyframes[keyframes.length-1].Pos;

    for (let i = 0; i < keyframes.length - 1; i++) {
        const p0 = keyframes[i];
        const p1 = keyframes[i+1];
        if (time >= p0.At && time <= p1.At) {
            const t = (time - p0.At) / (p1.At - p0.At);
            return p0.Pos + (p1.Pos - p0.Pos) * t;
        }
    }
    return 0.5;
}

// Append TCode to text console
const consoleTerminal = document.getElementById("terminal-console");
function appendTCode(tcode) {
    const textNode = document.createTextNode(tcode);
    consoleTerminal.appendChild(textNode);

    // Limit line count to prevent lag
    const lines = consoleTerminal.innerHTML.split("\n");
    if (lines.length > 25) {
        consoleTerminal.innerHTML = lines.slice(lines.length - 25).join("\n");
    }

    consoleTerminal.scrollTop = consoleTerminal.scrollHeight;
}

// Setup standard UI Listeners
function setupEventListeners() {
    // Language change listener
    const langSelect = document.getElementById("header-lang-select");
    if (langSelect) {
        langSelect.onchange = (e) => {
            ws.send(JSON.stringify({
                action: "save_language",
                language: e.target.value
            }));
        };
    }

    // Playback control
    document.getElementById("play-btn").onclick = () => ws.send(JSON.stringify({ action: "play" }));
    document.getElementById("pause-btn").onclick = () => ws.send(JSON.stringify({ action: "pause" }));
    document.getElementById("progress-slider").oninput = (e) => {
        ws.send(JSON.stringify({ action: "seek", position: parseFloat(e.target.value) }));
    };

    const canvas = document.getElementById("heatmap-canvas");
    if (canvas) {
        canvas.onclick = (e) => {
            if (!activeState || activeState.duration <= 0) return;
            const rect = canvas.getBoundingClientRect();
            const clickX = e.clientX - rect.left;
            const pct = clickX / rect.width;
            const targetTime = pct * activeState.duration;
            ws.send(JSON.stringify({ action: "seek", position: targetTime }));
        };
    }

    // Player Save
    document.getElementById("player-save-btn").onclick = () => {
        const name = document.getElementById("player-select").value;
        const ep = document.getElementById("player-endpoint").value;
        const pw = document.getElementById("player-password").value;
        const auto = document.getElementById("player-autoconnect").checked;
        ws.send(JSON.stringify({
            action: "save_player",
            player: name,
            endpoint: ep,
            password: pw,
            auto_connect: auto
        }));
    };

    // Save active Player Selection
    document.getElementById("player-select").onchange = (e) => {
        ws.send(JSON.stringify({
            action: "set_player",
            name: e.target.value
        }));
    };

    document.getElementById("player-connect-btn").onclick = () => {
        const name = document.getElementById("player-select").value;
        const ep = document.getElementById("player-endpoint").value;
        const pw = document.getElementById("player-password").value;
        const auto = document.getElementById("player-autoconnect").checked;
        
        // Save first so any unsaved changes (like password) are recorded
        ws.send(JSON.stringify({
            action: "save_player",
            player: name,
            endpoint: ep,
            password: pw,
            auto_connect: auto
        }));

        ws.send(JSON.stringify({ action: "connect_player", name: name }));
    };

    // Clear Terminal
    document.getElementById("terminal-clear-btn").onclick = () => {
        consoleTerminal.innerHTML = "";
    };

    // Axis limits dual range visual helpers
    const minRange = document.getElementById("axis-min");
    const maxRange = document.getElementById("axis-max");
    minRange.oninput = (e) => {
        const val = parseFloat(e.target.value);
        if (val > parseFloat(maxRange.value)) maxRange.value = val;
        document.getElementById("axis-min-lbl").innerText = val.toFixed(2);
        document.getElementById("axis-max-lbl").innerText = parseFloat(maxRange.value).toFixed(2);
    };
    maxRange.oninput = (e) => {
        const val = parseFloat(e.target.value);
        if (val < parseFloat(minRange.value)) minRange.value = val;
        document.getElementById("axis-min-lbl").innerText = parseFloat(minRange.value).toFixed(2);
        document.getElementById("axis-max-lbl").innerText = val.toFixed(2);
    };

    // Offset slider helper
    document.getElementById("axis-offset").oninput = (e) => {
        document.getElementById("axis-offset-lbl").innerText = parseFloat(e.target.value).toFixed(2);
    };

    // Global Offset controls
    const useGlobal = document.getElementById("use-global-offset");
    const axisOffsetSlider = document.getElementById("axis-offset");

    function saveGlobalOffset() {
        ws.send(JSON.stringify({
            action: "save_global_offset",
            global_offset: parseFloat(axisOffsetSlider.value),
            use_global_offset: useGlobal.checked
        }));
    }
    useGlobal.onchange = saveGlobalOffset;

    // Save Active Axis Configuration
    document.getElementById("axis-save-btn").onclick = () => {
        if (!appSettings) return;
        const settings = {
            name: currentAxis,
            friendly_name: document.getElementById("axis-editor-name").innerText,
            funscript_names: appSettings.axes[currentAxis].funscript_names,
            enabled: document.getElementById("axis-enabled").checked,
            default_value: appSettings.axes[currentAxis].default_value,
            min: parseFloat(document.getElementById("axis-min").value),
            max: parseFloat(document.getElementById("axis-max").value),
            offset: parseFloat(document.getElementById("axis-offset").value),
            invert: document.getElementById("axis-invert").checked,
            interpolation: document.getElementById("axis-interpolation").value,
            auto_home: document.getElementById("axis-autohome").checked,
            auto_home_delay: parseFloat(document.getElementById("axis-autohomedelay").value),
            motion_type: document.getElementById("axis-motion-type").value,
            motion_speed: parseFloat(document.getElementById("axis-motion-speed").value),
            motion_range: parseFloat(document.getElementById("axis-motion-range").value),
            motion_offset: parseFloat(document.getElementById("axis-motion-offset").value)
        };

        ws.send(JSON.stringify({
            action: "save_axis",
            axis: currentAxis,
            settings: settings
        }));
    };

    // Setup Drag-and-Drop dropzone file handling
    const dropzone = document.getElementById("script-dropzone");
    const fileInput = document.getElementById("script-file-input");

    dropzone.onclick = () => fileInput.click();
    fileInput.onchange = (e) => {
        if (e.target.files.length > 0) {
            uploadScriptFile(e.target.files[0]);
        }
    };

    dropzone.ondragover = (e) => {
        e.preventDefault();
        dropzone.classList.add("dragover");
    };

    dropzone.ondragleave = () => {
        dropzone.classList.remove("dragover");
    };

    dropzone.ondrop = (e) => {
        e.preventDefault();
        dropzone.classList.remove("dragover");
        if (e.dataTransfer.files.length > 0) {
            uploadScriptFile(e.dataTransfer.files[0]);
        }
    };

    // Script Directory Add
    document.getElementById("script-dir-add-btn").onclick = () => {
        if (!appSettings) return;
        const input = document.getElementById("script-dir-input");
        const val = input.value.trim();
        if (!val) return;

        const dirs = [...(appSettings.script_directories || [])];
        if (dirs.includes(val)) {
            alert("이미 등록된 경로입니다.");
            return;
        }

        dirs.push(val);
        ws.send(JSON.stringify({
            action: "save_script_directories",
            directories: dirs
        }));
        input.value = "";
    };



    // Folder browser browse trigger
    const browseBtn = document.getElementById("script-dir-browse-btn");
    if (browseBtn) {
        browseBtn.onclick = () => {
            openFolderBrowser();
        };
    }

    // Folder browser helper triggers
    document.getElementById("browse-parent-btn").onclick = () => {
        if (folderBrowseState.parentPath) {
            loadBrowseDir(folderBrowseState.parentPath);
        }
    };

    document.getElementById("browse-drive-select").onchange = (e) => {
        loadBrowseDir(e.target.value);
    };

    document.getElementById("browse-select-btn").onclick = () => {
        const inputField = document.getElementById("script-dir-input");
        if (inputField) {
            inputField.value = folderBrowseState.currentPath;
        }
        closeFolderBrowser();
    };
}

// Upload Funscript file via HTTP multipart post
function uploadScriptFile(file) {
    const formData = new FormData();
    formData.append("axis", currentAxis);
    formData.append("script", file);

    document.getElementById("loaded-script-info").innerText = "업로드 중...";
    
    fetch("/api/upload", {
        method: "POST",
        body: formData
    })
    .then(res => {
        if (res.ok) {
            return res.text();
        }
        throw new Error();
    })
    .then(msg => {
        document.getElementById("loaded-script-info").innerText = msg;
        document.getElementById("loaded-script-info").style.color = "var(--success)";
        // Invalidate cache and fetch again
        fetchHeatmapData();
    })
    .catch(() => {
        document.getElementById("loaded-script-info").innerText = "업로드 실패";
        document.getElementById("loaded-script-info").style.color = "var(--danger)";
    });
}

// Toggle connection of output channels
function toggleOutput(name, checked) {
    if (checked) {
        ws.send(JSON.stringify({ action: "connect_output", name: name }));
    } else {
        ws.send(JSON.stringify({ action: "disconnect_output", name: name }));
    }
}

// Modals Management
function openOutputConfig(name) {
    if (!appSettings || !appSettings.output_targets[name]) return;
    const out = appSettings.output_targets[name];

    document.getElementById("modal-output-name").value = name;
    document.getElementById("modal-title").innerText = `${name.toUpperCase()} 출력 채널 설정`;
    document.getElementById("modal-endpoint").value = out.endpoint;
    document.getElementById("modal-autoconnect").checked = out.auto_connect;

    const serialSelectRow = document.getElementById("serial-port-select-row");
    const serialSelect = document.getElementById("modal-serial-ports-select");

    // Show/hide specific inputs
    if (name === "serial") {
        document.getElementById("modal-endpoint-label").innerText = "시리얼 포트 직접 입력 (e.g. COM3)";
        document.getElementById("modal-baudrate-group").style.display = "flex";
        document.getElementById("modal-filepath-group").style.display = "none";
        document.getElementById("modal-endpoint-group").style.display = "flex";
        document.getElementById("modal-baudrate").value = out.baud_rate.toString();

        // Fetch serial ports dynamically
        if (serialSelectRow && serialSelect) {
            serialSelectRow.style.display = "flex";
            serialSelect.innerHTML = "<option value=''>포트를 조회하는 중...</option>";
            
            fetch("/api/serial-ports")
                .then(res => res.json())
                .then(ports => {
                    serialSelect.innerHTML = "<option value=''>포트 선택 (스캔 결과)</option>";
                    if (ports && ports.length > 0) {
                        ports.forEach(port => {
                            const opt = document.createElement("option");
                            opt.value = port;
                            opt.innerText = port;
                            if (port === out.endpoint) {
                                opt.selected = true;
                            }
                            serialSelect.appendChild(opt);
                        });
                    } else {
                        const opt = document.createElement("option");
                        opt.value = "";
                        opt.innerText = "사용 가능한 포트 없음";
                        serialSelect.appendChild(opt);
                    }
                })
                .catch(() => {
                    serialSelect.innerHTML = "<option value=''>포트 조회 실패</option>";
                });

            // Bind change event to sync with endpoint input
            serialSelect.onchange = (e) => {
                if (e.target.value) {
                    document.getElementById("modal-endpoint").value = e.target.value;
                }
            };
        }
    } else {
        if (serialSelectRow) {
            serialSelectRow.style.display = "none";
        }
        if (name === "file") {
            document.getElementById("modal-endpoint-label").innerText = "연결 주소";
            document.getElementById("modal-baudrate-group").style.display = "none";
            document.getElementById("modal-filepath-group").style.display = "flex";
            document.getElementById("modal-endpoint-group").style.display = "none";
            document.getElementById("modal-filepath").value = out.file_path;
        } else {
            document.getElementById("modal-endpoint-label").innerText = "연결 주소 (IP:PORT)";
            document.getElementById("modal-baudrate-group").style.display = "none";
            document.getElementById("modal-filepath-group").style.display = "none";
            document.getElementById("modal-endpoint-group").style.display = "flex";
        }
    }

    document.getElementById("output-modal").classList.add("open");
}

function closeOutputConfig() {
    document.getElementById("output-modal").classList.remove("open");
}

function saveOutputConfig() {
    const name = document.getElementById("modal-output-name").value;
    const endpoint = document.getElementById("modal-endpoint").value;
    const baud = parseInt(document.getElementById("modal-baudrate").value);
    const filepath = document.getElementById("modal-filepath").value;
    const auto = document.getElementById("modal-autoconnect").checked;

    ws.send(JSON.stringify({
        action: "save_output",
        output: name,
        endpoint: endpoint,
        baud_rate: baud,
        file_path: filepath,
        auto_connect: auto
    }));

    closeOutputConfig();
}

// Helpers
function formatTime(seconds) {
    if (isNaN(seconds) || seconds === null) return "00:00";
    const m = Math.floor(seconds / 60);
    const s = Math.floor(seconds % 60);
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

// ==========================================
// Folder Browser Modal Logic
// ==========================================
let folderBrowseState = {
    currentPath: "",
    parentPath: ""
};

window.openFolderBrowser = function() {
    const modal = document.getElementById("folder-browse-modal");
    if (modal) {
        modal.classList.add("open");
        // Read current text field path, or defaults
        const currentInputVal = document.getElementById("script-dir-input").value.trim();
        loadBrowseDir(currentInputVal);
    }
};

window.closeFolderBrowser = function() {
    const modal = document.getElementById("folder-browse-modal");
    if (modal) {
        modal.classList.remove("open");
    }
};

function loadBrowseDir(path) {
    const url = `/api/browse-dir?path=${encodeURIComponent(path)}`;
    const listContainer = document.getElementById("browse-list");
    const lang = appSettings ? appSettings.language || "ko" : "ko";
    const dict = translations[lang];
    listContainer.innerHTML = `<div style='text-align:center; padding: 20px; color:var(--text-secondary);'>${dict.folder_modal_loading}</div>`;

    fetch(url)
        .then(res => {
            if (res.ok) return res.json();
            throw new Error();
        })
        .then(data => {
            folderBrowseState.currentPath = data.current_path;
            folderBrowseState.parentPath = data.parent_path;
            renderBrowseUI(data);
        })
        .catch(() => {
            listContainer.innerHTML = `<div style='text-align:center; padding: 20px; color:var(--danger);'>${dict.folder_modal_fail}</div>`;
        });
}

function renderBrowseUI(data) {
    document.getElementById("browse-current-path-lbl").innerText = data.current_path;
    
    // Parent button control
    const parentBtn = document.getElementById("browse-parent-btn");
    if (data.parent_path) {
        parentBtn.style.display = "flex";
    } else {
        parentBtn.style.display = "none";
    }

    // Windows drives select control
    const driveGroup = document.getElementById("browse-drive-group");
    const driveSelect = document.getElementById("browse-drive-select");
    
    if (data.drives && data.drives.length > 1) {
        driveGroup.style.display = "flex";
        driveSelect.innerHTML = "";
        data.drives.forEach(drv => {
            const opt = document.createElement("option");
            opt.value = drv;
            opt.innerText = drv;
            // Check if current path starts with this drive
            if (data.current_path.toUpperCase().startsWith(drv.toUpperCase())) {
                opt.selected = true;
            }
            driveSelect.appendChild(opt);
        });
    } else {
        driveGroup.style.display = "none";
    }

    // Folder listing
    const listContainer = document.getElementById("browse-list");
    listContainer.innerHTML = "";

    const lang = appSettings ? appSettings.language || "ko" : "ko";
    const dict = translations[lang];

    if (data.directories && data.directories.length > 0) {
        data.directories.sort().forEach(dir => {
            const row = document.createElement("div");
            row.style = "display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 6px; cursor: pointer; transition: background 0.15s ease; color: var(--text-primary); font-size: 0.9rem; user-select: none;";
            row.className = "browse-folder-row";
            
            // Hover styling via hover listener
            row.onmouseenter = () => row.style.background = "rgba(255,255,255,0.06)";
            row.onmouseleave = () => row.style.background = "transparent";

            row.innerHTML = `
                <span class="material-icons-round" style="color: #ffd600; font-size: 1.2rem;">folder</span>
                <span style="flex: 1; word-break: break-all;">${dir}</span>
            `;

            // Double click to enter folder
            row.ondblclick = () => {
                let targetPath = data.current_path;
                if (!targetPath.endsWith("/") && !targetPath.endsWith("\\")) {
                    // Detect Windows separator or unix
                    const sep = targetPath.includes("\\") ? "\\" : "/";
                    targetPath += sep;
                }
                targetPath += dir;
                loadBrowseDir(targetPath);
            };

            // Single click visual active highlight
            row.onclick = () => {
                document.querySelectorAll(".browse-folder-row").forEach(el => el.style.border = "none");
                row.style.border = "1px solid var(--primary)";
            };

            listContainer.appendChild(row);
        });
    } else {
        listContainer.innerHTML = `<div style='text-align:center; padding: 20px; color:var(--text-secondary); font-size: 0.85rem;'>${dict.folder_modal_no_sub}</div>`;
    }
}

function applyLanguage(lang) {
    if (!translations[lang]) return;
    const dict = translations[lang];

    // Translate innerText for elements with data-i18n
    document.querySelectorAll("[data-i18n]").forEach(el => {
        const key = el.getAttribute("data-i18n");
        if (dict[key]) {
            const icon = el.querySelector(".material-icons-round");
            const strong = el.querySelector("strong");
            
            if (icon || strong) {
                if (key === "dropzone_desc") {
                    el.innerHTML = lang === "ko" 
                        ? "여기에 <strong>.funscript</strong> 또는 <strong>.csv</strong> 파일을 드래그 앤 드롭 하거나 클릭하여 로드하세요."
                        : "Drag & drop a <strong>.funscript</strong> or <strong>.csv</strong> file here, or click to browse.";
                } else if (key === "axis_save_btn") {
                    el.innerHTML = `<span class="material-icons-round">check</span> ` + dict[key];

                } else if (key === "folder_modal_parent") {
                    el.innerText = dict[key];
                }
            } else {
                el.innerText = dict[key];
            }
        }
    });

    // Translate title tooltips
    document.querySelectorAll("[data-i18n-title]").forEach(el => {
        const key = el.getAttribute("data-i18n-title");
        if (dict[key]) {
            el.title = dict[key];
        }
    });

    // Translate placeholders
    document.querySelectorAll("[data-i18n-placeholder]").forEach(el => {
        const key = el.getAttribute("data-i18n-placeholder");
        if (dict[key]) {
            el.placeholder = dict[key];
        }
    });
}
