package warlot

import "context"

// Pager iterates through table rows using repeated Browse calls.
// It maintains limit/offset state and stops when no rows are returned.
type Pager struct {
	Project Project
	Table   string
	Limit   int
	Offset  int
	Done    bool
}

// Next returns the next batch of rows, or nil when iteration finishes.
func (p *Pager) Next(ctx context.Context) ([]map[string]any, error) {
	if p.Done {
		return nil, nil
	}
	resp, err := p.Project.Browse(ctx, p.Table, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}
	if len(resp.Rows) == 0 {
		p.Done = true
		return nil, nil
	}
	p.Offset += len(resp.Rows)
	return resp.Rows, nil
}
