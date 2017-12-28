package log

import "fmt"

// One represents single log interface.
type One interface {
	// Level returns the Level of the log.
	Level() Level
	// Content returns the contents of the log.
	Content() string
	// String returns the string form of the log.
	String() string
}

type one struct {
	level   Level
	content string
}

// NewOne creates a One's instance.
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
		GetLevelStr(o.level), o.content)
}
