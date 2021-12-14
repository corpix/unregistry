package main

import (
	"runtime"

	"git.backbone/corpix/unregistry/cli"
)

func init() { runtime.GOMAXPROCS(runtime.NumCPU()) }
func main() { cli.Run() }
