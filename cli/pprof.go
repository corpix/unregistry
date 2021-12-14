package cli

import (
	"context"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"git.backbone/corpix/unregistry/pkg/log"
)

// writeProfile writes cpu and heap profile into files.
func writeProfile(ctx context.Context, l log.Logger) error {
	cpu, err := os.Create("cpu.prof")
	if err != nil {
		return err
	}
	heap, err := os.Create("heap.prof")
	if err != nil {
		return err
	}

	err = pprof.StartCPUProfile(cpu)
	if err != nil {
		return err
	}

	go func() {
		defer cpu.Close()
		defer heap.Close()

		<-ctx.Done()

		pprof.StopCPUProfile()
		err := pprof.WriteHeapProfile(heap)
		if err != nil {
			l.Error().Err(err).Msg("failed to write heap profile")
			os.Exit(1)
			return
		}

		os.Exit(0)
	}()

	return nil
}

// writeTrace writes tracing data to file.
func writeTrace(ctx context.Context, l log.Logger) error {
	t, err := os.Create("trace.prof")
	if err != nil {
		return err
	}

	err = trace.Start(t)
	if err != nil {
		return err
	}

	go func() {
		defer t.Close()

		<-ctx.Done()

		trace.Stop()
		os.Exit(0)
	}()

	return nil
}
