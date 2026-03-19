package utils

import (
	"regexp"
	"strings"
)

func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	re := regexp.MustCompile("[^a-z0-9]+")
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
