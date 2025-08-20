package validatable

type Validatable interface {
	Validate() error
}
