package firewall_verdict

type Verdict int

const (
	Accept Verdict = iota
	Drop
	Reject
)
