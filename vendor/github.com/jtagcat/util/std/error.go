package std

import (
	"context"
	"errors"
	"log/slog"
)

func SlogErr(err error) slog.Attr {
	return SlogNamedErr("downstreamError", err)
}

func SlogNamedErr(key string, err error) slog.Attr {
	str := ""
	if err != nil {
		str = err.Error()
	}

	return slog.String(key, str)
}

func SlogWrap(level slog.Level, msg string, args ...any) SlogError {
	return SlogError{
		highestLevel: level,
		combinedMsg:  msg,
		args:         args,
	}
}

type SlogError struct {
	highestLevel slog.Level
	combinedMsg  string
	args         []any
}

func (e SlogError) Error() string {
	return e.combinedMsg
}

func (e SlogError) Wrap(level slog.Level, msg string, args ...any) SlogError {
	if level > e.highestLevel {
		e.highestLevel = level
	}

	if e.combinedMsg == "" {
		e.combinedMsg = msg
	} else {
		e.combinedMsg = msg + ": " + e.combinedMsg
	}

	e.args = append(e.args, args...)

	return e
}

func (e SlogError) Log(logger slog.Logger) {
	logger.Log(context.Background(), e.highestLevel, e.combinedMsg, e.args...)
}

func (e SlogError) LogD() {
	slog.Default().Log(context.Background(), e.highestLevel, e.combinedMsg, e.args...)
}

// for errors.Is(err, ERr)
type GenericErr struct {
	Err     error
	Wrapped error
}

func (a GenericErr) Is(target error) bool {
	return errors.Is(a.Err, target)
}

func (a GenericErr) Unwrap() error {
	return a.Wrapped
}

func (a GenericErr) Error() string {
	return a.Err.Error() + ": " + a.Wrapped.Error()
}
