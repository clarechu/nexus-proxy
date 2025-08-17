package models

// ExtDirectRequest represents a generic Ext Direct RPC call
type ExtDirectRequest struct {
	Action string           `json:"action"`
	Method string           `json:"method"`
	Data   []BrowseDataItem `json:"data"`
	Type   string           `json:"type"`
	TID    int              `json:"tid"` // Transaction ID
}

// BrowseDataItem represents the data object for coreui_Browse.read
type BrowseDataItem struct {
	RepositoryName string `json:"repositoryName"`
	Node           string `json:"node"`
}

// ExtDirectResponse represents the response of an Ext Direct RPC call
type ExtDirectResponse struct {
	TID    int          `json:"tid,omitempty"`
	Action string       `json:"action,omitempty"`
	Method string       `json:"method,omitempty"`
	Result BrowseResult `json:"result,omitempty"`
	Type   string       `json:"type,omitempty"`
}

// BrowseResult is the result of coreui_Browse.read
type BrowseResult struct {
	Success bool         `json:"success,omitempty"`
	Data    []BrowseNode `json:"data,omitempty"`
}

// BrowseNode represents a file or folder in the repository browser
type BrowseNode struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Type        string `json:"type"`                  // "asset" or "folder"
	Leaf        bool   `json:"leaf"`                  // true = file, false = directory
	ComponentID string `json:"componentId,omitempty"` // may be null
	AssetID     string `json:"assetId,omitempty"`     // may be null
	PackageURL  string `json:"packageUrl,omitempty"`  // may be null
}
