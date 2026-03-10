package aggregator

import (
	"sort"

	"github.com/slack-go/slack"
)

func Aggregate(sources ...[]slack.SearchMessage) []slack.SearchMessage {
	seen := make(map[string]struct{})
	var result []slack.SearchMessage

	for _, msgs := range sources {
		for _, m := range msgs {
			key := m.Channel.ID + ":" + m.Timestamp
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, m)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result
}
