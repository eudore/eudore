package eudore_test

import (
	"testing"

	"github.com/eudore/eudore/component/server/eudore"
)

func TestStart(t *testing.T) {
	srv := eudore.NewServerEudore()
	srv.ListenAndServe(":8084")
}
