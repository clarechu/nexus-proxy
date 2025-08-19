package repo

import (
	"fmt"
	"github.com/clarechu/go-client/rest"
	"github.com/emicklei/go-restful/v3"
	"net/http"
	"net/url"
	"nexus3-fsnotify/src/models"
	"path/filepath"
	"strings"
)

func (n *NexusRepository) Asset(request *restful.Request, response *restful.Response) {
	// 获取 /repository 后面的所有路径
	asset := request.PathParameter("asset")
	repository := request.PathParameter("repository")
	assetResponse, err := getAsset(n.clientSet, asset, repository)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("{\"message\":\"%s\"}", err))
		return
	}
	name := assetResponse.Result.Data.Name
	path, err := url.JoinPath(n.nexusUrl, "repository", repository, name)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("{\"message\":\"%s\"}", err))
		return
	}
	parts := strings.Split(name, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	response.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
	encoded := strings.Join(parts, "/")
	err = n.clientSet.Get().RequestURI(fmt.Sprintf("/repository/%s%s", repository, encoded)).
		Stream(response.ResponseWriter)
	if err != nil {
		response.Header().Del("Content-Disposition")
		response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "%s"}`, err.Error()))
		return
	}
}

func getAsset(clientSet *rest.RESTClient,
	componentId, repository string) (models.ExtComponentResponse, error) {

	extdirect := models.ExtComponentRequest{
		Action: "coreui_Component",
		Method: asset,
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
