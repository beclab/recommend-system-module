package storage

import (
	"database/sql"
	"fmt"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (s *Storage) GetEntryById(entryID string) (*model.Entry, error) {
	var entry model.Entry
	query := `SELECT id, feed, title, url,full_content,author,sources 
			  FROM entries WHERE id=$1`
	err := s.db.QueryRow(query, entryID).Scan(&entry.ID,
		&entry.FeedID,
		&entry.Title,
		&entry.URL,
		&entry.FullContent,
		&entry.Author,
		pq.Array(&entry.Sources),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		common.Logger.Error("entry get by id fail", zap.Error(err))
		return nil, fmt.Errorf("unable to fetch entry by id: %v", err)
	}

	return &entry, nil
}
func (s *Storage) GetEntryByUrl(feedID, url string) *model.Entry {
	var entry model.Entry
	query := `SELECT id, feed, title, url,full_content,author,sources 
			  FROM entries WHERE feed=$1 and url=$2`
	err := s.db.QueryRow(query, feedID, url).Scan(&entry.ID,
		&entry.FeedID,
		&entry.Title,
		&entry.URL,
		&entry.FullContent,
		&entry.Author,
		pq.Array(&entry.Sources),
	)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		common.Logger.Error("entry get by url fail", zap.Error(err))
		return nil
	}

	return &entry
}

func (s *Storage) UpdateEntryContent(entry *model.Entry) {
	_, err := s.db.Exec(`UPDATE entries SET crawler=true,published_at=$1,language=$2,author=$3,title=$4,raw_content=$5,full_content=$6 where id=$8`,
		entry.PublishedAt, entry.Language, entry.Author, entry.Title, entry.RawContent, entry.FullContent, entry.ID)
	if err != nil {
		common.Logger.Error("update entry content  fail", zap.Error(err))
	}
}
