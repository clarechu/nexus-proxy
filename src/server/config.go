package server

import (
	"context"
	"embed"
	"fmt"
	"nexus3-fsnotify/src/models"
	"nexus3-fsnotify/src/server/router"
	routerv1 "nexus3-fsnotify/src/server/router/v1"
	"os"
)

type CmdbConfig struct {
	StaticAssets  embed.FS
	DataRoot      string               `yaml:"data_root"`
	Port          int32                `yaml:"port"`
	ProxyPort     int32                `yaml:"proxy_port"`
	NexusMetadata models.NexusMetadata `yaml:"nexus_metadata"`
}

func NewCmdb(config *CmdbConfig) (Bootstrap, error) {
	dirEntries, err := config.StaticAssets.ReadDir("static")
	if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	for _, dirEntry := range dirEntries {
		dirName := dirEntry.Name()
		file, _ := config.StaticAssets.ReadFile(fmt.Sprintf("static/%s", dirName))
		files[dirName] = string(file)
	}
	config.NexusMetadata.Files = files
	nexusUrl := os.Getenv("NEXUS_URL")
	if nexusUrl != "" {
		config.NexusMetadata.URL = nexusUrl
	}
	nexusUsername := os.Getenv("NEXUS_USERNAME")
	if nexusUsername != "" {
		config.NexusMetadata.Username = nexusUsername
	}
	nexusPassword := os.Getenv("NEXUS_PASSWORD")
	if nexusPassword != "" {
		config.NexusMetadata.Password = nexusPassword
	}
	ctx, cancel := context.WithCancel(context.Background())
	routeServer := routerv1.NewRouteServer(config.NexusMetadata)
	return &CMDB{
		static:    files,
		port:      config.Port,
		proxyPort: config.ProxyPort,
		server:    router.NewServer(routeServer),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}
