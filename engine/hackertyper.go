package engine

import (
	"bufio"
	"io"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/davidzech/autopilot/term"
)

func (e *Engine) runHackerTyper(shell string, r io.Reader) error {
	cmd := exec.Command(shell)
	cmd.Env = e.environ
	tty, err := pty.Start(cmd)
	if err != nil {
		return err
	}

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
	_, err = io.Copy(e.stdout, tty)
	if err != nil {
		return err
	}
	return nil
}
