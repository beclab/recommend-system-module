package service

import (
	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/crawler"
	"bytetrade.io/web3os/backend-server/knowledge"
	"bytetrade.io/web3os/backend-server/model"
	"bytetrade.io/web3os/backend-server/storage"
	"go.uber.org/zap"
)

func CopyEntry(entry *model.Entry, newEntry *model.Entry) {
	if newEntry == nil {
		return
	}
	if newEntry.RawContent != "" {
		entry.RawContent = newEntry.FullContent
	}
	if newEntry.FullContent != "" {
		entry.FullContent = newEntry.FullContent
	}
	if newEntry.MediaContent != "" {
		entry.MediaContent = newEntry.MediaContent
	}
	if newEntry.FileType != "" {
		entry.FileType = newEntry.FileType
	}
	if newEntry.DownloadFileName != "" {
		entry.DownloadFileName = newEntry.DownloadFileName
	}
	if newEntry.DownloadFileType != "" {
		entry.DownloadFileType = newEntry.DownloadFileType
	}
	if newEntry.DownloadFileUrl != "" {
		entry.DownloadFileUrl = newEntry.DownloadFileUrl
	}
	if newEntry.Author != "" {
		entry.Author = newEntry.Author
	}
	if newEntry.Title != "" {
		entry.Title = newEntry.Title
	}
	if newEntry.PublishedAt != 0 {
		entry.PublishedAt = newEntry.PublishedAt
	}
	if newEntry.Language != "" {
		entry.Language = newEntry.Language
	}
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
			newEntry := crawler.EntryCrawler(entry.URL, feed.BflUser, feed.ID)
			CopyEntry(entry, newEntry)
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
			if !common.Contains(savedEntry.Sources, common.FeedSource) {
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
