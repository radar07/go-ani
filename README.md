# go-ani

Watch anime from your terminal. Built with Go, inspired by [ani-cli](https://github.com/pystardust/ani-cli).

## Features

**Search** — Find anime by title via GraphQL API
**Stream** — Play episodes directly in mpv or VLC
**Interactive TUI** — Arrow-key navigation, filtering, episode picker (powered by [bubbletea](https://github.com/charmbracelet/bubbletea))
**Quality selection** — Choose from available streams (1080p, 720p, etc.) or auto-select best/worst
**Sub & Dub** — Switch between subbed and dubbed versions
**Post-playback menu** — Next/previous episode, replay, pick another episode

## Requirements

**Go 1.26+** (to build)
**mpv** or **VLC** (video player)

### Install a video player

# macOS

```bash
brew install mpv
```

# Linux (Debian/Ubuntu)

```bash
sudo apt install mpv
```

# Linux (Arch)

```bash
sudo pacman -S mpv
```



## Installation

### From source

```bash
git clone https://github.com/radar07/go-ani.git
cd go-ani
go build -o go-ani .
./go-ani --help
```

### With go install

```bash
go install github.com/radar07/go-ani@latest
```

## Usage

### Search for anime

```bash
# Interactive: opens TUI search input
go-ani search

# Direct: search by name
go-ani search "one punch man"
```

### Play an episode

```bash
# Full interactive flow: search → select → pick episode → play
go-ani play

# Search and select interactively, then play
go-ani play "one punch man"

# Skip straight to episode 1
go-ani play "one punch man" -e 1

# Dubbed, 720p quality, using VLC
go-ani play "one punch man" -e 1 --dub -q 720 -p vlc
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| --player | -p | Video player (mpv / vlc) | mpv |
| --quality | -q | Quality (best / worst / 1080 / 720) | best |
| --dub | | Use dubbed version | false (sub) |
| --episode | -e | Episode number (play/download only) | interactive |

## TUI Keybindings

| Key | Action |
|-----|--------|
| ↑ ↓ | Navigate lists |
| Enter | Select / confirm |
| / | Filter results or episodes |
| Esc | Go back |
| q | Quit |
| Ctrl+C | Force quit |

## Configuration

Environment variables (set in your shell profile):

| Variable | Description | Default |
|----------|-------------|---------|
| ANI_PLAYER | Default video player | mpv |
| ANI_QUALITY | Default quality preference | best |
| ANI_MODE | Default mode (sub / dub) | sub |
| ANI_DOWNLOAD_DIR | Download directory | . |
| ANI_NO_DETACH | Keep player in foreground | false |

## Roadmap

- [ ] Watch history + continue command
- [ ] Episode download support
- [ ] IINA player support (macOS)
- [ ] Homebrew tap

## Credits

Inspired by [ani-cli](https://github.com/pystardust/ani-cli)
TUI powered by [Charm](https://charm.sh) ([bubbletea](https://github.com/charmbracelet/bubbletea), [bubbles](https://github.com/charmbracelet/bubbles), [lipgloss](https://github.com/charmbracelet/lipgloss))

## License

[MIT](LICENSE)
