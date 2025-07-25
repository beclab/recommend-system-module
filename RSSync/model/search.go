package model

type FeedNotification struct {
	FeedId   string `json:"feed_id"`
	FeedName string `json:"feed_name"`
	FeedIcon string `json:"feed_icon"`
}

type NotificationData struct {
	Name      string             `json:"name"`
	EntryId   string             `json:"entry_id"`
	Created   int64              `json:"created"`
	FeedInfos []FeedNotification `json:"feed_infos"`
	Content   string             `json:"content"`
}

type MessageDataResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

type MessageNotificationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data"`
}

type DataResponse struct {
	AccessToken string `json:"access_token"`
}

type SystemServerResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message,omitempty"`
	Data    DataResponse `json:"data"`
}
