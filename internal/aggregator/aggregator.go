package aggregator

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/slack-go/slack"
)

// threadKey extracts a key that identifies the thread a message belongs to.
// For thread replies: uses channel + thread_ts from permalink query param.
// For thread roots or non-thread messages: uses channel + message ts from permalink path.
func threadKey(m slack.SearchMessage) (string, error) {
	u, err := url.Parse(m.Permalink)
	if err != nil {
		return "", fmt.Errorf("failed to parse permalink %q: %w", m.Permalink, err)
	}

	// Normalize to "p" + digits format for consistent comparison.
	// thread_ts "1773055460.624939" → "p1773055460624939"
	// path      ".../p1773055460624939" → "p1773055460624939"
	if threadTS := u.Query().Get("thread_ts"); threadTS != "" {
		normalized := "p" + strings.ReplaceAll(threadTS, ".", "")
		return m.Channel.ID + ":" + normalized, nil
	}

	// permalink path: /archives/{channel_id}/{p<ts>}
	parts := strings.Split(u.Path, "/")
	if len(parts) >= 4 && strings.HasPrefix(parts[3], "p") {
		return m.Channel.ID + ":" + parts[3], nil
	}

	return "", fmt.Errorf("unexpected permalink format: %q", m.Permalink)
}

func Aggregate(sources ...[]slack.SearchMessage) ([]slack.SearchMessage, error) {
	best := make(map[string]slack.SearchMessage)

	for _, msgs := range sources {
		for _, m := range msgs {
			key, err := threadKey(m)
			if err != nil {
				return nil, err
			}
			if existing, ok := best[key]; ok {
				if m.Timestamp > existing.Timestamp {
					best[key] = m
				}
			} else {
				best[key] = m
			}
		}
	}

	result := make([]slack.SearchMessage, 0, len(best))
	for _, m := range best {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result, nil
}
