package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const agentVersion = "2.0.0"

// htmlPage is the HTML content for the registration setup screen.
const htmlPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Employee Desktop Agent Setup</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #2196F3;
            --primary-hover: #1e88e5;
            --bg: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --text: #f8fafc;
            --text-muted: #94a3b8;
            --border: #334155;
        }
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
            font-family: 'Outfit', sans-serif;
        }
        body {
            background: radial-gradient(circle at top right, #1e1b4b, #0f172a);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: var(--card-bg);
            backdrop-filter: blur(10px);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 40px;
            width: 100%;
            max-width: 480px;
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3), 0 8px 10px -6px rgba(0, 0, 0, 0.3);
            text-align: center;
        }
        h2 {
            font-size: 28px;
            font-weight: 700;
            margin-bottom: 8px;
            background: linear-gradient(to right, #38bdf8, #818cf8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        p.subtitle {
            color: var(--text-muted);
            font-size: 14px;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
            text-align: left;
        }
        label {
            display: block;
            font-size: 13px;
            font-weight: 600;
            color: var(--text-muted);
            margin-bottom: 6px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        input {
            width: 100%;
            background: #020617;
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 12px 16px;
            color: #fff;
            font-size: 15px;
            transition: all 0.2s;
        }
        input:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.2);
        }
        button {
            width: 100%;
            background: linear-gradient(135deg, var(--primary) 0%, #4f46e5 100%);
            border: none;
            border-radius: 8px;
            padding: 14px;
            color: white;
            font-weight: 600;
            font-size: 16px;
            cursor: pointer;
            transition: opacity 0.2s;
            margin-top: 10px;
        }
        button:hover {
            opacity: 0.9;
        }
        .alert {
            border-radius: 8px;
            padding: 12px;
            font-size: 14px;
            margin-top: 20px;
            display: none;
        }
        .alert-error {
            background: rgba(239, 68, 68, 0.2);
            border: 1px solid rgba(239, 68, 68, 0.4);
            color: #fca5a5;
        }
        .alert-success {
            background: rgba(34, 197, 94, 0.2);
            border: 1px solid rgba(34, 197, 94, 0.4);
            color: #86efac;
        }
        .loader {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid rgba(255,255,255,0.3);
            border-radius: 50%;
            border-top-color: #fff;
            animation: spin 1s ease-in-out infinite;
            margin-right: 8px;
            vertical-align: middle;
            display: none;
        }
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Agent Device Registration</h2>
        <p class="subtitle">Enter your company's HRM settings to register this device.</p>
        
        <form id="regForm">
            <div class="form-group">
                <label for="hrmUrl">Company HRM URL</label>
                <input type="url" id="hrmUrl" placeholder="https://hrm.yourcompany.com" required>
            </div>
            <div class="form-group">
                <label for="employeeCode">Employee Code / ID</label>
                <input type="text" id="employeeCode" placeholder="EMP-12345" required>
            </div>
            <div class="form-group">
                <label for="regKey">Secure Registration Key</label>
                <input type="password" id="regKey" placeholder="••••••••••••••••" required>
            </div>
            <button type="submit" id="btnSubmit">
                <span class="loader" id="loader"></span>
                Register Device
            </button>
        </form>

        <div class="alert alert-error" id="errorMsg"></div>
        <div class="alert alert-success" id="successMsg">Device registered successfully! You can close this window now.</div>
    </div>

    <script>
        document.getElementById('regForm').addEventListener('submit', function(e) {
            e.preventDefault();
            const hrmUrl = document.getElementById('hrmUrl').value;
            const employeeCode = document.getElementById('employeeCode').value;
            const regKey = document.getElementById('regKey').value;
            
            const loader = document.getElementById('loader');
            const errorMsg = document.getElementById('errorMsg');
            const successMsg = document.getElementById('successMsg');
            const btnSubmit = document.getElementById('btnSubmit');

            errorMsg.style.display = 'none';
            successMsg.style.display = 'none';
            loader.style.display = 'inline-block';
            btnSubmit.disabled = true;

            fetch('/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    hrm_url: hrmUrl,
                    employee_code: employeeCode,
                    registration_key: regKey
                })
            })
            .then(res => res.json().then(data => ({ status: res.status, body: data })))
            .then(res => {
                loader.style.display = 'none';
                btnSubmit.disabled = false;
                if (res.status === 200 && res.body.success) {
                    successMsg.style.display = 'block';
                    setTimeout(() => {
                        window.close();
                    }, 2000);
                } else {
                    errorMsg.innerText = res.body.message || 'Registration failed.';
                    errorMsg.style.display = 'block';
                }
            })
            .catch(err => {
                loader.style.display = 'none';
                btnSubmit.disabled = false;
                errorMsg.innerText = 'Unable to connect to the agent setup service.';
                errorMsg.style.display = 'block';
            });
        });
    </script>
</body>
</html>`

// ActivityLog represents individual check logs to be synced.
type ActivityLog struct {
	ActivityStatus string `json:"activity_status"`
	ActiveWindow   string `json:"active_window"`
	IdleSeconds    int    `json:"idle_seconds"`
	Timestamp      string `json:"timestamp"`
}

func parseAppName(title string) string {
	title = strings.TrimSpace(title)
	if title == "" || title == "Idle" || title == "System" {
		return title
	}

	// Normalize browser names
	var browserName string
	var tabTitle string
	lowerTitle := strings.ToLower(title)

	if strings.Contains(lowerTitle, "google chrome") || strings.Contains(lowerTitle, "chrome") {
		browserName = "Chrome"
	} else if strings.Contains(lowerTitle, "firefox") {
		browserName = "Firefox"
	} else if strings.Contains(lowerTitle, "edge") {
		browserName = "Edge"
	} else if strings.Contains(lowerTitle, "safari") {
		browserName = "Safari"
	} else if strings.Contains(lowerTitle, "opera") {
		browserName = "Opera"
	} else if strings.Contains(lowerTitle, "brave") {
		browserName = "Brave"
	}

	if browserName != "" {
		// Clean the tab title by removing browser suffixes
		tabTitle = title
		suffixes := []string{
			" - Google Chrome", " Google Chrome", " - Chrome", " Chrome",
			" - Mozilla Firefox", " Mozilla Firefox", " - Firefox", " Firefox",
			" - Microsoft Edge", " Microsoft Edge", " - Edge", " Edge",
			" - Brave", " Brave", " - Safari", " Safari", " - Opera", " Opera",
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(strings.ToLower(tabTitle), strings.ToLower(suffix)) {
				tabTitle = tabTitle[:len(tabTitle)-len(suffix)]
				break
			}
		}
		tabTitle = strings.TrimSpace(tabTitle)
		if tabTitle == "" || strings.EqualFold(tabTitle, browserName) {
			return browserName
		}
		if len(tabTitle) > 40 {
			tabTitle = tabTitle[:40] + "..."
		}
		return fmt.Sprintf("%s (%s)", browserName, tabTitle)
	}

	// For non-browser apps, try to extract app name using splitters
	parts := strings.Split(title, " - ")
	if len(parts) > 1 {
		app := strings.TrimSpace(parts[len(parts)-1])
		if app != "" {
			if len(app) > 60 {
				return app[:60]
			}
			return app
		}
	}

	parts = strings.Split(title, " | ")
	if len(parts) > 1 {
		app := strings.TrimSpace(parts[len(parts)-1])
		if app != "" {
			if len(app) > 60 {
				return app[:60]
			}
			return app
		}
	}

	if len(title) > 60 {
		return title[:60]
	}
	return title
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return name
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Single Instance Lock
// ──────────────────────────────────────────────────────────────────────────────

// acquireLock creates a PID file to prevent multiple instances.
// Returns true if the lock was acquired, false if another instance is running.
func acquireLock() bool {
	pidPath, err := getPIDFilePath()
	if err != nil {
		log.Printf("[WARNING] Could not determine PID file path: %v", err)
		return true // Proceed anyway
	}

	// Check if PID file exists and if the process is still running
	data, err := os.ReadFile(pidPath)
	if err == nil {
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if isProcessRunning(pid) {
				log.Printf("[WARNING] Another instance is already running (PID %d). Exiting.", pid)
				return false
			}
		}
		// Stale PID file, remove it
		_ = os.Remove(pidPath)
	}

	// Write our PID
	pid := os.Getpid()
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Printf("[WARNING] Could not create PID file: %v", err)
	}

	return true
}

// releaseLock removes the PID file.
func releaseLock() {
	pidPath, err := getPIDFilePath()
	if err != nil {
		return
	}
	_ = os.Remove(pidPath)
}

func getPIDFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".employee-agent")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "agent.pid"), nil
}

// isProcessRunning checks if a process with the given PID is running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. We need to send signal 0 to check.
	if runtime.GOOS != "windows" {
		err = process.Signal(os.Signal(nil))
		// Use a platform-agnostic approach: try to find the process
		if err != nil {
			return false
		}
		return true
	}
	// On Windows, FindProcess succeeds only if the process exists
	process.Release()
	return true
}

// ──────────────────────────────────────────────────────────────────────────────
// Network Readiness
// ──────────────────────────────────────────────────────────────────────────────

// waitForNetwork waits until network connectivity is available, with a timeout.
// Returns true if network is available, false if timed out.
func waitForNetwork(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		conn, err := net.DialTimeout("tcp", "dns.google:443", 3*time.Second)
		if err == nil {
			conn.Close()
			if attempt > 1 {
				log.Printf("[INFO] Network ready after %d attempts.", attempt)
			}
			return true
		}

		// Also try a DNS resolution as fallback
		_, err = net.LookupHost("google.com")
		if err == nil {
			if attempt > 1 {
				log.Printf("[INFO] Network ready (DNS) after %d attempts.", attempt)
			}
			return true
		}

		log.Printf("[INFO] Waiting for network... (attempt %d)", attempt)
		time.Sleep(3 * time.Second)
	}

	log.Println("[WARNING] Network readiness timed out. Proceeding anyway.")
	return false
}

// startSetupServer starts the local web server on port 8089 to allow registration.
func startSetupServer(config *Config) {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8089",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(htmlPage))
		}
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			var req struct {
				HRMURL          string `json:"hrm_url"`
				EmployeeCode    string `json:"employee_code"`
				RegistrationKey string `json:"registration_key"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				w.WriteHeader(400)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request body"})
				return
			}

			// Clean HRM URL (strip trailing slash)
			hrmUrl := strings.TrimRight(req.HRMURL, "/")

			// Prepare registration payload
			regReqBody, _ := json.Marshal(map[string]string{
				"employee_code":    req.EmployeeCode,
				"registration_key": req.RegistrationKey,
				"device_uuid":      config.DeviceUUID,
				"hostname":         getHostname(),
				"operating_system": runtime.GOOS,
				"ip_address":       GetLocalIP(),
				"mac_address":      GetMACAddress(),
			})

			resp, err := http.Post(hrmUrl+"/api/agent/register", "application/json", bytes.NewBuffer(regReqBody))
			if err != nil {
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to connect to HRM server: " + err.Error()})
				return
			}
			defer resp.Body.Close()

			var regResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Token   string `json:"token"`
				Config  struct {
					HeartbeatInterval int `json:"heartbeat_interval"`
					ActivityInterval  int `json:"activity_interval"`
				} `json:"config"`
			}

			err = json.NewDecoder(resp.Body).Decode(&regResp)
			if err != nil {
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to decode HRM response"})
				return
			}

			if resp.StatusCode != 200 || !regResp.Success {
				w.WriteHeader(resp.StatusCode)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": regResp.Message})
				return
			}

			// Save settings
			config.HRMURL = hrmUrl
			config.EmployeeCode = req.EmployeeCode
			config.RegistrationKey = req.RegistrationKey
			config.Token = regResp.Token
			config.HeartbeatInterval = regResp.Config.HeartbeatInterval
			config.ActivityInterval = regResp.Config.ActivityInterval
			config.Registered = true

			err = SaveConfig(*config)
			if err != nil {
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to save local config: " + err.Error()})
				return
			}

			// Register autostart
			_ = ConfigureAutostart()

			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Registration successful"})

			// Gracefully close server in background
			go func() {
				time.Sleep(1 * time.Second)
				_ = server.Close()
			}()
		}
	})

	log.Println("[INFO] Starting registration server on http://localhost:8089")
	openBrowser("http://localhost:8089")
	_ = server.ListenAndServe()
}

// autoRegister attempts to register the device in the background using stored credentials.
// Returns (success, isNetworkError)
func autoRegister(config *Config) (bool, bool) {
	if config.HRMURL == "" || config.EmployeeCode == "" || config.RegistrationKey == "" {
		return false, false
	}
	log.Printf("[INFO] Attempting automatic re-registration for employee %s...", config.EmployeeCode)

	regReqBody, _ := json.Marshal(map[string]string{
		"employee_code":    config.EmployeeCode,
		"registration_key": config.RegistrationKey,
		"device_uuid":      config.DeviceUUID,
		"hostname":         getHostname(),
		"operating_system": runtime.GOOS,
		"ip_address":       GetLocalIP(),
		"mac_address":      GetMACAddress(),
	})

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(config.HRMURL+"/api/agent/register", "application/json", bytes.NewBuffer(regReqBody))
	if err != nil {
		log.Printf("[ERROR] Auto-registration request failed: %v", err)
		return false, true
	}
	defer resp.Body.Close()

	var regResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Token   string `json:"token"`
		Config  struct {
			HeartbeatInterval int `json:"heartbeat_interval"`
			ActivityInterval  int `json:"activity_interval"`
		} `json:"config"`
	}

	err = json.NewDecoder(resp.Body).Decode(&regResp)
	if err != nil || resp.StatusCode != 200 || !regResp.Success {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = regResp.Message
		}
		log.Printf("[ERROR] Auto-registration failed (status %d): %s", resp.StatusCode, errMsg)
		// If response is a client error like 400, 401, 403, 404, it means credentials are invalid
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return false, false
		}
		// Otherwise treat it as a server/network issue (should retry silently)
		return false, true
	}

	config.Token = regResp.Token
	config.HeartbeatInterval = regResp.Config.HeartbeatInterval
	config.ActivityInterval = regResp.Config.ActivityInterval
	config.Registered = true
	_ = SaveConfig(*config)

	log.Printf("[INFO] Auto-registration successful. New token acquired.")
	return true, false
}

// startDaemon starts the heartbeat monitor and activity synchronization daemon loops.
func startDaemon(config *Config) {
	log.Printf("[INFO] Daemon running. Tracking employee: %s", config.EmployeeCode)

	var accumulatedLogs []ActivityLog
	var logsMutex sync.Mutex

	hbTicker := time.NewTicker(time.Duration(config.HeartbeatInterval) * time.Second)
	actTicker := time.NewTicker(time.Duration(config.ActivityInterval) * time.Second)
	checkTicker := time.NewTicker(5 * time.Second) // Check active/idle state every 5 seconds

	defer hbTicker.Stop()
	defer actTicker.Stop()
	defer checkTicker.Stop()

	lastStatus := "ACTIVE"

	sendHeartbeat := func() {
		url := config.HRMURL + "/api/agent/heartbeat"
		body, _ := json.Marshal(map[string]interface{}{
			"device_uuid": config.DeviceUUID,
			"status":      lastStatus,
			"ip_address":  GetLocalIP(),
			"mac_address": GetMACAddress(),
		})

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[ERROR] Heartbeat creation: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+config.Token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[ERROR] Heartbeat connect: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 || resp.StatusCode == 404 {
			log.Printf("[WARNING] Authentication failed (%d). Attempting auto re-registration...", resp.StatusCode)
			success, isNetErr := autoRegister(config)
			if success {
				return
			}
			if isNetErr {
				log.Printf("[WARNING] Auto re-registration failed due to network/server error. Retaining registration and retrying later.")
				return
			}
			log.Printf("[WARNING] Auto re-registration failed (invalid credentials). Revoking registration.")
			config.Registered = false
			config.Token = ""
			_ = SaveConfig(*config)
			return
		}

		var hbResp struct {
			Success bool `json:"success"`
			Config  struct {
				HeartbeatInterval int `json:"heartbeat_interval"`
				ActivityInterval  int `json:"activity_interval"`
			} `json:"config"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&hbResp); err == nil && hbResp.Success {
			// Update polling intervals if changed by admin
			if hbResp.Config.HeartbeatInterval != config.HeartbeatInterval || hbResp.Config.ActivityInterval != config.ActivityInterval {
				config.HeartbeatInterval = hbResp.Config.HeartbeatInterval
				config.ActivityInterval = hbResp.Config.ActivityInterval
				_ = SaveConfig(*config)
				hbTicker.Reset(time.Duration(config.HeartbeatInterval) * time.Second)
				actTicker.Reset(time.Duration(config.ActivityInterval) * time.Second)
				log.Println("[INFO] Heartbeat and Activity intervals updated.")
			}
		}
		log.Printf("[INFO] Heartbeat success. Status: %s", lastStatus)
	}

	sendActivityLogs := func() {
		logsMutex.Lock()
		if len(accumulatedLogs) == 0 {
			logsMutex.Unlock()
			return
		}
		logsToSend := accumulatedLogs
		accumulatedLogs = nil
		logsMutex.Unlock()

		url := config.HRMURL + "/api/agent/activity"
		body, _ := json.Marshal(map[string]interface{}{
			"device_uuid": config.DeviceUUID,
			"logs":        logsToSend,
		})

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[ERROR] Activity log creation: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+config.Token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[ERROR] Activity log connect: %v", err)
			// Return back to queue
			logsMutex.Lock()
			accumulatedLogs = append(logsToSend, accumulatedLogs...)
			logsMutex.Unlock()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 || resp.StatusCode == 404 {
			log.Printf("[WARNING] Authentication failed (%d) during activity sync. Attempting auto re-registration...", resp.StatusCode)
			success, isNetErr := autoRegister(config)
			if success {
				// Return logs to queue so they can be resent on next cycle
				logsMutex.Lock()
				accumulatedLogs = append(logsToSend, accumulatedLogs...)
				logsMutex.Unlock()
				return
			}
			if isNetErr {
				log.Printf("[WARNING] Auto re-registration failed due to network/server error during activity sync. Retaining registration and retrying later.")
				logsMutex.Lock()
				accumulatedLogs = append(logsToSend, accumulatedLogs...)
				logsMutex.Unlock()
				return
			}
			log.Printf("[WARNING] Auto re-registration failed (invalid credentials). Revoking.")
			config.Registered = false
			config.Token = ""
			_ = SaveConfig(*config)
			return
		}

		log.Printf("[INFO] Synced %d activity logs successfully.", len(logsToSend))
	}

	// Send initial heartbeat
	sendHeartbeat()

	for {
		if !config.Registered {
			return
		}

		select {
		case <-hbTicker.C:
			sendHeartbeat()

		case <-actTicker.C:
			sendActivityLogs()

		case <-checkTicker.C:
			idleSecs, err := GetSystemIdleTime()
			if err != nil {
				log.Printf("[ERROR] Idle checker: %v", err)
				idleSecs = 0
			}

			status := "ACTIVE"
			if int(idleSecs) >= config.ActivityInterval {
				status = "IDLE"
			}

			lastStatus = status

			// Get and parse active window / application title
			rawWindow := GetActiveWindowTitle()
			activeWin := parseAppName(rawWindow)

			logsMutex.Lock()
			accumulatedLogs = append(accumulatedLogs, ActivityLog{
				ActivityStatus: status,
				ActiveWindow:   activeWin,
				IdleSeconds:    int(idleSecs),
				Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
			})
			logsMutex.Unlock()
		}
	}
}

func main() {
	log.Printf("[INFO] Starting HRM Employee Desktop Agent v%s (%s/%s)...", agentVersion, runtime.GOOS, runtime.GOARCH)

	// ── Single Instance Lock ──
	if !acquireLock() {
		log.Println("[INFO] Another instance is already running. Exiting gracefully.")
		os.Exit(0)
	}
	defer releaseLock()

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Error loading configuration: %v", err)
	}

	if config.DeviceUUID == "" {
		config.DeviceUUID = generateUUID()
		_ = SaveConfig(config)
	}

	// Always ensure autostart is configured when agent runs
	_ = ConfigureAutostart()

	// ── Wait for network on startup (up to 60 seconds) ──
	// This is critical after a reboot when the agent starts before network is ready.
	log.Println("[INFO] Checking network connectivity...")
	waitForNetwork(60 * time.Second)

	// ── Main Loop with Exponential Backoff ──
	retryDelay := 10 * time.Second
	maxRetryDelay := 5 * time.Minute

	for {
		if !config.Registered {
			// If we have saved credentials, try to automatically re-register in the background
			success := false
			isNetErr := false
			if config.HRMURL != "" && config.EmployeeCode != "" && config.RegistrationKey != "" {
				log.Println("[INFO] Device unregistered but credentials found. Attempting background auto-registration...")
				success, isNetErr = autoRegister(&config)
			}

			if !success {
				if isNetErr {
					log.Printf("[WARNING] Auto-registration failed due to network error. Retrying in %v...", retryDelay)
					time.Sleep(retryDelay)
					// Exponential backoff
					retryDelay = retryDelay * 2
					if retryDelay > maxRetryDelay {
						retryDelay = maxRetryDelay
					}
					continue
				}
				log.Println("[INFO] Device unregistered or invalid credentials. Launching setup web interface...")
				startSetupServer(&config)
				// Reset retry delay after manual registration attempt
				retryDelay = 10 * time.Second
			} else {
				// Reset retry delay on success
				retryDelay = 10 * time.Second
			}
		} else {
			log.Println("[INFO] Device registered. Starting daemon monitoring...")
			startDaemon(&config)
			// Reset retry delay when daemon exits normally
			retryDelay = 10 * time.Second
		}
		time.Sleep(2 * time.Second)
	}
}
