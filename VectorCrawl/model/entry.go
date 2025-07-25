package model

import (
	"time"
)

type Entry struct {
	URL               string `json:"url"`
	Title             string `json:"title"`
	PublishedAtParsed time.Time
	PublishedAt       int64  `json:"published_at"`
	RawContent        string `json:"raw_content"`
	FullContent       string `json:"full_content"`
	Author            string `json:"author"`
	ImageUrl          string `json:"image_url"`
	Language          string `json:"language"`
	MediaContent      string `json:"media_content"`
	DownloadFileUrl   string `json:"download_file_url"`
	DownloadFileType  string `json:"download_file_type"`
}

type EntryFetchResponseModel struct {
	Code int   `json:"code"`
	Data Entry `json:"data"`
}
