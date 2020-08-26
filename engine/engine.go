package engine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidzech/autopilot/term"
)

type Engine struct {
	shell         string
	script        io.Reader
	cruiseControl bool
	rawMode       bool
	stdin         *os.File
	stdout        *os.File
	environ       []string
}

type Option func(e *Engine)

func CruiseControl(cc bool) Option {
	return func(e *Engine) {
		e.cruiseControl = cc
	}
}

func Stdin(stdin *os.File) Option {
	return func(e *Engine) {
		e.stdin = stdin
	}
}

func RawMode(rm bool) Option {
	return func(e *Engine) {
		e.rawMode = rm
	}
}

func Environ(environ []string) Option {
	return func(e *Engine) {
		e.environ = environ
	}
}

func New(options ...Option) *Engine {
	var engine Engine = Engine{
		cruiseControl: false,
		rawMode:       true,
		environ:       os.Environ(),
		stdin:         os.Stdin,
		stdout:        os.Stdout,
	}
	for _, opt := range options {
		opt(&engine)
	}
	return &engine
}

func (e *Engine) Run(shell string, script io.Reader) error {
	e.printEngaged()
	defer e.printDisengaged()

	restore, err := e.prepareStdin()
	if err != nil {
		return err
	}
	defer restore()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		restore()
		os.Exit(0)
	}()

	switch e.cruiseControl {
	case true:
		return e.runCruiseControl(shell, script)

	case false:
		return e.runHackerTyper(shell, script)

	}

	return nil
}

func (e *Engine) printEngaged() {
	fmt.Fprint(e.stdout, "autopilot: [Engaged]\r\n")
}

func (e *Engine) printDisengaged() {
	fmt.Fprint(e.stdout, "\r\nautopilot: [Disengaged]\r\n")
}

func (e *Engine) prepareStdin() (restore func(), err error) {
	if !e.rawMode {
		return func() {}, nil
	}

	fd := int(e.stdin.Fd())
	if !term.IsTerminal(fd) {
		return func() {}, errors.New("not a terminal")
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}

	return func() {
		_ = term.Restore(fd, state)
	}, nil
}
