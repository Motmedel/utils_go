package service

import (
	"slices"
	"testing"
)

func TestServiceImages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		service *Service
		want    []string
	}{
		{
			name:    "empty service",
			service: &Service{},
			want:    nil,
		},
		{
			name: "v2 template containers",
			service: &Service{
				Template: &Template{
					Containers: []*Containers{
						{Image: "img-a"},
						{Image: "img-b"},
					},
				},
			},
			want: []string{"img-a", "img-b"},
		},
		{
			name: "v1 knative spec.template.spec.containers",
			service: &Service{
				Spec: &Spec{
					Template: &Template{
						Spec: &TemplateSpec{
							Containers: []*Containers{
								{Image: "knative-img"},
							},
						},
					},
				},
			},
			want: []string{"knative-img"},
		},
		{
			name: "both v2 and v1 combined",
			service: &Service{
				Template: &Template{
					Containers: []*Containers{{Image: "v2-img"}},
				},
				Spec: &Spec{
					Template: &Template{
						Spec: &TemplateSpec{
							Containers: []*Containers{{Image: "v1-img"}},
						},
					},
				},
			},
			want: []string{"v2-img", "v1-img"},
		},
		{
			name: "empty image strings are skipped",
			service: &Service{
				Template: &Template{
					Containers: []*Containers{
						{Image: ""},
						{Image: "keep"},
						{Image: ""},
					},
				},
			},
			want: []string{"keep"},
		},
		{
			name: "template present but nil containers",
			service: &Service{
				Template: &Template{},
			},
			want: nil,
		},
		{
			name: "spec template with nil inner spec",
			service: &Service{
				Spec: &Spec{
					Template: &Template{},
				},
			},
			want: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.service.Images()
			if !slices.Equal(got, testCase.want) {
				t.Errorf("Images() = %#v, want %#v", got, testCase.want)
			}
		})
	}
}
