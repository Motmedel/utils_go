package types

type BomFormat string

const (
	BomFormatCycloneDX BomFormat = "CycloneDX"
)

type ComponentType string

const (
	ComponentTypeLibrary     ComponentType = "library"
	ComponentTypeApplication ComponentType = "application"
	ComponentTypeFramework   ComponentType = "framework"
	ComponentTypeContainer   ComponentType = "container"
)

type Component struct {
	Type               ComponentType       `json:"type"`
	Name               string              `json:"name"`
	Version            string              `json:"version,omitempty"`
	Purl               string              `json:"purl,omitempty"`
	BomRef             string              `json:"bom-ref,omitempty"`
	Hashes             []Hash              `json:"hashes,omitzero"`
	Licenses           []LicenseChoice     `json:"licenses,omitzero"`
	ExternalReferences []ExternalReference `json:"externalReferences,omitzero"`
}

type HashAlgorithm string

const (
	HashAlgorithmSHA256 HashAlgorithm = "SHA-256"
	HashAlgorithmSHA512 HashAlgorithm = "SHA-512"
)

type Hash struct {
	Algorithm HashAlgorithm `json:"alg"`
	Content   string        `json:"content"`
}

type LicenseChoice struct {
	License *License `json:"license,omitzero"`
}

type License struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ExternalReference struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Tool struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Metadata struct {
	Tools     []Tool     `json:"tools,omitzero"`
	Component *Component `json:"component,omitzero"`
}

type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitzero"`
}

type Bom struct {
	BomFormat    BomFormat    `json:"bomFormat"`
	SpecVersion  string       `json:"specVersion"`
	Version      int          `json:"version"`
	Metadata     *Metadata    `json:"metadata,omitzero"`
	Components   []Component  `json:"components,omitzero"`
	Dependencies []Dependency `json:"dependencies,omitzero"`
}
