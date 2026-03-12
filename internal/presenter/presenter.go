package presenter

import (
	"regexp"
)

var mentionRe = regexp.MustCompile(`<@(U[A-Z0-9]+)>`)

func ResolveUser(id string, userMap map[string]string) string {
	if name, ok := userMap[id]; ok {
		return name
	}
	return id
}

func ResolveText(text string, userMap map[string]string) string {
	return mentionRe.ReplaceAllStringFunc(text, func(match string) string {
		id := mentionRe.FindStringSubmatch(match)[1]
		if name, ok := userMap[id]; ok {
			return "@" + name
		}
		return match
	})
}
