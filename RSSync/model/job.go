package model

import (
	"time"
)

// Job represents a payload sent to the processing queue.
type Job struct {
	FeedID string
}

// JobList represents a list of jobs.
type JobList []Job

type ContentJob struct {
	EntryID          string
	EntryUrl         string
	EntryTitle       string
	EntryImageUrl    string
	EntryAuthor      string
	EntryPublishedAt time.Time
	Feed             *Feed
}

type ContentJobList []ContentJob
