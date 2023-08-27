package lib

import (
	"fmt"
	"regexp"
)

var rangeRegex = regexp.MustCompile(`(?m)^[0-2][0-3]:[0-5][0-9]:[0-5][0-9](?:,(40|[0-3][0-9]|[12][0-9]|[1-9]))?(?:,(40|[0-3][0-9]|[12][0-9]|[1-9])){0,2}$`)

func IsRangeValid(s string) bool {
	return rangeRegex.MatchString(s)
}

var videosRegex = []*regexp.Regexp{}

func SetupVideoSitesRegex(patterns ...string) error {
	for _, p := range patterns {
		// Escape special characters in the domain to use it in a regex pattern.
		escapedDomain := regexp.QuoteMeta(p)

		// Construct a regex pattern to match URLs from the domain.
		pattern := fmt.Sprintf(`^https?://%s.*$`, escapedDomain)
		videosRegex = append(videosRegex, regexp.MustCompile(pattern))
	}
	return nil
}

func IsVideoURLValid(s string) bool {
	for _, r := range videosRegex {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}
