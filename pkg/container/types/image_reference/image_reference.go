package image_reference

import (
	"fmt"
	"strings"

	"github.com/Motmedel/utils_go/pkg/schema"
)

type Reference struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string
}

func (r *Reference) ContainerImage() *schema.ContainerImage {

	var name string
	if r.Registry != "" && r.Repository != "" {
		name = r.Registry + "/" + r.Repository
	}

	var imageHash *schema.ContainerImageHash
	if r.Digest != "" {
		imageHash = &schema.ContainerImageHash{All: []string{r.Digest}}
	}

	if name == "" && imageHash == nil {
		return nil
	}

	return &schema.ContainerImage{Name: name, Tag: r.Tag, Hash: imageHash}
}

func Parse(data string) (*Reference, error) {
	reference := &Reference{}

	if atIdx := strings.Index(data, "@"); atIdx != -1 {
		reference.Digest = data[atIdx+1:]
		data = data[:atIdx]
	} else if colonIdx := strings.LastIndex(data, ":"); colonIdx != -1 {
		afterColon := data[colonIdx+1:]
		if !strings.Contains(afterColon, "/") {
			reference.Tag = afterColon
			data = data[:colonIdx]
		}
	}

	slashIdx := strings.Index(data, "/")
	if slashIdx == -1 {
		return nil, fmt.Errorf("invalid image reference: missing registry")
	}
	reference.Registry = data[:slashIdx]
	reference.Repository = data[slashIdx+1:]

	if reference.Repository == "" {
		return nil, fmt.Errorf("invalid image reference: empty repository")
	}

	return reference, nil
}
