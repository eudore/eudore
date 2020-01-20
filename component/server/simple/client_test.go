package simple

import (
	"github.com/eudore/eudore/component/server/simple"
	"testing"
)

func TestClient(t *testing.T) {
	t.Log(simple.NewRequest("GET", "localhost:80", "/"))
}
