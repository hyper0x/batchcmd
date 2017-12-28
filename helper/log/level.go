package log

// Level represents the log Level.
type Level uint8

// Log level constants.
const (
	LEVEL_DEBUG Level = 1 << iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_FATAL
)

// GetLevelStr is used to get the string form of the log level.
func GetLevelStr(level Level) string {
	var levelStr string
	switch level {
	case LEVEL_DEBUG:
		levelStr = "DEBUG"
	case LEVEL_INFO:
		levelStr = "INFO"
	case LEVEL_WARN:
		levelStr = "WARN"
	case LEVEL_ERROR:
		levelStr = "ERROR"
	case LEVEL_FATAL:
		levelStr = "FATAL"
	default:
		levelStr = "UNKNOWN"
	}
	return levelStr
}
