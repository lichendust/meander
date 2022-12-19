@echo off
setlocal enabledelayedexpansion

set build_dir=build

echo [building]

call :build windows amd64 ".exe" || exit /B

call :build linux amd64          || exit /B
call :build linux arm64          || exit /B

call :build darwin amd64         || exit /B
call :build darwin arm64         || exit /B

call :build freebsd amd64 || exit /B
call :build netbsd  amd64 || exit /B
call :build openbsd amd64 || exit /B

exit /B 0

:build
echo %1 %2
set name=%1_%2

if %1 == darwin (
	set name=macOS_

	if %2 == arm64 (
		set name=!name!M1
	) else (
		set name=!name!Intel
	)
)

if not exist "%build_dir%\%name%" md "%build_dir%\%name%"

set GOOS=%1
set GOARCH=%2

go build -ldflags "-s -w" -trimpath -o %build_dir%\%name%\meander%3 ./source

if %errorlevel% neq 0 exit 1
exit /B 0