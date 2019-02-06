package sri


import (
	"testing"
	"eudore/util/sri"
)


func TestCalculate(t *testing.T) {
	t.Log(sri.HashSHA256File("/data/web/static/js/lib/mithril.min.js"))
	t.Log(sri.Match("/data/web/static/html/m2.html"))
}