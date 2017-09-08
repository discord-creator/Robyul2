package models

import "time"

const (
	ElasticIndex        = "robyul"
	ElasticTypeMessage  = "message"
	ElasticTypeJoin     = "join"
	ElasticTypeLeave    = "leave"
	ElasticTypeReaction = "reaction"
)

type ElasticMessage struct {
	CreatedAt     time.Time
	MessageID     string
	Content       string
	ContentLength int
	Attachments   []string
	AuthorID      string
	GuildID       string
	ChannelID     string
	Embeds        int
}

type ElasticJoin struct {
	CreatedAt time.Time
	GuildID   string
	UserID    string
}

type ElasticLeave struct {
	CreatedAt time.Time
	GuildID   string
	UserID    string
}

type ElasticReaction struct {
	CreatedAt time.Time
	UserID    string
	MessageID string
	ChannelID string
	GuildID   string
	EmojiID   string
	EmojiName string
}
