package conn

type connInvalidErr struct {
}

func (e *connInvalidErr) Error() string {
	return "conn invalid"
}
