@echo off
cd /d "%~dp0"

echo =====================================================================
echo   Stopping Campus Sensor Backend and Frontend Services...
echo =====================================================================

:: Kill processes listening on port 8081 (Backend)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8081 ^| findstr LISTENING') do (
    echo [INFO] Stopping backend process - PID: %%a
    taskkill /F /PID %%a >nul 2>nul
)

:: Kill processes listening on port 5173 (Vite frontend)
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :5173 ^| findstr LISTENING') do (
    echo [INFO] Stopping frontend process - PID: %%a
    taskkill /F /PID %%a >nul 2>nul
)

echo.
echo Done! All services stopped.
pause
