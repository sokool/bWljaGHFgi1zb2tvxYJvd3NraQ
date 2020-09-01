package log

import (
	"io"

	"github.com/sokool/gokit/log"
)

type Printer func(format string, args ...interface{})

type Logger interface {
	Tag(n string) Printer
	Print(format string, args ...interface{})
}

func NewLogger(w io.Writer) Logger {
	if w == nil {
		return &logger{log.Default}
	}

	return &logger{log.New(w, "", true)}

}

type logger struct{ *log.Logger }

func (l *logger) Tag(n string) Printer                     { return l.Logger.WithTag(n).Print }
func (l *logger) Print(format string, args ...interface{}) { l.Logger.Print(format, args...) }
