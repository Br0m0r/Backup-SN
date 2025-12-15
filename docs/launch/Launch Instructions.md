# Starting All Services - Step by Step

## Prerequisites
- Go installed (the launchers can install it when supported package managers are available).
- Node.js 18+ and npm installed (also auto-installed by the launchers).
- SQLite database file at `<repo>/social_network.db` (automatically created by the launchers if it is missing).

> Replace `<repo>` with the absolute path to your `social-network` clone in any commands below.

## Automated Launcher Scripts
Both launchers bring up Go microservices plus the Vite frontend, wiring up `DATABASE_PATH`, `AUTH_SERVICE_URL`, and installing dependencies as needed. Use `Ctrl+C` at any time to tear everything down.

### Windows (PowerShell 5.1 or 7+)

```powershell
cd <repo>
# Windows PowerShell 5.1
powershell.exe -ExecutionPolicy Bypass -File .\Launch.ps1
# or PowerShell 7+
pwsh -File ./Launch.ps1
```

What this does:
- Works on stock WindowsPowerShell 5.1 and PowerShell 7+.
- Verifies (and when possible auto-installs via winget/choco) Go, Node.js/npm, and a GCC-compatible toolchain needed by `go-sqlite3`.
- Runs `go mod download`, `npm install` in the repo root, and `npm install` in `frontend/`.
- Starts Auth, Users, Posts, Groups, Chat, and Notifications via background PowerShell jobs plus the Vite dev server, then opens the browser.

Useful flags:
- `-SkipInstall` - trust that prerequisites already exist.
- `-SkipFrontend` - start backend services only.
- `-FrontendHost <host>` / `-FrontendPort <port>` - customize the Vite dev server binding.

Log streaming examples:
- `Get-Job` shows service/job status.
- `Receive-Job -Name "Auth Service" -Keep` streams backend output.

### Linux / macOS (Bash)

```bash
cd <repo>
chmod +x ./Launch.sh   # first run
./Launch.sh
```

What this does:
- Detects your package manager (apt, dnf, pacman, zypper, or Homebrew) and installs Go/Node.js/npm/gcc when `sudo` privileges are available. The script also works when run as root (no sudo required).
- Ensures the SQLite DB file exists, downloads Go modules, installs repo-level Node packages plus `frontend/` dependencies.
- Starts every Go microservice plus `npm run dev`, writing logs to `./logs/*.log`, and opens the browser via `xdg-open`/`open`/`sensible-browser`.

Useful flags:
- `--skip-install`, `--skip-frontend`, `--host <host>`, `--port <port>`.

Log streaming example: `tail -f logs/Auth_Service.log`.

> Tip: If you want to keep the launchers running while inspecting logs, run them inside a dedicated terminal window or `tmux`/`screen` pane.

## Manual Terminal Setup (Fallback)
If you prefer to run services manually or need to debug without the launchers, open four terminals and replace `<repo>` with your clone path.

### Terminal 1: Auth Service (Port 8081)
```powershell
cd <repo>\services\auth
$env:DATABASE_PATH="<repo>\social_network.db"
go run main.go
```
Expected output:
```
Auth Service starting on :8081
Database initialized successfully
```

### Terminal 2: Chat Service (Port 8085)
```powershell
cd <repo>\services\chat
$env:DATABASE_PATH="<repo>\social_network.db"
go run main.go
```
Expected output:
```
Chat Service starting on :8085
WebSocket hub running
```

### Terminal 3: Notification Service (Port 8086)
```powershell
cd <repo>\services\notifications
$env:DATABASE_PATH="<repo>\social_network.db"
go run main.go
```
Expected output:
```
Notification Service starting on :8086
WebSocket hub running
```

### Terminal 4: Frontend Dev Server (Port 5173)
```bash
cd <repo>/frontend
npm run dev
```
Expected output:
```
VITE ready in XXX ms
  ->  Local:   http://localhost:5173/
```

## Verification Checklist
- [ ] Auth Service running on `http://localhost:8081`
- [ ] Chat Service running on `http://localhost:8085`
- [ ] Notification Service running on `http://localhost:8086`
- [ ] Frontend running on `http://localhost:5173`

## Quick Test
1. Open the browser at `http://localhost:5173`.
2. Log in with the seeded test user or register a new account.
3. Browser console should show:
   ```
   Login response: {success: true, data: {...}}
   Connecting to WebSocket servers...
   Chat WebSocket connected
   Notification WebSocket connected
   ```
4. Confirm notification bell and chat widget are active in the UI.

## Troubleshooting
- **Port already in use**  
  `netstat -ano | findstr :8081` (Windows) or `lsof -i :8081` (Unix) to find the PID, then terminate it (`taskkill /PID <pid> /F` or `kill <pid>`).
- **Database not found**  
  `Test-Path "<repo>\social_network.db"` (PowerShell) or `ls <repo>/social_network.db` (Bash). Create it manually with `New-Item`/`touch` if necessary.
- **WebSocket connection failed**  
  Ensure Chat and Notifications services are running, and verify `.env` contains:
  ```
  VITE_CHAT_WS_URL=ws://localhost:8085/ws
  VITE_NOTIFICATION_WS_URL=ws://localhost:8086/ws
  ```
- **Frontend won't start**  
  Run `npm install` inside both `<repo>` and `<repo>/frontend`, remove `node_modules` if necessary, then retry `npm run dev`.

## Stop All Services
- Launchers: press `Ctrl+C` once in the launcher terminal; they will clean up every child process.
- Manual run: press `Ctrl+C` in each terminal (or stop the individual Go processes and the Vite dev server).

## Docker Alternative (Optional)
```bash
cd <repo>
docker-compose up
```
Use this for a containerized setup when Docker is available and you prefer not to run the Go processes directly.
