package service

import (
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/storage"
)

func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, entries model.Entries) {
	newEntries := make([]*model.Entry, 0)
	updateEntries := make([]*model.Entry, 0)
	var feedSearchRSSList []model.FeedNotification
	feedNotification := model.FeedNotification{
		//FeedId:   feed.ID.Hex(),
		FeedId:   feed.ID,
		FeedName: feed.Title,
		FeedIcon: "",
	}
	feedSearchRSSList = append(feedSearchRSSList, feedNotification)

	for _, entry := range entries {
		savedEntry := store.GetEntryByUrl(feed.ID, entry.URL)

		if savedEntry == nil {
			//crawler.EntryCrawler(entry, feed)
			crawler.EntryCrawler(entry, feed.FeedURL, feed.UserAgent, feed.Cookie, feed.AllowSelfSignedCertificates, feed.FetchViaProxy)
			if entry.PublishedAt == 0 {
				entry.PublishedAt = entry.PublishedAtParsed.Unix()
			}

			if entry.FullContent != "" {
				newEntries = append(newEntries, entry)
				if len(newEntries) > 20 {
					knowledge.SaveFeedEntries(store, newEntries, feed, feedSearchRSSList)
					newEntries = make([]*model.Entry, 0)
				}
			}
		} else {
			if !contains(savedEntry.Sources, "wise") {
				entry.FullContent = savedEntry.FullContent
				updateEntries = append(updateEntries, entry)
			}
		}
	}
	knowledge.SaveFeedEntries(store, newEntries, feed, feedSearchRSSList)
	knowledge.UpdateFeedEntries(store, updateEntries, feed, feedSearchRSSList)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
