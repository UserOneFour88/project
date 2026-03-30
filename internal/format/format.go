package format

import (
	"fmt"
	"sort"
	"strings"

	"pipelineapp/internal/model"
)

func Summaries(items []model.PostSummary) string {
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })

	var b strings.Builder
	for _, s := range items {
		line := fmt.Sprintf(
			"post_id=%d user_id=%d words=%d title=%q\n",
			s.ID,
			s.UserID,
			s.BodyWords,
			s.Title,
		)
		b.WriteString(line)
	}
	return b.String()
}
