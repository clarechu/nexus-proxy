package v1

import (
	"nexus3-fsnotify/src/models"
	"nexus3-fsnotify/src/server/router/v1/repo"
)

type RouteInterface interface {
	GetNexusInterface() repo.RepositoryInterface
}

type RouteServer struct {
	nexusMetadata models.NexusMetadata
}

func NewRouteServer(nexusMetadata models.NexusMetadata) RouteInterface {
	return &RouteServer{
		nexusMetadata: nexusMetadata,
	}
}

const (
	StatusOK = "ok"
)

func (r *RouteServer) GetNexusInterface() repo.RepositoryInterface {
	return repo.NewNexusRepository(r.nexusMetadata)
}
