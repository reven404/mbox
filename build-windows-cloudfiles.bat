@echo off
REM Build script for Windows Cloud Files API support

echo Building mdriver with Windows Cloud Files API support...
echo.

REM Check Windows version
echo Checking Windows version...
for /f "tokens=3" %%i in ('reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion" /v ReleaseId ^| findstr ReleaseId') do set RELEASE_ID=%%i
echo Windows Release ID: %RELEASE_ID%

if %RELEASE_ID% LSS 1809 (
    echo WARNING: Windows 10 version 1809 or later required for Cloud Files API
    echo Current version may not support all features
)

echo.

REM Build the Cloud Files bridge library
echo Building Cloud Files bridge library...
cd providers\windows
if exist Makefile (
    nmake bridge
    if errorlevel 1 (
        echo ERROR: Failed to build Cloud Files bridge
        echo Make sure Visual Studio build tools are installed
        pause
        exit /b 1
    )
) else (
    echo WARNING: Windows Makefile not found, assuming bridge is already built
)
cd ..\..

echo.

REM Set build environment
echo Setting up build environment...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

echo CGO_ENABLED=%CGO_ENABLED%
echo GOOS=%GOOS%
echo GOARCH=%GOARCH%

echo.

REM Build with Cloud Files support
echo Building Go application with Cloud Files support...
go build -tags cloudfiles -ldflags="-X main.version=1.0.0-cloudfiles -X main.commit=%date%" -o mdriver-cloudfiles.exe .

if errorlevel 1 (
    echo ERROR: Go build failed
    echo Make sure Go is installed and CGO is working properly
    pause
    exit /b 1
)

echo.
echo SUCCESS: Built mdriver-cloudfiles.exe with Windows Cloud Files API support
echo.

REM Test the build
echo Testing Cloud Files API support...
mdriver-cloudfiles.exe --platform-info | findstr "windows-cloudfiles"
if errorlevel 1 (
    echo WARNING: Cloud Files API capability not detected
) else (
    echo SUCCESS: Cloud Files API capability detected
)

echo.

REM Show usage instructions
echo Usage Instructions:
echo   1. Run as Administrator (required for sync root registration)
echo   2. Configure test-windows-cloudfiles.yaml with your settings
echo   3. Execute: mdriver-cloudfiles.exe --mount --config test-windows-cloudfiles.yaml
echo   4. Check Windows Explorer for Smart Folder integration
echo.

echo Build completed successfully!
pause