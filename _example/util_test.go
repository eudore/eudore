package eudore_test

import (
	"encoding/json"
	"testing"

	"github.com/eudore/eudore"
)

func TestUtilGet2(t *testing.T) {
	t.Log(eudore.GetBool(true))
	t.Log(eudore.GetBool("false") == false)
	t.Log(eudore.GetBool(nil) == false)
	t.Log(eudore.GetBool(222) == false)
	t.Log(eudore.GetInt(true) == 1)
	t.Log(eudore.GetInt(false) == 0)

	t.Log(eudore.GetInt(nil) == 0)
	t.Log(eudore.GetInt(int(1)) == 1)
	t.Log(eudore.GetInt(int64(2)) == 2)
	t.Log(eudore.GetInt(uint(3)) == 3)
	t.Log(eudore.GetInt(float32(4)) == 4)
	t.Log(eudore.GetInt("5") == 5)
	t.Log(eudore.GetInt("a") == 0)
	t.Log(eudore.GetInt64(nil) == 0)
	t.Log(eudore.GetInt64(int(1)) == 1)
	t.Log(eudore.GetInt64(int64(2)) == 2)
	t.Log(eudore.GetInt64(uint64(3)) == 3)
	t.Log(eudore.GetInt64(float32(4)) == 4)
	t.Log(eudore.GetInt64("5") == 5)
	t.Log(eudore.GetInt64("a") == 0)

	t.Log(eudore.GetUint(nil) == 0)
	t.Log(eudore.GetUint(int32(1)) == 1)
	t.Log(eudore.GetUint(uint32(2)) == 2)
	t.Log(eudore.GetUint(int64(3)) == 3)
	t.Log(eudore.GetUint(float32(4)) == 4)
	t.Log(eudore.GetUint("5") == 5)
	t.Log(eudore.GetUint("a") == 0)
	t.Log(eudore.GetUint64(nil) == 0)
	t.Log(eudore.GetUint64(int(1)) == 1)
	t.Log(eudore.GetUint64(uint(2)) == 2)
	t.Log(eudore.GetUint64(int64(3)) == 3)
	t.Log(eudore.GetUint64(float32(4)) == 4)
	t.Log(eudore.GetUint64("5") == 5)
	t.Log(eudore.GetUint64("a") == 0)

	t.Log(eudore.GetFloat32(nil) == 0)
	t.Log(eudore.GetFloat32(float32(1)) == 1)
	t.Log(eudore.GetFloat32(float64(2)) == 2)
	t.Log(eudore.GetFloat32(int8(3)) == 3)
	t.Log(eudore.GetFloat32(uint8(4)) == 4)
	t.Log(eudore.GetFloat32("5") == 5)
	t.Log(eudore.GetFloat32("a") == 0)
	t.Log(eudore.GetFloat64(nil) == 0)
	t.Log(eudore.GetFloat64(float32(1)) == 1)
	t.Log(eudore.GetFloat64(float64(2)) == 2)
	t.Log(eudore.GetFloat64(int16(3)) == 3)
	t.Log(eudore.GetFloat64(uint16(4)) == 4)
	t.Log(eudore.GetFloat64("5") == 5)
	t.Log(eudore.GetFloat64("a") == 0)

	t.Log(eudore.GetString(nil) == "")
	t.Log(eudore.GetString(true) == "true")
	t.Log(eudore.GetString(1) == "1")
	t.Log(eudore.GetString("2") == "2")

	t.Log(eudore.GetArrayString(nil))
	t.Log(eudore.GetArrayString(TestUtilGet2))
	t.Log(eudore.GetArrayString("1"))
	t.Log(eudore.GetArrayString([]string{"1", "2", "3"}))
	t.Log(eudore.GetArrayString([]int{1, 2, 3}))
	t.Log(eudore.GetArrayString([]uint{1, 2, 3}))
	t.Log(eudore.GetArrayString([]float32{1, 2, 3}))
	t.Log(eudore.GetArrayString([]interface{}{1, 2, 3}))
}

func TestUtilGetString2(t *testing.T) {
	t.Log(eudore.GetStringBool("true"))
	t.Log(eudore.GetStringBool("false") == false)
	t.Log(eudore.GetStringBool("222") == false)

	t.Log(eudore.GetStringInt("1") == 1)
	t.Log(eudore.GetStringInt("2") == 2)
	t.Log(eudore.GetStringInt("3") == 3)
	t.Log(eudore.GetStringInt("a") == 0)
	t.Log(eudore.GetStringInt64("1") == 1)
	t.Log(eudore.GetStringInt64("2") == 2)
	t.Log(eudore.GetStringInt64("3") == 3)
	t.Log(eudore.GetStringInt64("a") == 0)

	t.Log(eudore.GetStringUint("1") == 1)
	t.Log(eudore.GetStringUint("2") == 2)
	t.Log(eudore.GetStringUint("3") == 3)
	t.Log(eudore.GetStringUint("4") == 4)
	t.Log(eudore.GetStringUint("a") == 0)
	t.Log(eudore.GetStringUint64("1") == 1)
	t.Log(eudore.GetStringUint64("2") == 2)
	t.Log(eudore.GetStringUint64("3") == 3)
	t.Log(eudore.GetStringUint64("4") == 4)
	t.Log(eudore.GetStringUint64("a") == 0)

	t.Log(eudore.GetStringFloat32("1") == 1)
	t.Log(eudore.GetStringFloat32("2") == 2)
	t.Log(eudore.GetStringFloat32("3") == 3)
	t.Log(eudore.GetStringFloat32("a") == 0)
	t.Log(eudore.GetStringFloat64("1") == 1)
	t.Log(eudore.GetStringFloat64("2") == 2)
	t.Log(eudore.GetStringFloat64("3") == 3)
	t.Log(eudore.GetStringFloat64("a") == 0)

	t.Log(eudore.GetStringDefault("1", "") == "1")
	t.Log(eudore.GetStringDefault("", "2") == "2")
	t.Log(eudore.GetStringsDefault("3", "", "") == "3")
	t.Log(eudore.GetStringsDefault("", "4", "") == "4")
	t.Log(eudore.GetStringsDefault("", "", "5") == "5")
	t.Log(eudore.GetStringsDefault("", "", "") == "")
}

func TestUtilGetWarp2(t *testing.T) {
	app := eudore.NewApp()
	eudore.NewGetWarp(func(key string) interface{} {
		return app.Get(key)
	})
	eudore.NewGetWarpWithApp(app).GetBool("key")
	eudore.NewGetWarpWithMapString(map[string]interface{}{
		"key": true,
	}).GetBool("key")
	eudore.NewGetWarpWithConfig(app.Config).GetBool("key")
	eudore.NewGetWarpWithObject(app).GetBool("key")

	data := map[string]interface{}{
		"int":       1,
		"int8":      int8(2),
		"float":     3.4,
		"string":    "warp string",
		"bytes":     []byte("warp bytes"),
		"arrayint":  []int{1, 2, 3},
		"arraystr":  []string{"a", "b", "c"},
		"arraybyte": []byte{'a', 'b', 'c'},
	}

	warp := eudore.NewGetWarpWithObject(data)
	t.Logf("%#v", warp.GetInterface(""))

	t.Log(warp.GetInt("int"))
	t.Log(warp.GetInt64("int"))
	t.Log(warp.GetUint("int"))
	t.Log(warp.GetUint64("int"))
	t.Log(warp.GetFloat32("int"))
	t.Log(warp.GetFloat64("int"))
	t.Log(warp.GetInt("int8"))
	t.Log(warp.GetInt64("int8"))
	t.Log(warp.GetUint("int8"))
	t.Log(warp.GetUint64("int8"))
	t.Log(warp.GetFloat32("int8"))
	t.Log(warp.GetFloat64("int8"))

	t.Log(warp.GetInt("int1"))
	t.Log(warp.GetInt64("int1"))
	t.Log(warp.GetUint("int1"))
	t.Log(warp.GetUint64("int1"))
	t.Log(warp.GetFloat32("int1"))
	t.Log(warp.GetFloat64("int1"))
	t.Log(warp.GetInt("int1", 3))
	t.Log(warp.GetInt64("int1", 3))
	t.Log(warp.GetUint("int1", 3))
	t.Log(warp.GetUint64("int1", 3))
	t.Log(warp.GetFloat32("int1", 3))
	t.Log(warp.GetFloat64("int1", 3))

	t.Log(warp.GetString("int"))
	t.Log(warp.GetString("string"))
	t.Log(warp.GetString("nil"))
	t.Log(warp.GetString("bytes"))
	t.Log(warp.GetString("int", "default"))
	t.Log(warp.GetString("string", "default"))
	t.Log(warp.GetString("nil", "default"))

	t.Log(warp.GetBytes("int"))
	t.Log(warp.GetBytes("string"))
	t.Log(warp.GetBytes("nil"))
	t.Log(warp.GetBytes("bytes"))

	t.Log(warp.GetStrings("nil"))
	t.Log(warp.GetStrings("string"))
	t.Log(warp.GetStrings("arrayint"))
	t.Log(warp.GetStrings("arraystr"))
	t.Log(warp.GetStrings("arraybyte"))
}

func TestTimeDuration2(t *testing.T) {
	srv := &eudore.ServerStdConfig{}
	body := `{"readtimeout":12000000,"readheadertimeout":"6s","writetimeout":"xxx"}`
	json.Unmarshal([]byte(body), srv)
}
