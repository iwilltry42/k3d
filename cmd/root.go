/*
Copyright © 2020 The k3d Author(s)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rancher/k3d/cmd/create"
	"github.com/rancher/k3d/cmd/delete"
	"github.com/rancher/k3d/cmd/get"
	"github.com/rancher/k3d/cmd/load"
	"github.com/rancher/k3d/cmd/start"
	"github.com/rancher/k3d/cmd/stop"

	"github.com/rancher/k3d/version"

	log "github.com/sirupsen/logrus"
)

// RootFlags describes a struct that holds flags that can be set on root level of the command
type RootFlags struct {
	debugLogging bool
	runtime      string
}

var flags = RootFlags{}

// var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k3d",
	Short: "Run k3s in Docker!",
	Long: `k3d is a wrapper CLI that helps you to easily create k3s clusters inside docker.
Nodes of a k3d cluster are docker containers running a k3s image.
All Nodes of a k3d cluster are part of the same docker network.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogging)

	// add persistent flags (present to all subcommands)
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k3d/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&flags.debugLogging, "verbose", false, "Enable verbose output (debug logging)")
	rootCmd.PersistentFlags().StringVarP(&flags.runtime, "runtime", "r", "docker", "Choose a container runtime environment [docker, containerd]")

	// add local flags

	// add subcommands
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(create.NewCmdCreate())
	rootCmd.AddCommand(delete.NewCmdDelete())
	rootCmd.AddCommand(get.NewCmdGet())
	rootCmd.AddCommand(stop.NewCmdStop())
	rootCmd.AddCommand(start.NewCmdStart())
	rootCmd.AddCommand(load.NewCmdLoad())

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print k3d version",
		Long:  "Print k3d version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("k3d version %s\n", version.GetVersion())
		},
	})
}

// initLogging initializes the logger
func initLogging() {
	if flags.debugLogging {
		log.SetLevel(log.DebugLevel)
	} else {
		switch logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL")); logLevel {
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
		case "WARN":
			log.SetLevel(log.WarnLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	}
}

// Completion
var completionFunctions = map[string]func(io.Writer) error{
	"bash":       rootCmd.GenBashCompletion,
	"zsh":        rootCmd.GenZshCompletion, // FIXME: zsh completion requires https://github.com/spf13/cobra/pull/899 due to square brackets in our help texts
	"psh":        rootCmd.GenPowerShellCompletion,
	"powershell": rootCmd.GenPowerShellCompletion,
}

// NewCmdCompletion creates a new completion command
func NewCmdCompletion() *cobra.Command {
	// create new cobra command
	cmd := &cobra.Command{
		Use:   "completion SHELL",
		Short: "Generate completion scripts for [bash, zsh, powershell | psh]",
		Long:  `Generate completion scripts for [bash, zsh, powershell | psh]`,
		Args:  cobra.ExactArgs(1), // TODO: add support for 0 args = auto detection
		Run: func(cmd *cobra.Command, args []string) {
			if f, ok := completionFunctions[args[0]]; ok {
				if err := f(os.Stdout); err != nil {
					log.Fatalf("Failed to generate completion script for shell '%s'", args[0])
				}
				return
			}
			log.Fatalf("Shell '%s' not supported for completion", args[0])
		},
	}
	return cmd
}
