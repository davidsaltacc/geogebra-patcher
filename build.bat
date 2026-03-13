@echo off
setlocal EnableDelayedExpansion

set "BASE_EXE_NAME=ggb_patcher"
set "DEBUG=1"
set "VERSION=0.1"

call :build "windows" "386" "x32" "installer"
call :build "windows" "386" "x32" "uninstaller"

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

go build -ldflags="-s -w !HFLAG! -X main.BUILD_TYPE=%~4" -o temp.exe ../main.go

if "%GOARCH%"=="arm64" ( 
    ren temp.exe %name%.exe
) else ( 
    upx -9 -o %name%.exe temp.exe > nul
    del temp.exe
)

echo Built %~4 executable for %~3 %~1

cd ..

exit /B 0

:end
