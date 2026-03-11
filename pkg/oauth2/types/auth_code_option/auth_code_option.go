package auth_code_option

type AuthCodeOption struct {
	Key   string
	Value string
}

func New(key, value string) AuthCodeOption {
	return AuthCodeOption{Key: key, Value: value}
}
