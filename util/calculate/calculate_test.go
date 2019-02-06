package calculate

import (
	"os"
	"testing"
	"strings"
	"eudore/util/calculate"
)


func TestCalculate(t *testing.T) {

	a := "1+23*2+(4 + 2 *5)"
	t.Log(a)

	a = strings.Replace(a, " ", "", -1)
	exps, err := calculate.ParseExp(a)
	if err != nil {
		os.Exit(1)
	}
	t.Log(exps)
	exps2 := calculate.Pre2stuf(exps)
	t.Log(exps2)
	t.Log(calculate.Caculate(exps2))
}