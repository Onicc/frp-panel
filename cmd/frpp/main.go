package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/Onicc/frp-panel/cmd/frpp/shared"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/fatedier/golib/crypto"
	"github.com/spf13/cobra"
)

//go:embed all:out
var fs embed.FS

func main() {
	crypto.DefaultSalt = "frp"
	logger.InitLogger()
	cobra.MousetrapHelpText = ""

	rootCmd := shared.BuildCommand(fs)
	if len(os.Args) == 1 {
		rootCmd.SetArgs([]string{"master"})
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
