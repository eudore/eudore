package sri


import (
	"testing"
	"eudore/util/sri"
)


func TestCalculate(t *testing.T) {
	t.Log(sri.GetStatic("/data/web/static/html/sri.html"))
	t.Log(sri.NewSrier().Calculate("/data/web/static/html/sri.html"))
}