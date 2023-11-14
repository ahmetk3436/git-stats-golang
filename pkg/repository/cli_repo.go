package repository

import "github.com/spf13/cobra"

type Cli struct {
	Cobra *cobra.Command
}

func NewCli(cobra *cobra.Command) Cli {
	return Cli{
		Cobra: cobra,
	}
}
