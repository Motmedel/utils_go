package tag

import "strings"

type Tag struct {
	Name      string
	OmitEmpty bool
	OmitZero  bool
	Skip      bool
	Format    string
}

func New(tagString string) *Tag {
	trimmedTagString := strings.TrimSpace(tagString)
	if trimmedTagString == "" {
		return nil
	}

	var tag Tag

	elements := strings.Split(trimmedTagString, ",")
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 && elements[0] == "-" {
		tag.Skip = true
		return &tag
	}

	tag.Name = elements[0]

	for _, option := range elements[1:] {
		switch {
		case option == "omitempty":
			tag.OmitEmpty = true
		case option == "omitzero":
			tag.OmitZero = true
		case strings.HasPrefix(option, "format="):
			tag.Format = strings.TrimPrefix(option, "format=")
		}
	}

	return &tag
}
