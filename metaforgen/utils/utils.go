package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)

func ToTitle(name string) string {
	return titleCaser.String(name)
}
