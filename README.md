# Employee Desktop Agent

This is a lightweight, cross-platform background tracking desktop agent built in Go for the HRM system. It supports secure registration, periodic heartbeat pings, system idle/activity detection, and automatic startup configuration for both Windows and Linux.

## Key Features
- **Zero Dependencies**: Standalone binaries with no runtime prerequisites (like CGo, GTK, or MinGW).
- **Background Daemon**: Runs silently in the background with minimal CPU/RAM footprint.
- **Web-based Setup**: On first run or if deauthorized, spins up a premium local setup wizard at `http://localhost:8089` and opens the browser.
- **Autostart Support**: Automatically registers to launch on system boot/user login.
- **Self-Healing**: Automatically logs out and re-launches the registration wizard if deauthorized or revoked by the company administrator.

---

## Building the Agent

You can build the binaries directly using Go (1.18+ recommended) or run the provided build script.

### 1. Using the Build Script (Linux)
Make the script executable and run it:
```bash
chmod +x build.sh
./build.sh
```
This produces two binaries in the `agent` directory:
- `employee-agent` (Linux executable)
- `employee-agent.exe` (Windows executable)

### 2. Manual Cross-Compilation
To build manually from the command line:

#### For Ubuntu Linux:
```bash
GOOS=linux GOARCH=amd64 go build -o employee-agent .
```

#### For Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o employee-agent.exe .
```

---

## Installation & Running

### 1. On Windows
1. Copy `employee-agent.exe` to a permanent location on the employee's machine (e.g., `C:\Program Files\EmployeeAgent\employee-agent.exe`).
2. Run the executable. It will detect that it is unregistered, spin up a local setup server, and open your web browser to `http://localhost:8089`.
3. Enter your **Company HRM URL**, **Employee ID/Code**, and the **Secure Registration Key** (generated in the HRM Admin Dashboard).
4. Upon successful registration, the agent will configure itself to run on system startup and begin its background monitoring cycle.
5. You can now close the browser tab. The agent runs fully headless in the background.

### 2. On Ubuntu Linux
1. Copy `employee-agent` to a permanent directory (e.g., `~/bin/employee-agent` or `/usr/local/bin/employee-agent`).
2. Run the agent in the background:
   ```bash
   nohup ./employee-agent > /dev/null 2>&1 &
   ```
3. A setup screen will open in your default browser at `http://localhost:8089`.
4. Enter the required registration details and submit.
5. The agent will automatically generate an autostart configuration desktop entry at `~/.config/autostart/employee-agent.desktop`.

---

## Configuration & Logs Location

All local registration details and intervals are saved in the user's home directory.
- **Linux config**: `~/.config/employee-agent/config.json`
- **Windows config**: `%USERPROFILE%\.employee-agent\config.json`

To force re-registration or manually update the configuration, you can delete the `config.json` file or clear its contents and restart the agent.
