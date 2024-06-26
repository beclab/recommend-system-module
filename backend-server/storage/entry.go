package storage

import (
	"database/sql"
	"fmt"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (s *Storage) GetEntryDocList(entryIDs []string) ([]string, error) {
	query := `SELECT id, doc_id FROM entries WHERE id=ANY($1) `
	rows, err := s.db.Query(query, entryIDs)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch feeds: %w`, err)
	}
	defer rows.Close()

	docIDList := make([]string, 0)
	for rows.Next() {
		var entry model.Entry
		err := rows.Scan(
			&entry.ID,
			&entry.DocId,
		)
		if err != nil {
			common.Logger.Error("get entry doc list err", zap.Error(err))
			return nil, fmt.Errorf(`store: unable to fetch entries row: %w`, err)
		}
		docIDList = append(docIDList, entry.DocId)
	}
	return docIDList, nil
}

func (s *Storage) GetEntryById(entryID string) (*model.Entry, error) {
	var entry model.Entry
	query := `SELECT id, feed, title, url,full_content,doc_id,author,sources 
			  FROM entries WHERE id=$1`
	err := s.db.QueryRow(query, entryID).Scan(&entry.ID,
		&entry.FeedID,
		&entry.Title,
		&entry.URL,
		&entry.FullContent,
		&entry.DocId,
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
	query := `SELECT id, feed, title, url,full_content,doc_id,author,sources 
			  FROM entries WHERE feed=$1 and url=$2`
	err := s.db.QueryRow(query, feedID, url).Scan(&entry.ID,
		&entry.FeedID,
		&entry.Title,
		&entry.URL,
		&entry.FullContent,
		&entry.DocId,
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
	_, err := s.db.Exec(`UPDATE entries SET crawler=true,published_at=$1,language=$2,author=$3,title=$4,raw_content=$5,full_content=$6,doc_id=$7 where id=$8`,
		entry.PublishedAt, entry.Language, entry.Author, entry.Title, entry.RawContent, entry.FullContent, entry.DocId, entry.ID)
	if err != nil {
		common.Logger.Error("update entry content  fail", zap.Error(err))
	}
}

func (s *Storage) UpdateEntryDocID(entry *model.Entry) {
	_, err := s.db.Exec(`UPDATE entries SET doc_id=$1 where id=$2`, entry.DocId, entry.ID)
	if err != nil {
		common.Logger.Error("update entry doc  fail", zap.Error(err))
	}
}
