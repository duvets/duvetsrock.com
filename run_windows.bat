@echo off

set GOPATH=%~dp0

FOR /f %%f IN ('DIR /b /od %USERPROFILE%\AppData\Local\GitHub\PortableGit*') DO @SET last=%%f

set PATH=%USERPROFILE%\AppData\Local\GitHub\%last%\bin;%PATH%
set TERM=msys
doskey test_server=go run src\duvetsrock.com\builder\builder.go

echo To run a test server, type test_server
echo ""

cmd.exe
