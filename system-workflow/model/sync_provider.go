package model

type SyncProvider struct {
	//Source        []string `json:"source"`
	FeedName      string `json:"feed_name"`
	Provider      string `json:"provider"`
	FeedUrl       string `json:"feed_url"`
	EntrySyncDate int    `json:"entry_sync_date"`
	EntryUrl      string `json:"entry_url"`
	BflUsers      []string
}
type SyncProviders []*SyncProvider
