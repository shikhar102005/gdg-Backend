@echo off

:: 0. Always execute from the project root directory
cd /d "%~dp0"

echo =====================================================================
echo   Campus Sensor Monitoring ^& Alerts - Startup Script
echo =====================================================================
echo.

:: 1. Check Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] 'go' is not installed or not in your PATH.
    echo Please install Go 1.22+ from https://go.dev/dl/
    pause
    exit /b 1
)

:: 2. Check Node / npm
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] 'npm' is not installed or not in your PATH.
    echo Please install Node.js v18+ from https://nodejs.org/
    pause
    exit /b 1
)

:: 3. Check frontend node_modules
if not exist "web\node_modules" (
    echo [INFO] Installing frontend dependencies in web...
    pushd web
    call npm install
    popd
)

:: 4. Terminate any stale processes on ports 8081 and 5173
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8081 ^| findstr LISTENING') do (
    echo [INFO] Cleaning up stale process on port 8081 - PID: %%a
    taskkill /F /PID %%a >nul 2>nul
)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :5173 ^| findstr LISTENING') do (
    echo [INFO] Cleaning up stale process on port 5173 - PID: %%a
    taskkill /F /PID %%a >nul 2>nul
)

echo [1/3] Starting Go Backend Server on http://localhost:8081...
start "Campus Sensor Backend (Port 8081)" cmd /k "cd /d ""%CD%"" && go run ./cmd/server"

echo [2/3] Starting React / Vite Frontend on http://localhost:5173...
start "Campus Sensor Web UI (Port 5173)" cmd /k "cd /d ""%CD%\web"" && npm run dev"

echo [3/3] Waiting for servers to initialize...
ping -n 4 127.0.0.1 >nul

echo.
echo =====================================================================
echo   Application Started Successfully!
echo =====================================================================
echo   Frontend Dashboard : http://localhost:5173
echo   Backend API Root   : http://localhost:8081
echo   Health Check       : http://localhost:8081/healthz
echo.
echo   Demo Login Credentials (Organizer):
echo     Email    : organizer@campus.edu
echo     Password : CampusOrganizerSecure2026!
echo.
echo   To stop all servers: run 'stop.bat' or close the terminal windows.
echo =====================================================================
echo.

:: Open default browser
start http://localhost:5173

pause
