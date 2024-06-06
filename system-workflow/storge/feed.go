package storge

import (
	"database/sql"
	"fmt"

	"bytetrade.io/web3os/system_workflow/common"

	"go.uber.org/zap"
)

func UpdateFeed(postgresClient *sql.DB, sources []string, updateFeedList map[string]map[string]interface{}) {

	for _, source := range sources {
		for _, updateFeed := range updateFeedList {
			var args []interface{}
			query := `update feeds set feed_url=feed_url `
			_, siteUrlExist := updateFeed["site_url"]
			if siteUrlExist {
				query += fmt.Sprintf(`,site_url= $%d`, len(args)+1)
				args = append(args, updateFeed["site_url"])
			}
			_, titleExist := updateFeed["title"]
			if titleExist {
				query += fmt.Sprintf(`,title=$%d`, len(args)+1)
				args = append(args, updateFeed["title"])
			}
			_, iconExist := updateFeed["icon_content"]
			if iconExist {
				query += fmt.Sprintf(`,icon_content=$%d`, len(args)+1)
				args = append(args, updateFeed["icon_content"])
			}
			_, iconTypeExist := updateFeed["icon_type"]
			if iconTypeExist {
				query += fmt.Sprintf(`,icon_type=$%d`, len(args)+1)
				args = append(args, updateFeed["icon_type"])
			}
			query += fmt.Sprintf(` where '{"`+source+`"}' && sources and feed_url=$%d`, len(args)+1)
			//args = append(args, source)
			args = append(args, updateFeed["feed_url"])

			if _, err := postgresClient.Exec(query, args...); err != nil {
				common.Logger.Error("unable to update entries ", zap.String("feed url", fmt.Sprintf("%v", updateFeed["feed_url"])), zap.Error(err))
			}
		}
	}
}
