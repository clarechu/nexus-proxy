package main

import (
	"embed"
	goflags "flag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"nexus3-fsnotify/cmd"
	"os"
)

//go:embed static/**
var staticAssets embed.FS

func init() {
	//	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflags.CommandLine)
}

func main() {
	rootCmd := cmd.GetRootCmd(staticAssets, os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(-1)
	}
}
