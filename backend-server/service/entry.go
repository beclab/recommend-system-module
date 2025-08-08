package service

import (
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

func copyEntry(entry *model.Entry, newEntry *model.Entry) {
	if newEntry == nil {
		return
	}
	entry.FullContent = newEntry.FullContent
	entry.MediaContent = newEntry.MediaContent
	entry.DownloadFileType = newEntry.DownloadFileType
	entry.DownloadFileUrl = newEntry.DownloadFileUrl
	entry.Author = newEntry.Author
	entry.Title = newEntry.Title
	entry.PublishedAt = newEntry.PublishedAt
	entry.ImageUrl = common.GetImageUrlFromContent(entry.FullContent)

}

func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, entries model.Entries) {
	newEntries := make([]*model.Entry, 0)
	updateEntries := make([]*model.Entry, 0)
	addEntryNum := 0
	for _, entry := range entries {
		savedEntry := store.GetEntryByUrl(feed.ID, entry.URL)

		if savedEntry == nil {
			entry.BflUser = feed.BflUser
			newEntry := crawler.EntryCrawler(entry.URL, feed.FeedURL, feed.ID)
			copyEntry(entry, newEntry)
			if entry.PublishedAt == 0 {
				entry.PublishedAt = entry.PublishedAtParsed.Unix()
			}

			if entry.FullContent != "" || entry.DownloadFileType != "" {
				if entry.DownloadFileType != "" {
					entry.Attachment = true
				}
				newEntries = append(newEntries, entry)
				if len(newEntries) > 10 {
					knowledge.SaveFeedEntries(entry.BflUser, store, newEntries, feed)
					newEntries = make([]*model.Entry, 0)
				}
			} else {
				common.Logger.Info("entry full content is empty", zap.String("url", entry.URL))
			}
			addEntryNum++
		} else {
			if !common.Contains(savedEntry.Sources, "wise") {
				entry.FullContent = savedEntry.FullContent
				updateEntries = append(updateEntries, entry)
			}
		}
		if addEntryNum >= 50 {
			break
		}
	}
	knowledge.SaveFeedEntries(feed.BflUser, store, newEntries, feed)
	knowledge.UpdateFeedEntries(feed.BflUser, store, updateEntries, feed)
}
