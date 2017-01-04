package json

// ErrInvalidJSON is returned from the various unmarshalling routines.
type ErrInvalidJSON struct {
	err error
}

// Error implements the error interface.
func (e *ErrInvalidJSON) Error() string {
	return e.err.Error()
}
