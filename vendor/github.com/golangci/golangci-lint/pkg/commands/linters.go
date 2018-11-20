package commands

import (
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/golangci/golangci-lint/pkg/lint/linter"
)

func (e *Executor) initLinters() {
	lintersCmd := &cobra.Command{
		Use:   "linters",
		Short: "List current linters configuration",
		Run:   e.executeLinters,
	}
	e.rootCmd.AddCommand(lintersCmd)
	e.initRunConfiguration(lintersCmd)
}

func IsLinterInConfigsList(name string, linters []linter.Config) bool {
	for _, lc := range linters {
		if lc.Name() == name {
			return true
		}
	}

	return false
}

func (e *Executor) executeLinters(cmd *cobra.Command, args []string) {
	enabledLCs, err := e.EnabledLintersSet.Get()
	if err != nil {
		log.Fatalf("Can't get enabled linters: %s", err)
	}

	color.Green("Enabled by your configuration linters:\n")
	printLinterConfigs(enabledLCs)

	var disabledLCs []linter.Config
	for _, lc := range e.DBManager.GetAllSupportedLinterConfigs() {
		if !IsLinterInConfigsList(lc.Name(), enabledLCs) {
			disabledLCs = append(disabledLCs, lc)
		}
	}

	color.Red("\nDisabled by your configuration linters:\n")
	printLinterConfigs(disabledLCs)

	os.Exit(0)
}
