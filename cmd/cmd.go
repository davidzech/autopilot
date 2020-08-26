package cmd

import (
	"fmt"
	"os"

	"github.com/davidzech/autopilot/engine"
	"github.com/spf13/cobra"
)

var shell string

const long = `A tool that automatically types your script into a new instance of your shell via PTY. 
Pressing a key causes the next character of your script is entered into the PTY. 
Newlines are only consumed when you press [enter].`

var rootCmd = &cobra.Command{
	Use:   "autopilot [flags] script",
	Short: "A tool that automatically types your script on keypress",
	Long:  long,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		e := engine.New()
		return e.Run(shell, file)
	},
	Version: Version,
}

func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&shell, "shell", "s", os.Getenv("SHELL"), "Shell to use for execution")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
