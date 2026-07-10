package internal

// ProcessError is
type ProcessError struct {
	OK  bool
	Err string
}

func (p *ProcessError) Error() string {
	return p.Err
}
