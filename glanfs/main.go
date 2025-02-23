package main

import (
	"github.com/overmighty/glan/glanfs/cmd"
	"log"
	"log/slog"
	"os"
	"runtime/pprof"
)

func main() {
	level := &slog.LevelVar{}
	level.Set(slog.LevelDebug)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))

	cpuProfile := os.Getenv("GLANFS_CPU_PROFILE")
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Starting CPU profile")
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("Failed to start CPU profile", "err", err)
		}
		defer pprof.StopCPUProfile()
	}

	cmd.Execute()

	memProfile := os.Getenv("GLANFS_MEM_PROFILE")
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Writing heap profile")
		if err := pprof.WriteHeapProfile(f); err != nil {
			slog.Error("Failed to write heap profile", "err", err)
		}
		f.Close()
		return
	}
}
