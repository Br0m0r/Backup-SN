# Starting All Services - Step by Step

## Prerequisites
- Go installed
- Node.js and npm installed
- Database file exists at: `C:\Users\Morie\Desktop\Social\social-network\social_network.db`
- All database migrations applied (migrations 000001-000015 include posts tables)

## Terminal Setup (5 terminals needed)

### Terminal 1: Auth Service (Port 8081)
```powershell
cd C:\Users\Morie\Desktop\Social\social-network\services\auth
$env:DATABASE_PATH="C:\Users\Morie\Desktop\Social\social-network\social_network.db"
go run main.go
```

**Expected Output:**
```
Auth Service starting on :8081
Database initialized successfully
```

### Terminal 2: Chat Service (Port 8085)
```powershell
cd C:\Users\Morie\Desktop\Social\social-network\services\chat
$env:DATABASE_PATH="C:\Users\Morie\Desktop\Social\social-network\social_network.db"
go run main.go
```

**Expected Output:**
```
Chat Service starting on :8085
WebSocket hub running
```

### Terminal 3: Posts Service (Port 8083)
```powershell
cd C:\Users\Morie\Desktop\Social\social-network\services\posts
$env:DATABASE_PATH="C:\Users\Morie\Desktop\Social\social-network\social_network.db"
go run main.go
```

**Expected Output:**
```
Connected to database: C:\Users\Morie\Desktop\Social\social-network\social_network.db
Post Service starting on port :8083
```

### Terminal 4: Notification Service (Port 8086)
```powershell
cd C:\Users\Morie\Desktop\Social\social-network\services\notifications
$env:DATABASE_PATH="C:\Users\Morie\Desktop\Social\social-network\social_network.db"
go run main.go
```

**Expected Output:**
```
Notification Service starting on :8086
WebSocket hub running
```

### Terminal 5: Frontend Dev Server (Port 5173)
```bash
cd C:\Users\Morie\Desktop\Social\social-network\frontend
npm run dev
```

**Expected Output:**
```
VITE ready in XXX ms
➜  Local:   http://localhost:5173/
```

## Verification Checklist

After starting all services, verify:

- [ ] Auth Service running on `http://localhost:8081`
- [ ] Chat Service running on `http://localhost:8085`
- [ ] Posts Service running on `http://localhost:8083`
- [ ] Notification Service running on `http://localhost:8086`
- [ ] Frontend running on `http://localhost:5173`

## Quick Test

1. **Open Browser:** `http://localhost:5173`

2. **Login with test credentials** (or register new user)

3. **Check Browser Console** - Should see:
   ```
   Login response: {success: true, data: {...}}
   Connecting to WebSocket servers...
   Chat WebSocket connected
   Notification WebSocket connected
   ```

4. **Check UI:**
   - Notification bell visible in header
   - "🟢 Connected" status on home page
   - Chat demo functional

## Troubleshooting

### Port Already in Use
If you get "address already in use" error:

**Find process using port:**
```powershell
netstat -ano | findstr :8081
```

**Kill the process:**
```powershell
taskkill /PID <PID_NUMBER> /F
```

### Database Not Found
Ensure the DATABASE_PATH is correct:
```powershell
# Check if file exists
Test-Path "C:\Users\Morie\Desktop\Social\social-network\social_network.db"
# Should return True
```

### WebSocket Connection Failed
1. Check Chat and Notification services are running
2. Verify `.env` has correct URLs:
   ```
   VITE_CHAT_WS_URL=ws://localhost:8085/ws
   VITE_NOTIFICATION_WS_URL=ws://localhost:8086/ws
   ```
3. Check browser console for errors

### Frontend Won't Start
```bash
# Install dependencies if needed
npm install

# Clear cache and restart
rm -rf node_modules
npm install
npm run dev
```

## Stop All Services

Press `Ctrl+C` in each terminal to gracefully stop services.

## Production Deployment

For production, consider:
1. Using environment variables for all configuration
2. Running services with process managers (pm2, systemd)
3. Using reverse proxy (nginx) for WebSocket connections
4. Implementing proper logging and monitoring
5. Setting up SSL/TLS for secure WebSocket connections (wss://)

## Docker Alternative (Optional)

If you have Docker, you can use docker-compose:
```bash
cd C:\Users\Morie\Desktop\Social\social-network
docker-compose up
```

This will start all services automatically (if docker-compose.yml is properly configured).
