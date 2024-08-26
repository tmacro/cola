package config

type ParseError struct {
	Err  error
	Path string
}

func (e ParseError) Error() string {
	return e.Path + ": " + e.Err.Error()
}
