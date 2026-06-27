package elo

import "math"

const K = 32

// expectedScore returns the expected score for playerA given both ratings.
func expectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}

// WinResult calculates the new ratings after a win.
// Returns (winnerNewRating, loserNewRating).
func WinResult(winnerRating, loserRating int) (int, int) {
	eWinner := expectedScore(winnerRating, loserRating)
	eLoser := 1.0 - eWinner

	newWinner := winnerRating + int(math.Round(K*(1.0-eWinner)))
	newLoser := loserRating + int(math.Round(K*(0.0-eLoser)))

	return newWinner, newLoser
}

// DrawResult calculates the new ratings after a draw.
// Returns (playerANewRating, playerBNewRating).
func DrawResult(ratingA, ratingB int) (int, int) {
	eA := expectedScore(ratingA, ratingB)
	eB := 1.0 - eA

	newA := ratingA + int(math.Round(K*(0.5-eA)))
	newB := ratingB + int(math.Round(K*(0.5-eB)))

	return newA, newB
}

// ResignPenalty calculates the rating loss for a resigning player.
// The penalty is 25% of what a normal loss would cost.
func ResignPenalty(resignerRating, opponentRating int) int {
	eResigner := expectedScore(resignerRating, opponentRating)
	normalLoss := int(math.Round(K * (0.0 - (1.0 - eResigner))))
	// normalLoss is negative; take 25% of that
	penalty := int(math.Round(float64(normalLoss) * 0.25))
	return penalty // negative number
}
