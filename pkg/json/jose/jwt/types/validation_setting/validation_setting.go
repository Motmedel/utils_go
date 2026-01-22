package validation_setting

type Setting int

const (
	Optional = Setting(iota)
	Required
	Skip
)
