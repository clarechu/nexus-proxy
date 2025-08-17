package repo

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/clarechu/go-client/rest"
	"github.com/emicklei/go-restful/v3"
	"html/template"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"net/url"
	"nexus3-fsnotify/src/models"
	"path/filepath"
	"time"
)

type RepositoryInterface interface {
	RepositoryHandler(ws *restful.WebService)
	Repository(request *restful.Request, response *restful.Response)
}

func (n *NexusRepository) RepositoryHandler(ws *restful.WebService) {
	ws.Route(
		ws.GET("/repository/{repository}").To(n.Repository).
			Doc("repository").
			Operation("repository"))
	ws.Route(
		ws.GET("/repository/{repository}/{blobs:*}").To(n.Repository).
			Doc("repository blobs").
			Operation("repository blobs"))

	ws.Route(
		ws.GET("/component/{repository}/{component}").To(n.Component).
			Doc("component").
			Operation("component"))
}

type NexusRepository struct {
	clientSet        *rest.RESTClient
	nexusUrl         string
	defaultTransport *http.Transport
}

func NewNexusRepository(metadata models.NexusMetadata) RepositoryInterface {
	clientSet, err := rest.RESTClientFor(rest.NewDefaultConfig(metadata.URL, metadata.Username, metadata.Password))
	if err != nil {
		klog.Fatalf("Failed to connect to Nexus: %v", err)
	}
	defaultTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &NexusRepository{clientSet, metadata.URL, defaultTransport}
}

func (n *NexusRepository) Repository(request *restful.Request, response *restful.Response) {
	// 获取 /repository 后面的所有路径
	repository := request.PathParameter("repository")
	blobs := request.PathParameter("blobs")
	if blobs == "" {
		blobs = "/"
	}
	repositories, err := getRepositories(n.clientSet, blobs, repository)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("{\"message\":\"%s\"}", err))
		return
	}
	index := autoindex(response, repository, blobs, repositories)
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Write([]byte(index))
	//writeResponse(response, repositories)
}

func (n *NexusRepository) Component(request *restful.Request, response *restful.Response) {
	// 获取 /repository 后面的所有路径
	component := request.PathParameter("component")
	repository := request.PathParameter("repository")
	componentResponse, err := getComponent(n.clientSet, component, repository)
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
	nexusPath, err := url.Parse(path)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("{\"message\":\"%s\"}", err))
		return
	}
	request.Request.URL = nexusPath
	trip, err := n.defaultTransport.RoundTrip(request.Request)
	if err != nil {
		response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "%s"}`, err.Error()))
		return
	}
	if trip.StatusCode != 200 {
		data, err := io.ReadAll(trip.Body)
		if err != nil {
			response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "%s"}`, err.Error()))
			return
		}
		response.WriteError(trip.StatusCode, errors.New(string(data)))
		return
	}
	response.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
	_, err = io.Copy(response.ResponseWriter, trip.Body)
	if err != nil {
		response.WriteError(http.StatusBadRequest, fmt.Errorf(`{"message": "%s"}`, err.Error()))
		return
	}
}

func getComponent(clientSet *rest.RESTClient,
	componentId, repository string) (models.ExtComponentResponse, error) {

	extdirect := models.ExtComponentRequest{
		Action: "coreui_Component",
		Method: "readAsset",
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

func getRepositories(clientSet *rest.RESTClient,
	blobs, repository string) (*models.ExtDirectResponse, error) {
	klog.Info("repository blobs:", blobs)
	repositories := make([]models.Repository, 0)
	err := clientSet.Get().RequestURI("/service/rest/v1/repositories").Do().Into(&repositories)
	if err != nil {
		return nil, err
	}
	e := false
	for _, repo := range repositories {
		if repo.Name == repository {
			e = true
		}
	}
	if !e {
		return nil, fmt.Errorf("repository %s not found", repository)
	}
	extdirect := models.ExtDirectRequest{
		Action: "coreui_Browse",
		Method: "read",
		Type:   "rpc",
		Data: []models.BrowseDataItem{
			{
				RepositoryName: repository,
				Node:           blobs,
			},
		},
		TID: 1,
	}
	directResponse := models.ExtDirectResponse{}
	err = clientSet.Post().RequestURI("/service/extdirect").
		Body(extdirect).
		Do().Into(&directResponse)
	if err != nil {
		return nil, err
	}
	return &directResponse, nil
}

// HTML Template for nginx-like UI
const indexHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Index of {{.Path}}</title>
    <style>
        body { font-family: monospace; padding: 20px; }
        a { text-decoration: none; color: #000; }
        a:hover { text-decoration: underline; }
        .dir { color: #00f; }
        .size { float: right; opacity: 0.7; margin-left: 20px; }
        .time { float: right; opacity: 0.7; width: 160px; }
    </style>
</head>
<body>
    <h1>Index of {{.Path}}</h1>
    <hr>
    <pre>
{{range .Entries}}<a href="{{.URL}}" class="dir">{{.Name}}{{if .IsDir}}/{{end}}</a>
    <span class="time">{{.ModTime}}</span> <span class="size">{{.Size}}</span>
{{end}}
    </pre>
    <hr>
</body>
</html>
`

var tpl = template.Must(template.New("index").Parse(indexHTML))

func autoindex(response *restful.Response, repository, fullPath string, ext *models.ExtDirectResponse) string {
	// Flatten to entries for template
	type Entry struct {
		Name    string `yaml:"name"`
		URL     string `yaml:"url"`
		ModTime string `yaml:"modTime"`
		Size    string `yaml:"size"`
		IsDir   bool   `yaml:"isDir"`
	}
	var entries []Entry
	// Add ../ if not root
	if fullPath != "/" {
		parent := filepath.Join(fullPath, "../")
		if parent == "" || parent == "." {
			parent = "/"
		}
		entries = append(entries, Entry{
			Name:    "..",
			URL:     fmt.Sprintf("/repository/%s/%s", repository, parent),
			IsDir:   true,
			ModTime: "-",
			Size:    "-",
		})
	}

	for _, entry := range ext.Result.Data {
		if entry.Type == "folder" {
			entries = append(entries, Entry{
				Name:    entry.Text,
				URL:     fmt.Sprintf("/repository/%s/%s", repository, entry.ID),
				ModTime: "-",
				IsDir:   true,
			})

		} else {
			entries = append(entries, Entry{
				Name:    entry.Text,
				URL:     fmt.Sprintf("/component/%s/%s", repository, entry.ComponentID),
				ModTime: "-",
				IsDir:   false,
			})
		}
	}

	data := struct {
		Path    string
		Entries []Entry
	}{
		Path:    fullPath,
		Entries: entries,
	}
	var buf bytes.Buffer
	tpl.Execute(&buf, data)
	return buf.String()
}
