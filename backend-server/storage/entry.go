package storage

import (
	"database/sql"
	"fmt"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

/*func (s *Storage) UpdateEntryContent(entry *model.Entry) {
	_, err := s.db.Exec(`UPDATE entries SET crawler=true,published_at=$1,language=$2,author=$3,title=$4,raw_content=$5,full_content=$6 where id=$7`,
		entry.PublishedAt, entry.Language, entry.Author, entry.Title, entry.RawContent, entry.FullContent, entry.ID)
	if err != nil {
		common.Logger.Error("update entry content  fail", zap.Error(err))
	}
}*/

func (s *Storage) CreateEnclosure(entry *model.Entry) (string, error) {
	enclosureID := primitive.NewObjectID().Hex()

	query := `
		INSERT INTO enclosures
			(id,entry_id, content, mime_type, url, local_file_path,download_status)
		VALUES
			($1, $2, $3, $4, $5,$6,$7)
	`

	_, err := s.db.Exec(
		query,
		enclosureID,
		entry.ID,
		entry.MediaContent,
		entry.MediaType,
		entry.MediaUrl,
		"",
		"downloading",
	)

	if err != nil {
		common.Logger.Error("unable to create enclosure", zap.Error(err))
		return "", fmt.Errorf(`store: unable to create enclosure %v`, err)
	}

	return enclosureID, nil
}

func (s *Storage) GetEntryByLocalFileName(fileName string) []model.Entry {
	var entries []model.Entry
	query := `SELECT id FROM entries WHERE  '{"library"}' && sources and file_type!='article' and local_file_path=$1`

	rows, err := s.db.Query(query, fileName)
	if err != nil {
		return entries
	}
	defer rows.Close()

	for rows.Next() {
		var entry model.Entry
		if err := rows.Scan(&entry.ID); err != nil {
			return entries
		}
		entries = append(entries, entry)
	}
	return entries
}

func (s *Storage) UpdateEntryFileRemove(entryID string) {
	_, err := s.db.Exec(`UPDATE entries SET local_file_path='' where id=$1`, entryID)
	if err != nil {
		common.Logger.Error("update entry when file remove error", zap.Error(err))
	}
	_, err = s.db.Exec(`UPDATE task SET status='loss' where entry_id=$1`, entryID)
	if err != nil {
		common.Logger.Error("update entry task when file remove error", zap.Error(err))
	}
}

func (s *Storage) GetEnclosureByLocalFileName(fileName string) []string {
	var enclosuredIDs []string
	query := `SELECT id FROM enclosures WHERE  local_file_path=$1`

	rows, err := s.db.Query(query, fileName)
	if err != nil {
		return enclosuredIDs
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return enclosuredIDs
		}
		enclosuredIDs = append(enclosuredIDs, id)
	}
	return enclosuredIDs
}

func (s *Storage) UpdateEnclosureFileRemove(enclosureID string) {
	_, err := s.db.Exec(`UPDATE enclosures SET download_status='loss' where id=$1`, enclosureID)
	if err != nil {
		common.Logger.Error("update enclouse when file remove error", zap.Error(err))
	}

	_, err = s.db.Exec(`UPDATE task SET status='loss' where enclosure_id=$1`, enclosureID)
	if err != nil {
		common.Logger.Error("update enclosure task when file remove error", zap.Error(err))
	}
}
