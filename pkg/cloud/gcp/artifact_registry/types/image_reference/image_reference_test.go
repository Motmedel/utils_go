package image_reference

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *Reference
		wantErr bool
	}{
		{
			name: "gcp artifact registry",
			data: "europe-north2-docker.pkg.dev/project/repo",
			want: &Reference{
				Region:    "europe-north2",
				ProjectId: "project",
			},
			wantErr: false,
		},
		{
			name: "gcp artifact registry with tag",
			data: "europe-north2-docker.pkg.dev/project/repo:v1",
			want: &Reference{
				Region:    "europe-north2",
				ProjectId: "project",
			},
			wantErr: false,
		},
		{
			name: "gcp artifact registry with digest",
			data: "europe-north2-docker.pkg.dev/project/repo@sha256:digest",
			want: &Reference{
				Region:    "europe-north2",
				ProjectId: "project",
			},
			wantErr: false,
		},
		{
			name: "gcp artifact registry without region",
			data: "docker.pkg.dev/project/repo",
			want: &Reference{
				Region:    "",
				ProjectId: "project",
			},
			wantErr: false,
		},
		{
			name: "gcr io",
			data: "gcr.io/project/repo",
			want: &Reference{
				Region:    "",
				ProjectId: "project",
			},
			wantErr: false,
		},
		{
			name: "eu gcr io",
			data: "eu.gcr.io/project/repo",
			want: &Reference{
				Region:    "",
				ProjectId: "project",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Region != tt.want.Region {
				t.Errorf("Parse().Region = %v, want %v", got.Region, tt.want.Region)
			}
			if got.ProjectId != tt.want.ProjectId {
				t.Errorf("Parse().ProjectId = %v, want %v", got.ProjectId, tt.want.ProjectId)
			}
		})
	}
}
