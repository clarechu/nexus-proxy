package models

type NexusMetadata struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type Repository struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Format string `json:"format"`
}
