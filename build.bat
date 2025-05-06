@echo off
setlocal

REM Create the build directory if it doesn't exist
if not exist build (
    mkdir build
)

REM Build the backend binary
echo Building backend...
cd backend
go build -o ..\build\backend.exe .
if %ERRORLEVEL% neq 0 (
    echo Backend build failed.
    exit /b %ERRORLEVEL%
)
cd ..

REM Publish the frontend binary
echo Building frontend...
dotnet publish .\frontend\BlazorApp\ -r win-x64 -c Release --self-contained true /p:PublishSingleFile=true -o .\build\frontend
if %ERRORLEVEL% neq 0 (
    echo Frontend build failed.
    exit /b %ERRORLEVEL%
)

echo Creating README.txt in .\build...

(
	echo This directory contains the built binaries for the project.
	echo The backend binary is named backend.exe and the frontend binary is in the frontend directory as BlazorApp.exe.
    echo They must be run from their own directory, seperatly.
	echo Backend takes --port=<port> as an argument to specify the port to run on.
) > ".\build\README.txt"

echo Build completed successfully!
exit /b 0