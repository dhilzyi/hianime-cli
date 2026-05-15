## Description
A simple CLI tool to stream anime from supported providers using [MPV](https://mpv.io/) player.

## Disclaimer
This project was built for educational purposes to practice web scraping and Go. I do not host any content, nor am I affiliated with the services being scraped. This tool simply accesses publicly available data. Please use it responsibly and support the original creators.

## Requirements
- MPV >= 0.35.0
  
## Installation

### Option 1 — Install with Go (recommended)

```bash
go install github.com/dhilzyi/hianime-cli/cmd/hianime-cli@latest
```

This will install hianime-cli into your $GOPATH/bin or $HOME/go/bin.

### Option 2 — Download prebuilt binary

Download from the Releases section and run it directly.
### Usage

After installation:

```bash
hianime-cli
```

Or if running manually:
#### Windows

```bash
hianime-windows-amd64.exe
```

#### Linux

```bash
./hianime-linux-amd64
```

### Build from source
#### Windows

```bash
GOOS=windows go build -C ./cmd/hianime-cli -ldflags="-s -w" -o hianime-windows.exe
```
#### Linux

```bash
GOOS=linux go build -C ./cmd/hianime-cli -ldflags="-s -w" -o hianime-linux
```


## Config
User can customize to their personal preference in `config.json`

Config path file can be found in:
| Platform | Path|
| --- | --- |
| Windows | %APPDATA%/hianimecli/config.json|
| Linux | ~/.config/hianimecli/config.json|


Field table explanations.

| Field | Description | Default Value |
| ---- | ---- | ---- |
| jimaku_enable | Toggle Jimaku API integration on or off. | false |
| auto_selectserver | Automatically select the first available server. | true |
| mpv_path | Custom path to your MPV executable (leave empty to use system default). | "" |
| english_only | Only load English subtitles; ignore other languages. | true |
| sort_type | Order type for search results. User can change the order as they like. | {"TV", "Movie", "OVA", "Special", "ONA", "Music"} |
| local_version | State local version control for determining version additional files | Parsed from static file version.txt|

## Troubleshoot
- Jimaku API issues: Get your key from [jimaku.cc](https://jimaku.cc) and add it to environment variables (e.g. JIMAKU_API_KEY=yourkey).
- MPV not found: Add your mpv program into PATH system.
## Acknowledgment
- [MediaVanced](https://github.com/yogesh-hacker/MediaVanced)
    Help me to build host video extractor.
- [anime-extensions](https://github.com/yuzono/anime-extensions)
    Inspirator.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
