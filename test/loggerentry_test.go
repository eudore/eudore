package test

import (
	"testing"
	"eudore"
	"encoding/json"
)

func TestLogger(t *testing.T) {
	e := &eudore.EntryStd{
		Level: 0,
	}
	body,err := json.Marshal(e)
	t.Log(string(body), err)

	body = []byte(`{
		"name":	"logger",
		"std":	true,
		"path":	"/tmp/access.log",
		"level": "0",
		"format": "json"
	}`)
	t.Log(json.Unmarshal(body, e))
}