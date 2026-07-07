@echo off
REM Build tt.exe with version metadata from git (for cmd.exe).
setlocal EnableExtensions

cd /d "%~dp0.."

set "VERSION=dev"
for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"

set "COMMIT=none"
for /f "delims=" %%i in ('git rev-parse HEAD 2^>nul') do set "COMMIT=%%i"

for /f "delims=" %%i in ('powershell -NoProfile -Command "(Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')"') do set "DATE=%%i"

go build -ldflags "-s -w -X github.com/stage3technical/time-tracker-cli/internal/version.Version=%VERSION% -X github.com/stage3technical/time-tracker-cli/internal/version.Commit=%COMMIT% -X github.com/stage3technical/time-tracker-cli/internal/version.Date=%DATE%" -o tt.exe ./cmd/tt
if errorlevel 1 exit /b 1

tt.exe version
