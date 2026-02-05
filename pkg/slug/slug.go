package slug

import (
	"regexp"
	"strings"

	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
)

func Create(inputString string) (slug string) {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9\- ]+`)

	lowerCase := strings.ToLower(inputString)
	noDiacritics := godiacritics.Normalize(lowerCase)
	noSimbols := nonAlphanumericRegex.ReplaceAllString(noDiacritics, "")
	oneSpace := strings.Join(strings.Fields(strings.TrimSpace(noSimbols)), " ")
	slug = strings.ReplaceAll(oneSpace, " ", "-")

	return slug
}
