package log

// List represents the log list interface.
type List interface {
	// Append appends a log.
	Append(one One)
	// GetAll returns all logs.
	GetAll() []One
	// Len returns the length of the log list.
	Len() int
}

type list struct {
	some []One
}

// NewList creates a log list instance.
func NewList() List {
	return &list{
		some: make([]One, 0),
	}
}

func (l *list) Append(one One) {
	if one == nil {
		return
	}
	l.some = append(l.some, one)
}

func (l *list) GetAll() []One {
	dst := make([]One, len(l.some))
	copy(dst, l.some)
	return dst
}

func (l *list) Len() int {
	length := len(l.some)
	return length
}
