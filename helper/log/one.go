package log

import "fmt"

// One 代表一条日志。
type One interface {
	// Level 会返回日志的级别。
	Level() Level
	// Content 会返回日志的内容。
	Content() string
	// String 会返回本条日志的字符串形式。
	String() string
}

type one struct {
	level   Level
	content string
}

// NewOne 会创建一个日志实例。
func NewOne(level Level, content string) One {
	return &one{
		level:   level,
		content: content,
	}
}

func (o *one) Level() Level {
	return o.level
}

func (o *one) Content() string {
	return o.content
}

func (o *one) String() string {
	return fmt.Sprintf("%s: %s",
		GetlevelStr(o.level), o.content)
}
