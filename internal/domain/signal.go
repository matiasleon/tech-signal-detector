package domain

import "time"

type Signal struct {
	ID             string
	RawFeedID      string
	RelevanceScore float64
	SentAt         *time.Time
	CreatedAt      time.Time
}

func (s *Signal) WasSent() bool {
	return s.SentAt != nil
}
