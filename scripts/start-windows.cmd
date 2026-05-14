@echo off
set KIRO_API_KEY=ksk_YOUR_KEY_HERE
set KIRO_CLI_PATH=C:\Users\user\AppData\Local\Kiro-Cli\kiro-cli.exe
set LISTEN_ADDR=:8080
cd /d %~dp0..
bin\kiro-proxy.exe >> logs\proxy.log 2>&1
