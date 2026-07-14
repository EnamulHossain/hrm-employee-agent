# HRM Employee Desktop Agent v2.0

A lightweight, cross-platform background monitoring agent that integrates with the HRM system. It provides secure device registration, real-time heartbeat monitoring, system idle/activity detection, active application tracking, and reliable automatic startup across **Windows**, **Linux**, and **macOS**.

## Key Features

- **Zero Dependencies**: Standalone static binaries — no runtime prerequisites (no CGo, no MinGW, no frameworks).
- **Background Daemon**: Runs silently with minimal CPU/RAM footprint.
- **Cross-Platform**: Native support for Windows 10/11, Linux (Ubuntu/Debian/Fedora), and macOS (Intel + Apple Silicon).
- **Web-based Setup**: On first run, launches a premium local setup wizard at `http://localhost:8089`.
- **Reliable Autostart**: Multiple redundant mechanisms per platform ensure the agent survives reboots.
- **Network Resilience**: Waits for network on boot, exponential backoff on failures, offline log queuing.
- **Single Instance Lock**: PID-based locking prevents duplicate agent processes.
- **Self-Healing**: Automatically re-registers if deauthorized, with graceful fallback to manual setup.
- **Anti-Virus Optimized**: Embedded Windows metadata (version info, manifest, company details) to minimize false positives.

---

## Building the Agent

### Prerequisites
- Go 1.21+ installed
- Internet access (for `goversioninfo` download on first build)

### Using the Build Script (Recommended)

```bash
chmod +x build.sh
./build.sh
```

This produces **four binaries**:

| Binary | Platform | Architecture |
|--------|----------|-------------|
| `employee-agent` | Linux | amd64 |
| `employee-agent.exe` | Windows | amd64 |
| `employee-agent-darwin-amd64` | macOS | Intel (x86_64) |
| `employee-agent-darwin-arm64` | macOS | Apple Silicon (M1/M2/M3/M4) |

### Manual Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o employee-agent .

# Windows (with version info)
go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -o resource_windows.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -H=windowsgui" -o employee-agent.exe .

# macOS Intel
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o employee-agent-darwin-amd64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o employee-agent-darwin-arm64 .
```

---

## Installation & Running

### Windows

1. Copy `employee-agent.exe` to a **permanent** location (e.g., `C:\Program Files\HRM Agent\employee-agent.exe`).
2. Run the executable. It will launch a setup page at `http://localhost:8089` in your browser.
3. Enter the **Company HRM URL**, **Employee Code**, and **Secure Registration Key**.
4. The agent will configure itself to start on boot via:
   - **Registry Run key** (`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`)
   - **Windows Task Scheduler** (on-logon trigger with 15s delay)
5. Close the browser tab. The agent runs headless in the background.

#### Antivirus Notes (Windows)

The v2.0 binary includes embedded version info, an application manifest, and proper metadata to minimize false positives. However, **unsigned binaries may still be flagged** by some antivirus products. To resolve:

1. **Add an exclusion** in Windows Defender:
   - Settings → Privacy & Security → Windows Security → Virus & Threat Protection
   - Manage settings → Exclusions → Add an exclusion → File → Select `employee-agent.exe`
2. **For enterprise deployments**: Use Group Policy to whitelist the binary path, or sign the binary with an EV code signing certificate.

### Linux (Ubuntu/Debian/Fedora)

1. Copy `employee-agent` to a permanent directory:
   ```bash
   sudo mkdir -p /opt/hrm-agent
   sudo cp employee-agent /opt/hrm-agent/
   sudo chmod +x /opt/hrm-agent/employee-agent
   ```
2. Run the agent:
   ```bash
   /opt/hrm-agent/employee-agent &
   ```
3. Complete registration at `http://localhost:8089`.
4. The agent auto-configures two autostart mechanisms:
   - **systemd user service** at `~/.config/systemd/user/employee-agent.service` (primary)
   - **XDG desktop autostart** at `~/.config/autostart/employee-agent.desktop` (fallback for GNOME/KDE)

#### Linux Dependencies (Optional)
For full idle detection and window tracking:
```bash
# Ubuntu/Debian
sudo apt install xprintidle xdotool

# Fedora
sudo dnf install xprintidle xdotool
```

### macOS

1. Copy the correct binary for your Mac:
   - **Intel Mac**: `employee-agent-darwin-amd64`
   - **Apple Silicon Mac** (M1/M2/M3/M4): `employee-agent-darwin-arm64`

   ```bash
   mkdir -p ~/Applications/HRM-Agent
   cp employee-agent-darwin-arm64 ~/Applications/HRM-Agent/employee-agent
   chmod +x ~/Applications/HRM-Agent/employee-agent
   ```

2. Run the agent:
   ```bash
   ~/Applications/HRM-Agent/employee-agent &
   ```

3. Complete registration at `http://localhost:8089`.

4. The agent creates a **LaunchAgent** at `~/Library/LaunchAgents/com.hrm.employee-agent.plist` with `KeepAlive` enabled.

#### macOS Permissions

The agent needs **Accessibility permissions** to track active window titles:
1. Go to **System Settings → Privacy & Security → Accessibility**
2. Click the **+** button and add the `employee-agent` binary
3. If running from Terminal, you may also need to add **Terminal.app** to the list

Without Accessibility permissions, the agent will still track idle time and send heartbeats, but window/app titles may show as "macOS Desktop".

---

## Configuration & Data Locations

| Platform | Config File | PID Lock |
|----------|------------|----------|
| Windows | `%USERPROFILE%\.employee-agent\config.json` | `%USERPROFILE%\.employee-agent\agent.pid` |
| Linux | `~/.employee-agent/config.json` | `~/.employee-agent/agent.pid` |
| macOS | `~/.employee-agent/config.json` | `~/.employee-agent/agent.pid` |

### Autostart Locations

| Platform | Primary | Fallback |
|----------|---------|----------|
| Windows | Task Scheduler: `HRM_Employee_Desktop_Agent` | Registry: `HKCU\..\Run\HRM Employee Agent` |
| Linux | `~/.config/systemd/user/employee-agent.service` | `~/.config/autostart/employee-agent.desktop` |
| macOS | `~/Library/LaunchAgents/com.hrm.employee-agent.plist` | — |

---

## Troubleshooting

### Agent doesn't start after reboot
1. **Windows**: Check Task Scheduler for `HRM_Employee_Desktop_Agent` task. Verify the binary path is correct.
2. **Linux**: Run `systemctl --user status employee-agent.service` to check the service status.
3. **macOS**: Run `launchctl list | grep hrm` to verify the LaunchAgent is loaded.

### Agent is flagged as a virus (Windows)
- Add the binary to Windows Defender exclusions (see Antivirus Notes above).
- For enterprise deployment, distribute via GPO with a path exclusion policy.
- Consider purchasing an EV code signing certificate for production use.

### Force re-registration
Delete the config file and restart the agent:
```bash
# Linux/macOS
rm ~/.employee-agent/config.json

# Windows (PowerShell)
Remove-Item "$env:USERPROFILE\.employee-agent\config.json"
```

### Check agent logs
- **Linux**: Standard output goes to systemd journal: `journalctl --user -u employee-agent.service -f`
- **macOS**: Check `~/.employee-agent/agent-stdout.log` and `agent-stderr.log`
- **Windows**: Run from a terminal to see output, or check Event Viewer.
