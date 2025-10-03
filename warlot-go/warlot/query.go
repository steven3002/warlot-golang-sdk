package warlot

import (
	"context"
	"encoding/json"
	"fmt"
)

// Query maps a SELECT result set into a typed slice using JSON round-trip.
// For precise mapping, struct fields should be tagged with column names,
// for example: `json:"created_at"`.
func Query[T any](ctx context.Context, p Project, sql string, params []any, opts ...CallOption) ([]T, error) {
	res, err := p.SQL(ctx, sql, params, opts...)
	if err != nil {
		return nil, err
	}
	if len(res.Rows) == 0 {
		return []T{}, nil
	}
	out := make([]T, 0, len(res.Rows))
	for _, row := range res.Rows {
		b, _ := json.Marshal(row)
		var t T
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, fmt.Errorf("row decoding failed: %w", err)
		}
		out = append(out, t)
	}
	return out, nil
}
