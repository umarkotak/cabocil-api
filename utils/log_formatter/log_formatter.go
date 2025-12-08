package log_formatter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type Formatter struct{}

const (
	Separator     = " || "
	ColorReset    = "\x1b[0m"
	DateTimeMilli = "2006-01-02 15:04:05.000 -0700"
)

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.TraceLevel:
		return colorGray
	case logrus.DebugLevel, logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColorByLevel(entry.Level)
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "\x1b[%dm%s%s[%s]%s%s%s",
		levelColor,
		entry.Time.Format(DateTimeMilli),
		Separator,
		strings.ToUpper(entry.Level.String()),
		Separator,
		strings.TrimSpace(strings.ReplaceAll(entry.Message, "\n", "")),
		Separator,
	)

	f.writeFields(b, entry)

	b.WriteString(ColorReset)

	b.WriteString(Separator)

	f.writeCaller(b, entry)

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (*Formatter) writeFields(b *bytes.Buffer, entry *logrus.Entry) {
	dataLen := len(entry.Data)

	if dataLen == 0 {
		b.WriteString("field=empty")
		return
	}

	keys := make([]string, 0, dataLen)
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(b, "%s=%v ", key, entry.Data[key])
	}
}

func (*Formatter) writeCaller(b *bytes.Buffer, entry *logrus.Entry) {
	if entry.HasCaller() {
		fmt.Fprintf(
			b,
			"%s:%d%s%s",
			entry.Caller.File,
			entry.Caller.Line,
			Separator,
			entry.Caller.Function,
		)
	}
}
