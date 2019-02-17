package cast


type (
	Map struct {
		m map[string]interface{}
	}
)

func NewMap(i interface{}) *Map {
	v, ok := i.(map[string]interface{})
	if ok {
		return &Map{m: v}
	}
	return nil
}

func (m *Map) Get(key string) interface{} {
	if len(key) == 0 {
		return m.m
	}
	v, ok := m.m[key]
	if ok {
		return v
	}
	return nil
}

func (m *Map) Set(key string, val interface{}) {
	if len(key) == 0 {
		v, ok := val.(map[string]interface{})
		if ok {
			m.m = v
		}
	}else {
		m.m[key] = val
	}
}

func (m *Map) Del(key string) {
	delete(m.m, key)
}

func (m *Map) GetInt(key string) int {
	return GetInt(m.Get(key))
}

func (m *Map) GetDefultInt(key string,n int) int {
	return GetDefultInt(m.Get(key), n)
}

func (m *Map) GetString(key string) string {
	return GetString(m.Get(key))
}

func (m *Map) GetDefaultString(key string,str string) string {
	return GetDefaultString(m.Get(key), str)
}
