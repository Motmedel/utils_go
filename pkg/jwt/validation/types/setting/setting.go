package setting

type Setting int

const (
	SettingOptional = Setting(iota)
	SettingRequired
	SettingSkip
)
