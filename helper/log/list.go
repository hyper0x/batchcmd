package log

// List 代表日志列表。
type List interface {
	// Append 会追加一条日志。
	Append(one One)
	// GetAll 会返回所有的日志。
	GetAll() []One
	// Len 会返回日志列表的长度。
	Len() int
}

type list struct {
	//mu   sync.RWMutex // 暂时无并发需求。
	some []One
}

// NewList 会创建一个日志列表实例。
func NewList() List {
	return &list{
		some: make([]One, 0),
	}
}

func (l *list) Append(one One) {
	if one == nil {
		return
	}
	//l.mu.Lock()
	l.some = append(l.some, one)
	//l.mu.Unlock()
}

func (l *list) GetAll() []One {
	//l.mu.RLock()
	dst := make([]One, len(l.some))
	copy(dst, l.some)
	//l.mu.RUnlock()
	return dst
}

func (l *list) Len() int {
	//l.mu.RLock()
	length := len(l.some)
	//l.mu.RUnlock()
	return length
}
