@echo off
setlocal
go mod tidy
go build -ldflags "-s -w -H=windowsgui" -o KadastrApp.exe kadastr_gui_fyne.go
echo Done: KadastrApp.exe
