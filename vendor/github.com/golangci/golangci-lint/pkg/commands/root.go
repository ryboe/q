package commands

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (e *Executor) initRoot() {
	rootCmd := &cobra.Command{
		Use:   "golangci-lint",
		Short: "golangci-lint is a smart linters runner.",
		Long:  `Smart, fast linters runner. Run it in cloud for every GitHub pull request on https://golangci.com`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				logrus.Fatal(err)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			runtime.GOMAXPROCS(e.cfg.Run.Concurrency)

			log.SetFlags(0) // don't print time
			if e.cfg.Run.IsVerbose {
				logrus.SetLevel(logrus.InfoLevel)
			} else {
				logrus.SetLevel(logrus.WarnLevel)
			}

			if e.cfg.Run.CPUProfilePath != "" {
				f, err := os.Create(e.cfg.Run.CPUProfilePath)
				if err != nil {
					logrus.Fatal(err)
				}
				if err := pprof.StartCPUProfile(f); err != nil {
					logrus.Fatal(err)
				}
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if e.cfg.Run.CPUProfilePath != "" {
				pprof.StopCPUProfile()
			}
			if e.cfg.Run.MemProfilePath != "" {
				f, err := os.Create(e.cfg.Run.MemProfilePath)
				if err != nil {
					logrus.Fatal(err)
				}
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					logrus.Fatal("could not write memory profile: ", err)
				}
			}

			os.Exit(e.exitCode)
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&e.cfg.Run.IsVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&e.cfg.Run.CPUProfilePath, "cpu-profile-path", "", "Path to CPU profile output file")
	rootCmd.PersistentFlags().StringVar(&e.cfg.Run.MemProfilePath, "mem-profile-path", "", "Path to memory profile output file")
	rootCmd.PersistentFlags().IntVarP(&e.cfg.Run.Concurrency, "concurrency", "j", runtime.NumCPU(), "Concurrency")

	e.rootCmd = rootCmd
}
