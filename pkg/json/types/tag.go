package types

import "strings"

type Tag struct {
	Name         string
	OmitEmpty    bool
	OmitZero     bool
	String       bool
	Skip         bool
	OtherOptions []string
}

func New(tagString string) *Tag {
	trimmedTagString := strings.TrimSpace(tagString)
	if trimmedTagString == "" {
		return nil
	}

	var tag Tag

	elements := strings.Split(strings.TrimSpace(trimmedTagString), ",")
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 && elements[0] == "-" {
		tag.Skip = true
		return &tag
	}

	tag.Name = elements[0]

	for _, option := range elements[1:] {
		switch option {
		case "omitempty":
			tag.OmitEmpty = true
		case "omitzero":
			tag.OmitZero = true
		case "string":
			tag.String = true
		default:
			tag.OtherOptions = append(tag.OtherOptions, option)
		}
	}

	return &tag
}
