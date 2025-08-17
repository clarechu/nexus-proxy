package server

import (
	"context"
	"nexus3-fsnotify/src/models"
	"nexus3-fsnotify/src/server/router"
	routerv1 "nexus3-fsnotify/src/server/router/v1"
)

type CmdbConfig struct {
	DataRoot      string               `yaml:"data_root"`
	Port          int32                `yaml:"port"`
	ProxyPort     int32                `yaml:"proxy_port"`
	NexusMetadata models.NexusMetadata `yaml:"nexus_metadata"`
}

func NewCmdb(config *CmdbConfig) (Bootstrap, error) {

	ctx, cancel := context.WithCancel(context.Background())
	routeServer := routerv1.NewRouteServer(config.NexusMetadata)
	return &CMDB{
		port:      config.Port,
		proxyPort: config.ProxyPort,
		server:    router.NewServer(routeServer),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}
