## Description ##
CLI program that can play hianime url to your mpv.

## Disclaimer ##
This tool is for educational and personal use only. It demonstrates web scraping and media integration techniques—use responsibly, respect site terms of service, and be mindful of copyrights. I'm not affiliated with any third-party sites or services.

## Requirements ##
- MPV 0.35.0+
- yt-dlp 2025.12.08+

## Usage ##
- Windows
`hianime-windows-amd64.exe`
- Linux
`./hianime-linux-amd64`

## Build ##
- Windows
`GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o hianime-windows-amd64.exe`
- Linux
`GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hianime-linux-amd64`
