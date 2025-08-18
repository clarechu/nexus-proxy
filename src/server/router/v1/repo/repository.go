package repo

import (
	"bytes"
	"fmt"
	"github.com/clarechu/go-client/rest"
	"github.com/emicklei/go-restful/v3"
	"html/template"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"nexus3-fsnotify/src/models"
	"path/filepath"
	"time"
)

type RepositoryInterface interface {
	RepositoryHandler(ws *restful.WebService)
	Repository(request *restful.Request, response *restful.Response)
	Component(request *restful.Request, response *restful.Response)
}

func (n *NexusRepository) RepositoryHandler(ws *restful.WebService) {
	ws.Route(
		ws.GET("/repository").To(n.Repo).
			Doc("repository").
			Operation("repository"))

	ws.Route(
		ws.GET("/repository/{repository}").To(n.Repository).
			Doc("repository").
			Operation("repository"))
	ws.Route(
		ws.GET("/repository/{repository}/{blobs:*}").To(n.Repository).
			Doc("repository blobs").
			Operation("repository blobs"))
	ws.Route(
		ws.GET("/asset/{repository}/{asset}").To(n.Asset).
			Doc("asset").
			Operation("asset"))
	ws.Route(
		ws.GET("/component/{repository}/{component:*}").To(n.Component).
			Doc("component").
			Operation("component"))
}

type NexusRepository struct {
	clientSet        *rest.RESTClient
	nexusUrl         string
	defaultTransport *http.Transport
	files            map[string]string
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
	return &NexusRepository{clientSet, metadata.URL, defaultTransport, metadata.Files}
}

func (n *NexusRepository) Repo(request *restful.Request, response *restful.Response) {
	// 获取 /repository 后面的所有路径
	repositories := make([]models.Repository, 0)
	err := n.clientSet.Get().RequestURI("/service/rest/v1/repositories").Do().Into(&repositories)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("{\"message\":\"%s\"}", err))
		return
	}
	index := repositoryIndex(n.files["repository.html"], repositories)
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Write([]byte(index))
	//writeResponse(response, repositories)
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
	index := autoindex(n.files["repository.html"], repository, blobs, repositories)
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Write([]byte(index))
	//writeResponse(response, repositories)
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

func repositoryIndex(indexHtml string, ext []models.Repository) string {
	// Flatten to entries for template
	type Entry struct {
		Name    string `yaml:"name"`
		URL     string `yaml:"url"`
		ModTime string `yaml:"modTime"`
		Size    string `yaml:"size"`
		IsDir   bool   `yaml:"isDir"`
	}
	var entries []Entry

	for _, entry := range ext {
		entries = append(entries, Entry{
			Name:    entry.Name,
			URL:     fmt.Sprintf("/repository/%s/", entry.Name),
			ModTime: "-",
			IsDir:   true,
		})
	}

	data := struct {
		Path    string
		Entries []Entry
	}{
		Path:    "/",
		Entries: entries,
	}
	tpl := template.Must(template.New("index").Parse(indexHtml))
	var buf bytes.Buffer
	tpl.Execute(&buf, data)
	return buf.String()
}

func autoindex(indexHtml string, repository, fullPath string, ext *models.ExtDirectResponse) string {
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
		e := Entry{
			Name: entry.Text,
		}
		if !entry.Leaf {
			e.IsDir = true
		}
		if entry.Type == "folder" {
			e.URL = fmt.Sprintf("/repository/%s/%s", repository, entry.ID)
			entries = append(entries, e)
		} else {
			if entry.Type == "asset" {
				if entry.Leaf {
					e.URL = fmt.Sprintf("/asset/%s/%s", repository, entry.AssetID)
				} else {
					e.URL = fmt.Sprintf("/repository/%s/%s", repository, entry.ID)
				}
				entries = append(entries, e)
			} else if entry.Type == "component" {
				if entry.Leaf {
					e.URL = fmt.Sprintf("/component/%s/%s", repository, entry.ID)
				} else {
					e.URL = fmt.Sprintf("/repository/%s/%s", repository, entry.ID)
				}
				entries = append(entries, e)

			}
		}
	}

	data := struct {
		Path    string
		Entries []Entry
	}{
		Path:    fullPath,
		Entries: entries,
	}
	tpl := template.Must(template.New("index").Parse(indexHtml))
	var buf bytes.Buffer
	tpl.Execute(&buf, data)
	return buf.String()
}
