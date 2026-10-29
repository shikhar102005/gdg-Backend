# =====================================================================
#   Campus Sensor Monitoring & Alerts - PowerShell Startup Script
# =====================================================================

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Campus Sensor Monitoring & Alerts Engine - Starting Services       " -ForegroundColor Yellow
Write-Host "=====================================================================" -ForegroundColor Cyan

# 1. Check prerequisites
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] 'go' is not found in PATH. Please install Go from https://go.dev/dl/" -ForegroundColor Red
    exit 1
}

if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] 'npm' is not found in PATH. Please install Node.js from https://nodejs.org/" -ForegroundColor Red
    exit 1
}

# 2. Check web/node_modules
if (-not (Test-Path "web\node_modules")) {
    Write-Host "[INFO] Installing web dependencies..." -ForegroundColor Green
    Push-Location web
    npm install
    Pop-Location
}

# 3. Terminate stale processes on ports 8081 & 5173 if present
$ports = @(8081, 5173)
foreach ($port in $ports) {
    $conns = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conns) {
        $pidsToKill = $conns.OwningProcess | Select-Object -Unique
        foreach ($p in $pidsToKill) {
            Write-Host "[INFO] Cleaning up process $p on port $port..." -ForegroundColor Yellow
            Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
        }
    }
}

# 4. Start Go Backend
Write-Host "[1/3] Starting Go Backend Server on http://localhost:8081..." -ForegroundColor Green
$backendProc = Start-Process powershell -ArgumentList "-NoExit", "-Command", "Write-Host '=== Campus Sensor Backend Server (Port 8081) ===' -ForegroundColor Cyan; go run ./cmd/server" -PassThru

# 5. Start Vite Frontend
Write-Host "[2/3] Starting Vite Frontend on http://localhost:5173..." -ForegroundColor Green
$frontendProc = Start-Process powershell -ArgumentList "-NoExit", "-Command", "Write-Host '=== Campus Sensor Vite Dev Server (Port 5173) ===' -ForegroundColor Cyan; Set-Location web; npm run dev" -PassThru

# 6. Wait for backend readiness
Write-Host "[3/3] Waiting for backend readiness..." -ForegroundColor Green
$healthy = $false
for ($i = 0; $i -lt 15; $i++) {
    Start-Sleep -Seconds 1
    try {
        $res = Invoke-RestMethod -Uri "http://localhost:8081/healthz" -TimeoutSec 2 -ErrorAction Stop
        if ($res.status -eq "healthy") {
            $healthy = $true
            break
        }
    } catch {
        # continue waiting
    }
}

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Services Started Successfully!                                    " -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Web Dashboard UI   : http://localhost:5173" -ForegroundColor White
Write-Host "  Backend API Root   : http://localhost:8081" -ForegroundColor White
Write-Host "  Health Probe       : http://localhost:8081/healthz" -ForegroundColor White
Write-Host ""
Write-Host "  Demo Login Credentials (Organizer):" -ForegroundColor Yellow
Write-Host "    Email    : organizer@campus.edu" -ForegroundColor White
Write-Host "    Password : CampusOrganizerSecure2026!" -ForegroundColor White
Write-Host ""
Write-Host "  To stop services, run .\stop.ps1 or close the opened terminal windows." -ForegroundColor Gray
Write-Host "=====================================================================" -ForegroundColor Cyan

# Open default browser
Start-Process "http://localhost:5173"
