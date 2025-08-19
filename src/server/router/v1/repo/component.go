package repo

import (
	"fmt"
	"github.com/clarechu/go-client/rest"
	"github.com/emicklei/go-restful/v3"
	"net/http"
	"net/url"
	"nexus3-fsnotify/src/models"
	"path/filepath"
)

func (n *NexusRepository) Component(request *restful.Request, response *restful.Response) {
	// 获取 /repository 后面的所有路径
	component := request.PathParameter("component")
	repository := request.PathParameter("repository")
	leaf := request.QueryParameter("leaf")
	if leaf == "true" {
		componentResponse, err := getLeafComponent(n.clientSet, component, repository)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError,
				fmt.Sprintf("{\"message\":\"%s\"}", err))
			return
		}
		name := componentResponse.Result.Data.Name
		path, err := url.JoinPath(n.nexusUrl, "repository", repository, name)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError,
				fmt.Sprintf("{\"message\":\"%s\"}", err))
			return
		}
		response.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
		err = n.clientSet.Get().RequestURI(fmt.Sprintf("/repository/%s/%s", repository, name)).
			Stream(response.ResponseWriter)
		if err != nil {
			response.Header().Del("Content-Disposition")
			response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "%s"}`, err.Error()))
			return
		}
		return
	}

}

const (
	asset     string = "readAsset"
	component string = "readComponent"
)

func getLeafComponent(clientSet *rest.RESTClient,
	componentId, repository string) (models.ExtComponentResponse, error) {

	extdirect := models.ExtComponentRequest{
		Action: "coreui_Component",
		Method: component,
		Type:   "rpc",
		Data:   []string{componentId, repository},
		TID:    1,
	}
	componentResponse := models.ExtComponentResponse{}
	err := clientSet.Post().RequestURI("/service/extdirect").
		Body(extdirect).
		Do().Into(&componentResponse)
	if err != nil {
		return componentResponse, err
	}
	return componentResponse, nil
}
