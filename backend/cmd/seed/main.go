package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gosimple/slug"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxSeedSlugLength = 80
	seedAdminEmail    = "admin.com"
	seedAdminUsername = "admin"
	seedAdminPassword = "change-me-admin-password"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC()
	data := buildSeedData(now)

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = seedUsersData(tx, data.users); err != nil {
		log.Fatal(err)
	}
	if err = seedAdminUser(tx); err != nil {
		log.Fatal(err)
	}

	tagIDs, err := loadTagIDs(tx, data.debates)
	if err != nil {
		log.Fatal(err)
	}

	if err = seedDebates(tx, data.debates, tagIDs); err != nil {
		log.Fatal(err)
	}

	if err = seedTurnsData(tx, data.turns); err != nil {
		log.Fatal(err)
	}

	if err = seedCommentsData(tx, data.comments); err != nil {
		log.Fatal(err)
	}

	for _, seed := range data.debates {
		if err = updateDebate(tx, seed); err != nil {
			log.Fatalf("update debate %s: %v", seed.id, err)
		}
		if seed.turnSpacing > 0 {
			if err = updateTurns(tx, seed.id, seed.startedAt, seed.turnSpacing); err != nil {
				log.Fatalf("update turns %s: %v", seed.id, err)
			}
		}
		if err = updateComments(tx, seed.id, now); err != nil {
			log.Fatalf("update comments %s: %v", seed.id, err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}

	if err = updateCounts(db); err != nil {
		log.Fatal(err)
	}

	if err = applyUpvoteCounts(db, data.votes); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seed data refreshed.")
}

func seedUsersData(tx *sql.Tx, users []userSeed) error {
	const query = `
		INSERT INTO users (id, email, username, password_hash, rating)
		VALUES ($1, $2, $3, 'seeded', $4)
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email,
			username = EXCLUDED.username,
			rating = EXCLUDED.rating
	`
	for _, user := range users {
		if _, err := tx.Exec(query, user.id, user.email, user.username, user.rating); err != nil {
			return fmt.Errorf("seed user %s: %w", user.id, err)
		}
	}
	return nil
}

func seedAdminUser(tx *sql.Tx) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedAdminPassword), 12)
	if err != nil {
		return fmt.Errorf("hash seed admin password: %w", err)
	}

	var existingID string
	err = tx.QueryRow(`
		SELECT id
		FROM users
		WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($2)
		LIMIT 1
	`, seedAdminEmail, seedAdminUsername).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("lookup seed admin: %w", err)
	}

	if err == sql.ErrNoRows {
		err = tx.QueryRow(`
			INSERT INTO users (email, username, password_hash, email_verified, role, account_status)
			VALUES ($1, $2, $3, true, 'admin', 'active')
			RETURNING id
		`, seedAdminEmail, seedAdminUsername, string(passwordHash)).Scan(&existingID)
		if err != nil {
			return fmt.Errorf("insert seed admin: %w", err)
		}
	} else {
		if _, err := tx.Exec(`
			UPDATE users
			SET email = $2,
				username = $3,
				password_hash = $4,
				email_verified = true,
				role = 'admin',
				account_status = 'active',
				updated_at = now()
			WHERE id = $1
		`, existingID, seedAdminEmail, seedAdminUsername, string(passwordHash)); err != nil {
			return fmt.Errorf("update seed admin: %w", err)
		}
	}

	if _, err := tx.Exec(`
		INSERT INTO notification_settings (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, existingID); err != nil {
		return fmt.Errorf("ensure seed admin notification settings: %w", err)
	}

	return nil
}

func loadTagIDs(tx *sql.Tx, seeds []debateSeed) (map[string]string, error) {
	ids := make(map[string]string)
	for _, seed := range seeds {
		for _, tagSlug := range seed.tagSlugs {
			if _, ok := ids[tagSlug]; ok {
				continue
			}
			var id string
			if err := tx.QueryRow(`SELECT id FROM tags WHERE slug = $1`, tagSlug).Scan(&id); err != nil {
				return nil, fmt.Errorf("tag %s: %w", tagSlug, err)
			}
			ids[tagSlug] = id
		}
	}
	return ids, nil
}

func seedDebates(tx *sql.Tx, seeds []debateSeed, tagIDs map[string]string) error {
	const query = `
		INSERT INTO debates (
			id, slug, topic, time_mode, turn_limit, context, status,
			side_a_user_id, side_b_user_id, side_a_anonymous_id, side_b_anonymous_id,
			side_a_revealed, side_b_revealed, outcome, winner_side,
			turn_count, current_turn_side, turn_deadline,
			upvote_count, comment_count, created_at, started_at, ended_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22, $23
		)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			topic = EXCLUDED.topic,
			time_mode = EXCLUDED.time_mode,
			turn_limit = EXCLUDED.turn_limit,
			context = EXCLUDED.context,
			status = EXCLUDED.status,
			side_a_user_id = EXCLUDED.side_a_user_id,
			side_b_user_id = EXCLUDED.side_b_user_id,
			side_a_anonymous_id = EXCLUDED.side_a_anonymous_id,
			side_b_anonymous_id = EXCLUDED.side_b_anonymous_id,
			side_a_revealed = EXCLUDED.side_a_revealed,
			side_b_revealed = EXCLUDED.side_b_revealed,
			outcome = EXCLUDED.outcome,
			winner_side = EXCLUDED.winner_side,
			turn_count = EXCLUDED.turn_count,
			current_turn_side = EXCLUDED.current_turn_side,
			turn_deadline = EXCLUDED.turn_deadline,
			upvote_count = EXCLUDED.upvote_count,
			comment_count = EXCLUDED.comment_count,
			created_at = EXCLUDED.created_at,
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at
	`
	const deleteDebateTagsQuery = `DELETE FROM debate_tags WHERE debate_id = $1`
	const insertDebateTagQuery = `
		INSERT INTO debate_tags (debate_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT (debate_id, tag_id) DO NOTHING
	`
	slugCounts := make(map[string]int)
	for _, seed := range seeds {
		if len(seed.tagSlugs) == 0 {
			return fmt.Errorf("seed %s has no tags", seed.id)
		}
		baseSlug := slugifySeed(seed.topic)
		if baseSlug == "" {
			baseSlug = "debate"
		}
		slugCounts[baseSlug]++
		slug := baseSlug
		if slugCounts[baseSlug] > 1 {
			slug = fmt.Sprintf("%s-%d", baseSlug, slugCounts[baseSlug])
		}
		if _, err := tx.Exec(
			query,
			seed.id,
			slug,
			seed.topic,
			seed.timeMode,
			seed.turnLimit,
			seed.context,
			seed.status,
			seed.sideAUserID,
			seed.sideBUserID,
			seed.sideAAnonID,
			seed.sideBAnonID,
			seed.sideARevealed,
			seed.sideBRevealed,
			seed.outcome,
			seed.winnerSide,
			seed.turnCount,
			seed.currentTurnSide,
			seed.turnDeadline,
			seed.upvoteCount,
			seed.commentCount,
			seed.createdAt,
			seed.startedAt,
			seed.endedAt,
		); err != nil {
			return fmt.Errorf("seed debate %s: %w", seed.id, err)
		}
		if _, err := tx.Exec(deleteDebateTagsQuery, seed.id); err != nil {
			return fmt.Errorf("clear debate tags %s: %w", seed.id, err)
		}
		for _, tagSlug := range seed.tagSlugs {
			tagID, ok := tagIDs[tagSlug]
			if !ok {
				return fmt.Errorf("missing tag id for %s", tagSlug)
			}
			if _, err := tx.Exec(insertDebateTagQuery, seed.id, tagID); err != nil {
				return fmt.Errorf("seed debate tag %s/%s: %w", seed.id, tagSlug, err)
			}
		}
	}
	return nil
}

func seedTurnsData(tx *sql.Tx, turns []turnSeed) error {
	const query = `
		INSERT INTO turns (debate_id, turn_number, side, user_id, anonymous_id, content, is_system)
		VALUES ($1, $2, $3, $4, $5, $6, false)
		ON CONFLICT (debate_id, turn_number) DO UPDATE SET
			side = EXCLUDED.side,
			user_id = EXCLUDED.user_id,
			anonymous_id = EXCLUDED.anonymous_id,
			content = EXCLUDED.content
	`
	for _, turn := range turns {
		if _, err := tx.Exec(
			query,
			turn.debateID,
			turn.turnNumber,
			turn.side,
			turn.userID,
			turn.anonymousID,
			turn.content,
		); err != nil {
			return fmt.Errorf("seed turn %s-%d: %w", turn.debateID, turn.turnNumber, err)
		}
	}
	return nil
}

func seedCommentsData(tx *sql.Tx, comments []commentSeed) error {
	if len(comments) == 0 {
		return nil
	}
	const query = `
		INSERT INTO comments (id, debate_id, parent_id, user_id, content, is_reflection, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			is_reflection = EXCLUDED.is_reflection,
			created_at = EXCLUDED.created_at
	`
	for _, comment := range comments {
		if _, err := tx.Exec(
			query,
			comment.id,
			comment.debateID,
			comment.parentID,
			comment.userID,
			comment.content,
			comment.isReflection,
			comment.createdAt,
		); err != nil {
			return fmt.Errorf("seed comment %s: %w", comment.id, err)
		}
	}
	return nil
}

func updateDebate(tx *sql.Tx, seed debateSeed) error {
	const query = `
		UPDATE debates
		SET created_at = $2,
			started_at = $3,
			ended_at = $4,
			turn_deadline = $5
		WHERE id = $1
	`
	_, err := tx.Exec(query, seed.id, seed.createdAt, seed.startedAt, seed.endedAt, seed.turnDeadline)
	return err
}

func updateTurns(tx *sql.Tx, debateID string, startedAt *time.Time, spacing time.Duration) error {
	if startedAt == nil {
		return nil
	}

	const query = `
		UPDATE turns
		SET created_at = $2::timestamptz + ((turn_number - 1) * ($3 * interval '1 second'))
		WHERE debate_id = $1
	`
	seconds := int64(spacing.Seconds())
	_, err := tx.Exec(query, debateID, *startedAt, seconds)
	return err
}

func updateComments(tx *sql.Tx, debateID string, now time.Time) error {
	const query = `
		UPDATE comments
		SET created_at = $2::timestamptz + (random() * interval '12 hours')
		WHERE debate_id = $1
	`
	_, err := tx.Exec(query, debateID, now.Add(-36*time.Hour))
	return err
}

func updateCounts(db *sql.DB) error {
	const query = `
		UPDATE debates d
		SET turn_count = COALESCE(t.turn_total, 0),
			comment_count = COALESCE(c.comment_total, 0)
		FROM (
			SELECT debate_id, COUNT(*) AS turn_total
			FROM turns
			WHERE is_system = false
			GROUP BY debate_id
		) t
		FULL JOIN (
			SELECT debate_id, COUNT(*) AS comment_total
			FROM comments
			WHERE hidden = false
			GROUP BY debate_id
		) c
		ON t.debate_id = c.debate_id
		WHERE d.id = COALESCE(t.debate_id, c.debate_id)
	`
	_, err := db.Exec(query)
	return err
}

func applyUpvoteCounts(db *sql.DB, votes []voteSeed) error {
	for _, vote := range votes {
		if _, err := db.Exec(`UPDATE debates SET upvote_count = $2 WHERE id = $1`, vote.debateID, vote.count); err != nil {
			return fmt.Errorf("update upvotes %s: %w", vote.debateID, err)
		}
	}
	return nil
}

func slugifySeed(value string) string {
	slug.MaxLength = maxSeedSlugLength
	return strings.Trim(slug.Make(value), "-")
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
