package service

import (
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, entries model.Entries) {
	newEntries := make([]*model.Entry, 0)
	updateEntries := make([]*model.Entry, 0)
	addEntryNum := 0
	for _, entry := range entries {
		savedEntry := store.GetEntryByUrl(feed.ID, entry.URL)

		if savedEntry == nil {
			//crawler.EntryCrawler(entry, feed)
			crawler.EntryCrawler(entry, feed.FeedURL, feed.UserAgent, feed.Cookie, feed.AllowSelfSignedCertificates, feed.FetchViaProxy)
			if entry.PublishedAt == 0 {
				entry.PublishedAt = entry.PublishedAtParsed.Unix()
			}

			if entry.FullContent != "" || entry.MediaUrl != "" {
				newEntries = append(newEntries, entry)
				if len(newEntries) > 20 {
					knowledge.SaveFeedEntries(store, newEntries, feed)
					newEntries = make([]*model.Entry, 0)
				}
			} else {
				common.Logger.Info("entry full content is empty", zap.String("url", entry.URL))
			}
			addEntryNum++
		} else {
			if !contains(savedEntry.Sources, "wise") {
				entry.FullContent = savedEntry.FullContent
				updateEntries = append(updateEntries, entry)
			}
		}
		if addEntryNum >= 50 {
			break
		}
	}
	knowledge.SaveFeedEntries(store, newEntries, feed)
	knowledge.UpdateFeedEntries(store, updateEntries, feed)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
