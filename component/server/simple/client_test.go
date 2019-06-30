package simple

import (
	"eudore/component/server/simple"
	"testing"
)

func TestClient(t *testing.T) {
	t.Log(simple.NewRequest("GET", "localhost:80", "/"))
}
