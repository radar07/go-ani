package player

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type mpvPlayer struct {
	binary string
}

func newMPV() (*mpvPlayer, error) {
	bin, err := findBinary("mpv")
	if err != nil {
		return nil, fmt.Errorf("mpv not found: %w", err)
	}
	return &mpvPlayer{binary: bin}, nil
}

func (m *mpvPlayer) Name() string { return "mpv" }

func (m *mpvPlayer) Play(opts PlayOptions) error {
	args := []string{}

	if opts.Title != "" {
		args = append(args, fmt.Sprintf("--force-media-title=%s", opts.Title))
	}
	if opts.Referrer != "" {
		args = append(args, fmt.Sprintf("--referrer=%s", opts.Referrer))
	}
	if opts.SubtitlePath != "" {
		args = append(args, fmt.Sprintf("--sub-file=%s", opts.SubtitlePath))
	}

	args = append(args, opts.URL)

	cmd := exec.Command(m.binary, args...) //nolint:gosec

	if opts.NoDetach {
		// Attached mode: inherit stdio, wait for exit
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	// Detached mode: start in new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}

	fmt.Printf("▶️  mpv started (pid %d)\n", cmd.Process.Pid)
	// Don't wait — let it run detached
	go cmd.Wait() //nolint:errcheck
	return nil
}
