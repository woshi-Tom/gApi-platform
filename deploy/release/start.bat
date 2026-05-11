@echo off
REM gAPI Platform - Quick Start Script (Windows)
cd /d "%~dp0"

echo ========================================
echo   gAPI Platform - Starting...
echo ========================================
echo.

REM Check Docker
where docker >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Docker is not installed.
    echo Please install Docker Desktop from: https://docs.docker.com/get-docker/
    pause
    exit /b 1
)

REM Check .env
if not exist ".env" (
    echo [1/3] Creating .env from template...
    copy .env.example .env >nul
    echo   -^> Please edit .env and set your passwords before continuing!
    echo   -^> After editing, run this script again.
    echo.
    pause
    exit /b 0
)

REM Warn about default passwords
findstr /C:"CHANGE_ME" .env >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    echo Warning: You still have default passwords in .env
    echo Press Ctrl+C to abort, or wait 5 seconds to continue...
    timeout /t 5 /nobreak >nul
)

echo [2/3] Starting all services with Docker Compose...
docker compose up -d

echo.
echo [3/3] Services starting up...
echo.
echo ========================================
echo   All services are being started!
echo ========================================
echo.
echo   Backend API:   http://localhost:8080
echo   API Docs:      http://localhost:8080/swagger/index.html
echo   Frontend:      http://localhost:5173
echo   Admin Panel:   http://localhost:5174
echo.
echo   Check status:  docker compose ps
echo   View logs:     docker compose logs -f
echo   Stop all:      docker compose down
echo.
pause
