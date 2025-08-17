package router

import (
	"github.com/emicklei/go-restful/v3"
	v1 "nexus3-fsnotify/src/server/router/v1"
	"nexus3-fsnotify/src/server/router/v1/repo"
)

type Server struct {
	RestfulCont *restful.Container
}

// NewServer initializes and configures a kubelet.Server object to handle HTTP requests.
func NewServer(router v1.RouteInterface) Server {
	server := Server{
		RestfulCont: restful.NewContainer(),
	}
	nexusInterface := router.GetNexusInterface()
	ws := new(restful.WebService)
	nexusInterface.RepositoryHandler(ws)
	DefaultHandlers(ws)
	server.RestfulCont.Add(ws)

	return server
}

// DefaultHandlers registers the default set of supported HTTP request
// patterns with the restful Container.
func DefaultHandlers(ws *restful.WebService) *restful.WebService {
	ws.Route(
		ws.GET("/healthz").To(repo.Health).
			Doc("健康检查").
			Operation("health"))
	return ws
}
