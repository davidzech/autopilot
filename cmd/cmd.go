package cmd

import (
	"fmt"
	"os"

	"github.com/davidzech/autopilot/engine"
	"github.com/spf13/cobra"
)

var shell string

var rootCmd = &cobra.Command{
	Use:   "autopilot [flags] script",
	Short: "A tool that automatically types your script on keypress",
	Long: `A tool that automatically types your script into a new instance of your shell via PTY. 
			When you press a key, the next character of your script is entered into the PTY. Newlines are only consumed when you press [enter].`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		return engine.ExecuteScript(shell, file)
	},
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