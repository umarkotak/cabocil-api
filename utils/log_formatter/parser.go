package log_formatter

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Raw       string
	Level     string
	Fields    string
	Message   string
	File      string
	Function  string
}

var (
	ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*m")

	containerPrefixRegex = regexp.MustCompile(
		`^` +
			// 1. Container Timestamp
			`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[\+\-]\d{2}:\d{2})` +
			// Optional fractional seconds with more precision that time.RFC3339Nano
			`(?:\.\d+)?` +
			// 2. Stream/Tag (e.g., ' stderr F ' or similar)
			`\s+[^:\s]+\s+[A-Za-z]\s*` +
			// End of prefix match, looking for the space before the *actual* log line.
			`\s*`,
	)
)

func StripContainerPrefix(logLine string) string {
	return containerPrefixRegex.ReplaceAllString(logLine, "")
}

func stripColorLog(logEntry string) string {
	return ansiRegex.ReplaceAllString(logEntry, "")
}

func Parse(logLine string) (LogEntry, error) {
	logLine = stripColorLog(logLine)

	logLine = StripContainerPrefix(logLine)

	logLine = strings.TrimSpace(logLine)

	matches := strings.Split(logLine, Separator)
	if len(matches) < 6 {
		return LogEntry{}, fmt.Errorf("log line does not match expected format")
	}

	timestamp, err := time.Parse(DateTimeMilli, strings.TrimSpace(matches[0]))
	if err != nil {
		return LogEntry{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return LogEntry{
		Timestamp: timestamp,
		Raw:       logLine,
		Level:     strings.TrimSpace(matches[1]),
		Message:   strings.TrimSpace(matches[2]),
		Fields:    strings.TrimSpace(matches[3]),
		File:      strings.TrimSpace(matches[4]),
		Function:  strings.TrimSpace(matches[5]),
	}, nil
}
