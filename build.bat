:: usage:
:: build.bat [--debug] [--skip-compression]
:: options:
:: --debug: compile as console application instead of windows application
:: --skip-compression: skip upx compression
:: --devtools: make patcher open devtools on ggb start

@echo off
setlocal EnableDelayedExpansion

set "DEBUG=0"
set "SKIP_COMPRESSION=0"
set "DEVTOOLS=0"

:checkArgs
if "%~1"=="" goto checkArgsDone

if /I "%~1"=="--debug" set "DEBUG=1"
if /I "%~1"=="--skip-compression" set "SKIP_COMPRESSION=1"
if /I "%~1"=="--devtools" set "DEVTOOLS=1"

shift
goto checkArgs

:checkArgsDone

set "BASE_EXE_NAME=ggb_patcher"
set "VERSION=0.1.1"

call :build "windows" "amd64" "x64" "installer"
call :build "windows" "amd64" "x64" "uninstaller"

call :build "windows" "arm64" "arm64" "installer"
call :build "windows" "arm64" "arm64" "uninstaller"

goto :end
:build 

mkdir bin > nul 2>&1

cd bin

set "name=%BASE_EXE_NAME%_%VERSION%_%~1_%~3_%~4"
set "GOOS=%~1"
set "GOARCH=%~2"
set "CGO_ENABLED=0"
if "%DEBUG%"=="1" ( set "HFLAG=" ) else ( set "HFLAG=-H windowsgui" )

del "%name%.exe" > nul 2>&1
del temp.exe > nul 2>&1

go build -ldflags="-s -w !HFLAG! -X main.BUILD_TYPE=%~4 -X main.DEVTOOLS=%DEVTOOLS%" -o temp.exe ../main.go

if "%GOARCH%"=="arm64" (
    ren temp.exe %name%.exe
) else (
    if "%SKIP_COMPRESSION%"=="1" (
        ren temp.exe %name%.exe 
    ) else (
        upx --brute -9 -o %name%.exe temp.exe > nul
        del temp.exe
    )
)

echo Built %~4 executable for %~3 %~1

cd ..

exit /B 0

:end
