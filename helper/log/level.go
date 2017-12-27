package log

// Level 代表日志级别。
type Level uint8

// 日志级别常量。
const (
	LEVEL_DEBUG Level = 1 << iota // 调试级别。
	LEVEL_INFO                    // 常规级别。
	LEVEL_WARN                    // 警告级别。
	LEVEL_ERROR                   // 错误级别。
	LEVEL_FATAL                   // 严重级别。
)

// GetlevelStr 用于获取日志级别的字符串形式。
func GetlevelStr(level Level) string {
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
