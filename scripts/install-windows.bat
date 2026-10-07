@echo off
REM Double-clickable wrapper — launches the PowerShell installer elevated
powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Process powershell -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File \"%~dp0install-loki-system.ps1\" -Channel stable' -Verb runAs"
pause
