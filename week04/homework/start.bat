@echo off
set "PROJECT_DIR=%~dp0"

start "AI Study Agent - Backend" /D "%PROJECT_DIR%server" cmd /k "go run ."
start "AI Study Agent - Frontend" /D "%PROJECT_DIR%client" cmd /k "npm run dev"

timeout /t 2 /nobreak >nul
start "" "http://localhost:5173"
