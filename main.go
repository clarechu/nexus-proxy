package main

import (
	goflags "flag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"nexus3-fsnotify/cmd"
	"os"
)

var nexusPath = goflags.String("nexus-api", "http://localhost:8081",
	"Nexus data path (required) nexus-data")

//func main() {
//	flag.Parse()
//	klog.Infof("watch blobs path: %s", nexusPath)
//}

func init() {
	//	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflags.CommandLine)
}

func main() {
	rootCmd := cmd.GetRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(-1)
	}
}
