package player

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type vlcPlayer struct {
	binary string
}

func newVLC() (*vlcPlayer, error) {
	var candidates []string
	switch platform() {
	case "darwin":
		candidates = []string{
			"/Applications/VLC.app/Contents/MacOS/VLC",
			"vlc",
		}
	default:
		candidates = []string{"vlc"}
	}

	bin, err := findBinary(candidates...)
	if err != nil {
		return nil, fmt.Errorf("vlc not found: %w", err)
	}
	return &vlcPlayer{binary: bin}, nil
}

func (v *vlcPlayer) Name() string { return "vlc" }

func (v *vlcPlayer) Play(opts PlayOptions) error {
	args := []string{}

	if opts.Referrer != "" {
		args = append(args, fmt.Sprintf("--http-referrer=%s", opts.Referrer))
	}
	if opts.Title != "" {
		args = append(args, fmt.Sprintf("--meta-title=%s", opts.Title))
	}
	args = append(args, "--play-and-exit")
	args = append(args, opts.URL)

	cmd := exec.Command(v.binary, args...)

	if opts.NoDetach {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	// Detached mode
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start vlc: %w", err)
	}

	fmt.Printf("▶️  VLC started (pid %d)\n", cmd.Process.Pid)
	go cmd.Wait()
	return nil
}
