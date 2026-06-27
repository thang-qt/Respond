package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/elo"
)

func fetchRating(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	var rating int
	if err := tx.QueryRowContext(ctx, `SELECT rating FROM users WHERE id = $1`, userID).Scan(&rating); err != nil {
		return 0, fmt.Errorf("fetch rating for %s: %w", userID, err)
	}
	return rating, nil
}

func setRatingAndStats(ctx context.Context, tx *sql.Tx, userID uuid.UUID, newRating int, winsD, lossesD, drawsD int) error {
	const q = `
		UPDATE users
		SET rating = $2, wins = wins + $3, losses = losses + $4, draws = draws + $5
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, q, userID, newRating, winsD, lossesD, drawsD); err != nil {
		return fmt.Errorf("update rating for %s: %w", userID, err)
	}
	return nil
}

func updateRatingsWin(ctx context.Context, tx *sql.Tx, winnerID, loserID uuid.UUID) (int, int, error) {
	winnerRating, err := fetchRating(ctx, tx, winnerID)
	if err != nil {
		return 0, 0, err
	}
	loserRating, err := fetchRating(ctx, tx, loserID)
	if err != nil {
		return 0, 0, err
	}

	newWinner, newLoser := elo.WinResult(winnerRating, loserRating)

	if err := setRatingAndStats(ctx, tx, winnerID, newWinner, 1, 0, 0); err != nil {
		return 0, 0, err
	}
	if err := setRatingAndStats(ctx, tx, loserID, newLoser, 0, 1, 0); err != nil {
		return 0, 0, err
	}
	return newWinner - winnerRating, newLoser - loserRating, nil
}

func updateRatingsDraw(ctx context.Context, tx *sql.Tx, userAID, userBID uuid.UUID) (int, int, error) {
	ratingA, err := fetchRating(ctx, tx, userAID)
	if err != nil {
		return 0, 0, err
	}
	ratingB, err := fetchRating(ctx, tx, userBID)
	if err != nil {
		return 0, 0, err
	}

	newA, newB := elo.DrawResult(ratingA, ratingB)

	if err := setRatingAndStats(ctx, tx, userAID, newA, 0, 0, 1); err != nil {
		return 0, 0, err
	}
	if err := setRatingAndStats(ctx, tx, userBID, newB, 0, 0, 1); err != nil {
		return 0, 0, err
	}
	return newA - ratingA, newB - ratingB, nil
}

func updateRatingResign(ctx context.Context, tx *sql.Tx, resignerID, opponentID uuid.UUID) (int, error) {
	resignerRating, err := fetchRating(ctx, tx, resignerID)
	if err != nil {
		return 0, err
	}
	opponentRating, err := fetchRating(ctx, tx, opponentID)
	if err != nil {
		return 0, err
	}

	penalty := elo.ResignPenalty(resignerRating, opponentRating)
	newRating := resignerRating + penalty

	if err := setRatingAndStats(ctx, tx, resignerID, newRating, 0, 0, 0); err != nil {
		return 0, err
	}
	return newRating - resignerRating, nil
}

func setDebateRatingDeltas(ctx context.Context, tx *sql.Tx, debateID uuid.UUID, sideADelta, sideBDelta int) error {
	const q = `
		UPDATE debates
		SET side_a_rating_delta = $2,
			side_b_rating_delta = $3
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, q, debateID, sideADelta, sideBDelta); err != nil {
		return fmt.Errorf("set debate rating deltas: %w", err)
	}
	return nil
}
