@echo off
title FoodBridge Starter
color 0A
echo =============================================
echo        FoodBridge - Starting App
echo =============================================
echo.

echo [1/3] Starting MongoDB...
tasklist /FI "IMAGENAME eq mongod.exe" 2>NUL | find /I "mongod.exe" >NUL
if %ERRORLEVEL%==0 (
    echo        MongoDB already running
) else (
    start "MongoDB" cmd /k mongod
    echo        MongoDB started
)
timeout /t 3 /nobreak >nul

echo.
echo [2/3] Starting Backend...
cd /d "%~dp0backend"
start "FoodBridge Backend" cmd /k "go run main.go"
echo        Backend starting on http://localhost:8080
timeout /t 6 /nobreak >nul

echo.
echo [3/3] Starting Frontend...
cd /d "%~dp0frontend"
start "FoodBridge Frontend" cmd /k "python -m http.server 3000"
echo        Frontend starting on http://localhost:3000
timeout /t 2 /nobreak >nul

echo.
echo =============================================
echo   All services started successfully!
echo =============================================
echo.
echo   Frontend : http://localhost:3000
echo   Backend  : http://localhost:8080
echo.
echo   Admin Login:
echo   Email    : admin@foodbridge.com
echo   Password : admin123
echo =============================================
echo.
echo Opening browser...
timeout /t 2 /nobreak >nul
start http://localhost:3000
pause