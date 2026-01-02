## Description
A CLI program that can plays Hianime URLs directly in MPV. It automatically fetches matching subtitles from the [Jimaku API](https://github.com/Rapptz/jimaku).

## Disclaimer
This project was built for educational purposes to practice web scraping and Go. I do not host any content, nor am I affiliated with the services being scraped. This tool simply accesses publicly available data. Please use it responsibly and support the original creators.

## Requirements
- MPV 0.35.0+
- yt-dlp 2025.12.08+

## Usage
- Windows
`hianime-windows-amd64.exe`
- Linux
`./hianime-linux-amd64`

## Build
- Windows
`GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o hianime-windows-amd64.exe`
- Linux
`GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hianime-linux-amd64`

## Config
User can customize to their personal preference in `config.json`

Simple table for explanations.

| Name | Description | Default |
| ---- | ---- | ---- |
| jimaku_enable | Toggle Jimaku API integration on or off. | true |
| auto_selectserver | Automatically select the first available server. | true |
| mpv_path | Custom path to your MPV executable (leave empty to use system default). | "" |
| english_only | Only load English subtitles; ignore other languages. | true |

## Troubleshoot
- Jimaku API issues: Get your key from [jimaku.cc](https://jimaku.cc) and add it to environment variables (e.g. JIMAKU_API_KEY=yourkey).

## Thanks to
- [MediaVanced](https://github.com/yogesh-hacker/MediaVanced)

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
