package model

type AlgoMetaDataResponseModel struct {
	Name string `json:"name"`
}

type AlgoProviderDataResponseModel struct {
	SyncDate int    `json:"syncDate"`
	Url      string `json:"url"`
}
type AlgoSyncProviderResponseModel struct {
	Provider      string                        `json:"provider"`
	FeedName      string                        `json:"feedName"`
	FeedProvider  AlgoProviderDataResponseModel `json:"feedProvider"`
	EntryProvider AlgoProviderDataResponseModel `json:"entryProvider"`
}
type AlgoResponseModel struct {
	UUID           string                          `json:"uuid,omitempty" `
	Namespace      string                          `json:"namespace"`
	User           string                          `json:"user"`
	ResourceStatus string                          `json:"resourceStatus"`
	ResourceType   string                          `json:"resourceType"`
	Title          string                          `json:"title"`
	Version        string                          `json:"version"`
	Metadata       AlgoMetaDataResponseModel       `json:"metadata"`
	SyncProvider   []AlgoSyncProviderResponseModel `json:"syncProvider"`
}

type RecommendServiceResponseModel struct {
	Code int                 `json:"code"`
	Data []AlgoResponseModel `json:"data"`
}
