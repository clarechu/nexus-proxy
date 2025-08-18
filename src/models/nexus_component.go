package models

import "time"

// ExtComponentRequest represents a generic Ext Direct RPC call
type ExtComponentRequest struct {
	Action string   `json:"action"`
	Method string   `json:"method"`
	Data   []string `json:"data"`
	Type   string   `json:"type"`
	TID    int      `json:"tid"` // Transaction ID
}

// LeafComponentRequest represents a generic Ext Direct RPC call
type LeafComponentRequest struct {
	Action string   `json:"action"`
	Method string   `json:"method"`
	Data   []string `json:"data"`
	Type   string   `json:"type"`
	TID    int      `json:"tid"` // Transaction ID
}

// ExtComponentResponse represents the outer Ext Direct RPC envelope
type ExtComponentResponse struct {
	TID    int             `json:"tid"`
	Action string          `json:"action"`
	Method string          `json:"method"`
	Result AssetReadResult `json:"result"`
	Type   string          `json:"type"`
}

// AssetReadResult is the result of coreui_Component.readAsset
type AssetReadResult struct {
	Success bool        `json:"success"`
	Data    AssetDetail `json:"data"`
}

// AssetDetail represents a raw asset in Nexus
type AssetDetail struct {
	ID                       string          `json:"id"`
	Name                     string          `json:"name"`   // 路径 + 文件名
	Format                   string          `json:"format"` // 如 "raw", "maven2"
	ContentType              string          `json:"contentType"`
	Size                     int64           `json:"size"` // 字节大小
	RepositoryName           string          `json:"repositoryName"`
	ContainingRepositoryName string          `json:"containingRepositoryName"`
	BlobCreated              time.Time       `json:"blobCreated"`
	BlobUpdated              time.Time       `json:"blobUpdated"`
	LastDownloaded           time.Time       `json:"lastDownloaded"`
	BlobRef                  string          `json:"blobRef"`
	ComponentID              string          `json:"componentId"`
	CreatedBy                string          `json:"createdBy"`
	CreatedByIP              string          `json:"createdByIp"`
	Attributes               AssetAttributes `json:"attributes"`
}

// AssetAttributes contains format-specific metadata
type AssetAttributes struct {
	Checksum ChecksumInfo `json:"checksum"`
	Cache    CacheInfo    `json:"cache,omitempty"`
	Content  ContentInfo  `json:"content,omitempty"`
}

// ChecksumInfo contains all hash values
type ChecksumInfo struct {
	SHA1   string `json:"sha1"`
	MD5    string `json:"md5"`
	SHA512 string `json:"sha512"`
	SHA256 string `json:"sha256"`
}

// CacheInfo contains proxy cache metadata
type CacheInfo struct {
	LastVerified time.Time `json:"last_verified"`
}

// ContentInfo contains HTTP-like content metadata
type ContentInfo struct {
	LastModified string `json:"last_modified"`
	ETag         string `json:"etag"`
}
