package storge

import (
	"database/sql"

	"bytetrade.io/web3os/system_workflow/common"
	"bytetrade.io/web3os/system_workflow/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func RemoveDiscoveryFeed(postgresClient *sql.DB) error {
	_, err := postgresClient.Exec(`TRUNCATE table discoveries`)
	if err != nil {
		common.Logger.Error("unable to remove discovery ", zap.Error(err))
		return err
	}
	return nil
}

func CreateDiscoveryFeed(postgresClient *sql.DB, discovery *model.Discovery) error {
	query := `
		INSERT INTO discoveries
			(id,feed_url, site_url,title,description,icon_type,icon_content)
		VALUES
			($1, $2, $3, $4, $5, $6,$7)
	`
	id := primitive.NewObjectID().Hex()
	_, err := postgresClient.Exec(
		query,
		id,
		discovery.FeedUrl,
		discovery.SiteUrl,
		discovery.Title,
		discovery.Description,
		discovery.IconType,
		discovery.IconContent,
	)

	if err != nil {
		common.Logger.Error("unable to create discovery ", zap.Error(err))
		return err
	}

	return nil
}
