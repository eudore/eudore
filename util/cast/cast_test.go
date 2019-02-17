package cast


import (
	"testing"
	"eudore/util/cast"
)


func TestConv(t *testing.T) {
	var i interface{}
	var n int64 = 1546350714
	i = n
	t.Log(i)
	t.Log(i.(int64))
	t.Log(cast.GetDefaultInt64(i, 0))
}