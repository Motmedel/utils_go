package service

type Containers struct {
	Image string `json:"image,omitempty"`
}

type TemplateSpec struct {
	Containers []*Containers `json:"containers"`
}

type Template struct {
	Containers []*Containers `json:"containers,omitempty"`
	Spec       *TemplateSpec `json:"spec,omitempty"`
}

type Spec struct {
	Template *Template `json:"template,omitempty"`
}

type Service struct {
	// Cloud Run v2: template.containers[].image
	Template *Template `json:"template,omitempty"`
	// Cloud Run v1 (Knative): spec.template.spec.containers[].image
	Spec *Spec `json:"spec,omitempty"`
}

func (s *Service) Images() []string {
	var images []string

	// Cloud Run v2: template.containers[].image
	if s.Template != nil {
		for _, c := range s.Template.Containers {
			if c.Image != "" {
				images = append(images, c.Image)
			}
		}
	}

	// Cloud Run v1 (Knative): spec.template.spec.containers[].image
	if s.Spec != nil && s.Spec.Template != nil && s.Spec.Template.Spec != nil {
		for _, c := range s.Spec.Template.Spec.Containers {
			if c.Image != "" {
				images = append(images, c.Image)
			}
		}
	}

	return images
}
