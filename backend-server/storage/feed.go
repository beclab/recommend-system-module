package storage

import (
	"database/sql"
	"fmt"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/model"
	"go.uber.org/zap"
)

func (s *Storage) FeedExists(feedID string) bool {
	var result bool
	query := `SELECT true FROM feeds WHERE  id=$1`
	s.db.QueryRow(query, feedID).Scan(&result)
	return result
}

func (s *Storage) GetFeedById(feedID string) (*model.Feed, error) {
	var feed model.Feed
	query := `SELECT id, feed_url, site_url, title,etag_header,last_modified_header,checked_at,parsing_error_count,
				parsing_error_message,user_agent,cookie,username,password,ignore_http_cache,allow_self_signed_certificates,fetch_via_proxy,
				icon_type,icon_content,auto_download
			  FROM feeds WHERE id=$1`
	err := s.db.QueryRow(query, feedID).Scan(&feed.ID,
		&feed.FeedURL,
		&feed.SiteURL,
		&feed.Title,
		&feed.EtagHeader,
		&feed.LastModifiedHeader,
		&feed.CheckedAt,
		&feed.ParsingErrorCount,
		&feed.ParsingErrorMsg,
		&feed.UserAgent,
		&feed.Cookie,
		&feed.Username,
		&feed.Password,
		&feed.IgnoreHTTPCache,
		&feed.AllowSelfSignedCertificates,
		&feed.FetchViaProxy,
		&feed.IconMimeType,
		&feed.IconContent,
		&feed.AutoDownload,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		common.Logger.Error("unable to fetch", zap.Error(err))
		return nil, fmt.Errorf("unable to fetch icon by hash: %v", err)
	}

	return &feed, nil
}

func (s *Storage) FeedToUpdateList(batchSize int) (jobs model.JobList, err error) {
	errorLimit := common.GetPollingParsingErrorLimit()
	query := `
		SELECT
			id
		FROM
			feeds
		WHERE
			'{"wise"}' && sources AND 
			CASE WHEN $1 > 0 THEN parsing_error_count < $1 ELSE parsing_error_count >= 0 END
		ORDER BY checked_at ASC LIMIT $2
	`
	rows, err := s.db.Query(query, errorLimit, batchSize)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch batch of jobs: %v`, err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch job: %v`, err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *Storage) UpdateFeedError(feedID string, feed *model.Feed) {
	query := `
		UPDATE
			feeds
		SET
			parsing_error_count=$1
		WHERE
			id=$2 
	`
	_, err := s.db.Exec(query,
		feed.ParsingErrorCount,
		feed.ID,
	)

	if err != nil {
		common.Logger.Error("unable to update feed error", zap.Error(err))
	}

}

func (s *Storage) ResetFeedHeader(feedID string) {
	_, err := s.db.Exec(`UPDATE feeds SET etag_header='' where id=$1`, feedID)
	if err != nil {
		common.Logger.Error("reset  feed header error", zap.Error(err))
	}
}

func (s *Storage) UpdateFeed(feedID string, feed *model.Feed) error {
	query := `
		UPDATE
			feeds
		SET
			feed_url=$1,
			site_url=$2,
			title=$3,
			etag_header=$4,
			last_modified_header=$5,
			icon_type=$6,
			icon_content=$7,
			checked_at=$8,
			parsing_error_count=$9
		WHERE
			id=$10
	`
	_, err := s.db.Exec(query,
		feed.FeedURL,
		feed.SiteURL,
		feed.Title,
		feed.EtagHeader,
		feed.LastModifiedHeader,
		feed.IconMimeType,
		feed.IconContent,
		feed.CheckedAt,
		feed.ParsingErrorCount,
		feedID,
	)

	return err

}
