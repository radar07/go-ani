package player

import (
	"fmt"
	"os/exec"
	"runtime"
)

// PlayOptions holds configuration for playing a video
type PlayOptions struct {
	URL          string
	Title        string
	Referrer     string
	SubtitlePath string
	NoDetach     bool
}

// Player can launch a video player
type Player interface {
	Play(opts PlayOptions) error
	Name() string
}

// NewPlayer creates a player by name. Supported: "mpv", "vlc"
func NewPlayer(name string) (Player, error) {
	switch name {
	case "mpv":
		return newMPV()
	case "vlc":
		return newVLC()
	default:
		return nil, fmt.Errorf("unsupported player: %s", name)
	}
}

// findBinary locates an executable, checking platform-specific paths
func findBinary(candidates ...string) (string, error) {
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("none of %v found in PATH", candidates)
}

// platform returns the current OS
func platform() string {
	return runtime.GOOS
}
