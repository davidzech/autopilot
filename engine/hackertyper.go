package engine

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/davidzech/autopilot/term"
)

func (e *Engine) runHackerTyper(shell string, r io.Reader) error {
	cmd := exec.Command(shell)
	cmd.Env = e.environ
	myPty, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer myPty.Close()
	c := make(chan os.Signal, 1)
	defer close(c)
	signal.Notify(c, syscall.SIGWINCH)
	go func() {
		for range c {
			_ = pty.InheritSize(os.Stdin, myPty)
		}
	}()
	c <- syscall.SIGWINCH

	script := bufio.NewScanner(r)

	go func() {
		for script.Scan() {
			s := script.Text()
			if err == io.EOF {
				break
			}
			if s == "" || s == "\n" || s == "\r\n" {
				continue
			}
			if strings.HasPrefix(s, "#!") {
				continue
			}
			// got a string, consume from stdin until we find a \n
			for {
				var buf [1]byte
				_, err := e.stdin.Read(buf[:])

				if err != nil {
					panic(err)
				}

				if s != "" {
					_, err = myPty.Write([]byte(s[0:1]))
					if err != nil {
						panic(err)
					}
					s = s[1:]
				} else if buf[0] == '\r' || buf[0] == '\n' {
					_, err = myPty.Write(buf[:])
					if err != nil {
						panic(err)
					}
					break
				}
			}
		}
		// done reading our script file, EOT
		_, _ = myPty.Write([]byte{term.EOT})
	}()
	_, err = io.Copy(e.stdout, myPty)
	if err != nil {
		return err
	}
	return nil
}
