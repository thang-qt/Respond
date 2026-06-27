package main

import "time"

type debateSeed struct {
	id              string
	topic           string
	tagSlugs        []string
	timeMode        string
	turnLimit       int
	context         string
	status          string
	sideAUserID     string
	sideBUserID     *string
	sideAAnonID     string
	sideBAnonID     *string
	sideARevealed   *bool
	sideBRevealed   *bool
	outcome         *string
	winnerSide      *string
	currentTurnSide string
	turnCount       int
	upvoteCount     int
	commentCount    int
	createdAt       time.Time
	startedAt       *time.Time
	endedAt         *time.Time
	turnDeadline    *time.Time
	turnSpacing     time.Duration
}

type userSeed struct {
	id       string
	email    string
	username string
	rating   int
}

type turnSeed struct {
	debateID    string
	turnNumber  int
	side        string
	userID      string
	anonymousID string
	content     string
}

type commentSeed struct {
	id           string
	debateID     string
	parentID     *string
	userID       string
	content      string
	isReflection bool
	createdAt    time.Time
}

type voteSeed struct {
	debateID string
	count    int
}
