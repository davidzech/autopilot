// +build aix darwin dragonfly freebsd linux,!appengine netbsd openbsd

package term

import (
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
)

const EOT = 0x4

type State struct {
	termios unix.Termios
}

func IsTerminal(fd int) bool {
	return terminal.IsTerminal(fd)
}

func MakeRaw(fd int) (*State, error) {
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	oldState := State{termios: *termios}

	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		return nil, err
	}

	return &oldState, nil
}

func DisableEcho(fd int) error {
	s, err := GetState(fd)
	if err != nil {
		return err
	}
	s.termios.Lflag &^= unix.ECHO
	s.termios.Lflag |= unix.ICANON | unix.ISIG
	s.termios.Iflag |= unix.ICRNL
	return Restore(fd, s)
}

func (s State) Copy() State {
	return s
}

func GetState(fd int) (*State, error) {
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	return &State{termios: *termios}, nil
}

func Restore(fd int, state *State) error {
	return unix.IoctlSetTermios(int(fd), unix.TIOCSETA, &state.termios)
}
