package store

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func uuidArrayArg(ids []uuid.UUID) interface{} {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	return pq.Array(values)
}

func sqlStringArrayArg(values []string) interface{} {
	return pq.Array(values)
}
