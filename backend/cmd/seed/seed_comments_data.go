package main

import "time"

func buildSeedComments(finishedEnded, finishedBEnded time.Time) []commentSeed {
	return []commentSeed{
		{
			id:           "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
			debateID:     "cccccccc-cccc-cccc-cccc-cccccccccccc",
			userID:       "33333333-3333-3333-3333-333333333333",
			content:      "Balanced discussion. Side A made the stronger operational case, but Side B was right to keep returning to patient trust and the risk of overconfidence in intake automation.",
			isReflection: false,
			createdAt:    finishedEnded.Add(6 * time.Hour),
		},
		{
			id:           "ffffffff-ffff-ffff-ffff-ffffffffffff",
			debateID:     "cccccccc-cccc-cccc-cccc-cccccccccccc",
			userID:       "22222222-2222-2222-2222-222222222222",
			content:      "Post-match reflection: I should have drawn a clearer boundary between intake support and clinical decision-making earlier, because that distinction carried most of my case.",
			isReflection: true,
			createdAt:    finishedEnded.Add(3 * time.Hour),
		},
		{
			id:           "99999999-9999-9999-9999-999999999999",
			debateID:     "cccccccc-cccc-cccc-cccc-cccccccccccc",
			parentID:     stringPtr("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
			userID:       "44444444-4444-4444-4444-444444444444",
			content:      "Agreed. The best moment was when the debate stopped treating AI intake as all-or-nothing and focused on where human review has to remain non-negotiable.",
			isReflection: false,
			createdAt:    finishedEnded.Add(7 * time.Hour),
		},
		{
			id:           "abababab-abab-abab-abab-abababababab",
			debateID:     "56565656-5656-5656-5656-565656565656",
			userID:       "44444444-4444-4444-4444-444444444444",
			content:      "This felt realistic. Side B defended originality well, but Side A drew the more usable product principle by separating familiar structure from generic aesthetics.",
			isReflection: false,
			createdAt:    finishedBEnded.Add(5 * time.Hour),
		},
		{
			id:           "cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd",
			debateID:     "56565656-5656-5656-5656-565656565656",
			userID:       "11111111-1111-1111-1111-111111111111",
			content:      "Post-match reflection: the strongest move was conceding that novelty matters for brand identity while arguing that primary workflows should still feel instantly understandable.",
			isReflection: true,
			createdAt:    finishedBEnded.Add(2 * time.Hour),
		},
	}

}
