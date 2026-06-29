@echo off
REM Go Chrome Project Build Script for Windows

echo ====================================
echo Go Chrome Project Build Script
echo ====================================
echo.

REM 检查Go是否安装
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed. Please install Go first.
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo [INFO] Go is installed: %GO_VERSION%
echo.

REM 处理命令行参数
set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=help

if "%COMMAND%"=="init" goto init_backend
if "%COMMAND%"=="build" goto build_all
if "%COMMAND%"=="backend" goto build_backend
if "%COMMAND%"=="extension" goto build_extension
if "%COMMAND%"=="run" goto run_backend
if "%COMMAND%"=="help" goto show_help
if "%COMMAND%"=="--help" goto show_help
if "%COMMAND%"=="-h" goto show_help

echo [ERROR] Unknown command: %COMMAND%
goto show_help

:init_backend
echo [INFO] Initializing backend project...
cd backend

if not exist "go.sum" (
    echo [WARN] Go modules not initialized. Running 'go mod tidy'...
    go mod tidy
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to initialize Go modules
        cd ..
        exit /b 1
    )
    echo [SUCCESS] Go modules initialized successfully
) else (
    echo [SUCCESS] Go modules already initialized
)

echo [WARN] Generating Swagger documentation...
where swag >nul 2>nul
if %errorlevel% equ 0 (
    swag init -g main.go
    if %errorlevel% equ 0 (
        echo [SUCCESS] Swagger documentation generated
    ) else (
        echo [WARN] Failed to generate Swagger documentation
    )
) else (
    echo [WARN] Swagger CLI not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest
)

cd ..
goto end

:build_backend
echo [INFO] Building backend project...
cd backend

if not exist "..\dist" mkdir "..\dist"

echo [WARN] Building backend executable...
go build -o ..\dist\go_chrome_server.exe main.go

if %errorlevel% equ 0 (
    echo [SUCCESS] Backend built successfully: dist\go_chrome_server.exe
) else (
    echo [ERROR] Failed to build backend
    cd ..
    exit /b 1
)

cd ..
goto end

:build_extension
echo [INFO] Building Chrome extension...

if not exist "dist" mkdir "dist"
if not exist "dist\extension" mkdir "dist\extension"

echo [WARN] Copying extension files...
xcopy /E /I /Y extension\* dist\extension\ >nul

echo [SUCCESS] Chrome extension built successfully: dist\extension\
goto end

:build_all
call :build_backend
call :build_extension
goto end

:run_backend
echo [INFO] Starting backend server...
cd backend
go run main.go
goto end

:show_help
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   init       Initialize the project (install dependencies)
echo   build      Build both backend and extension
echo   backend    Build only the backend
echo   extension  Build only the extension
echo   run        Run the backend server
echo   help       Show this help message
echo.
echo Examples:
echo   build.bat init     # Initialize project
echo   build.bat build    # Build everything
echo   build.bat run      # Run backend server
echo.
goto end

:end
echo.
echo [SUCCESS] Operation completed successfully!
echo.