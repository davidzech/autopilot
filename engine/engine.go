package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/davidzech/autopilot/term"
)

func ExecuteScript(shell string, r io.Reader) error {
	fmt.Printf("autopilot: [Engaged]\r\n")

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	tty, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		_ = term.Restore(fd, state)
		os.Exit(0)
	}()
	defer func() {
		_ = term.Restore(fd, state)
	}()

	script := bufio.NewScanner(r)

	go func() {
		for script.Scan() {
			s := script.Text() + "\n"
			if err == io.EOF {
				break
			}
			if s == "" || s == "\n" {
				continue
			}
			if strings.HasPrefix(s, "#!") {
				continue
			}
			// got a string, consume from stdin until we find a \n
			for {
				var buf [1]byte
				_, err := os.Stdin.Read(buf[:])

				if err != nil {
					panic(err)
				}

				if s != "\n" {
					_, err = tty.Write([]byte(s[0:1]))
					if err != nil {
						panic(err)
					}
					s = s[1:]
				} else if buf[0] == '\r' || buf[0] == '\n' {
					_, err = tty.Write(buf[:])
					if err != nil {
						panic(err)
					}
					break
				}
			}
		}
		// done reading our script file, EOT
		_, _ = tty.Write([]byte{term.EOT})
	}()
	_, err = io.Copy(os.Stdout, tty)
	if err != nil {
		return err
	}
	fmt.Printf("\r\nautopilot: [Disengaged]\r\n")
	return nil
}
