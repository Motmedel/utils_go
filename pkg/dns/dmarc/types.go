package dmarc

const Prefix = "v=DMARC1"

type Record struct {
	Domain string `json:"domain,omitzero"`
	Raw    string `json:"raw,omitzero"`
	P      string `json:"p,omitzero"`
	Sp     string `json:"sp,omitzero"`
	Rua    string `json:"rua,omitzero"`
	Ruf    string `json:"ruf,omitzero"`
	Adkim  string `json:"adkim,omitzero"`
	Aspf   string `json:"aspf,omitzero"`
	Ri     string `json:"ri,omitzero"`
	Fo     string `json:"fo,omitzero"`
	Rf     string `json:"rf,omitzero"`
	Pct    string `json:"pct,omitzero"`
}
