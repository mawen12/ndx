package pool

import "errors"

var ErrInvalidConnString = errors.New("invalid connString")

var ErrRepeatConnString = errors.New("repeat connString")
