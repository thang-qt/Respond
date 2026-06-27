package elo

import "testing"

func TestWinResultEqualRatings(t *testing.T) {
	winner, loser := WinResult(1200, 1200)
	if winner != 1216 {
		t.Fatalf("winner rating = %d, want 1216", winner)
	}
	if loser != 1184 {
		t.Fatalf("loser rating = %d, want 1184", loser)
	}
}

func TestDrawResultStrongerPlayerLosesPoints(t *testing.T) {
	high, low := DrawResult(1400, 1200)
	if high >= 1400 {
		t.Fatalf("higher-rated player should lose points on draw, got %d", high)
	}
	if low <= 1200 {
		t.Fatalf("lower-rated player should gain points on draw, got %d", low)
	}
}

func TestResignPenaltyIsNegativeAndSmallerThanNormalLoss(t *testing.T) {
	resignerRating := 1200
	opponentRating := 1200

	_, loserNew := WinResult(opponentRating, resignerRating)
	normalLoss := loserNew - resignerRating
	penalty := ResignPenalty(resignerRating, opponentRating)

	if penalty >= 0 {
		t.Fatalf("penalty must be negative, got %d", penalty)
	}
	if penalty <= normalLoss {
		t.Fatalf("resign penalty should be smaller in magnitude than normal loss, penalty=%d normalLoss=%d", penalty, normalLoss)
	}
}
