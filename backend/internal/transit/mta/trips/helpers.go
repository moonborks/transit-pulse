package trips

import (
	"regexp"
	"strings"
)

var suffixRegex = regexp.MustCompile(`([NS]\d+)X.*`)

func normalizeShapeID(shortTripID string) string {
	parts := strings.Split(shortTripID, "_")
	if len(parts) <= 1 {
		return ""
	}
	rawShapeID := parts[1]

	switch {
	case strings.HasPrefix(rawShapeID, "SI.N"):
		return "SI..N03R"
	case strings.HasPrefix(rawShapeID, "SI.S"):
		return "SI..S03R"
	default:
		return suffixRegex.ReplaceAllString(rawShapeID, "$1")
	}
}
