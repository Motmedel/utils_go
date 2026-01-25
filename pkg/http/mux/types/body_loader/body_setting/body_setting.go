package body_setting

type Setting int

const (
	Required Setting = iota
	Optional
	Forbidden
)
