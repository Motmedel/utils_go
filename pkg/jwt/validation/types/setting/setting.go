package parsed_claims

type Setting int

const (
	SettingOptional = Setting(iota)
	SettingRequired
	SettingSkip
)
