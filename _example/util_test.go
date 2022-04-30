package eudore_test

import (
	"testing"

	"github.com/eudore/eudore"
)

func TestUtilContextKey(*testing.T) {
	app := eudore.NewApp()
	app.Info(eudore.NewContextKey("debug-key"))

	app.CancelFunc()
	app.Run()
}

func TestUtilTimeDuration(*testing.T) {
	type Data struct {
		Time eudore.TimeDuration `json:"time"`
	}

	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.AnyFunc("/time/*", func(ctx eudore.Context) interface{} {
		return eudore.TimeDuration(12000000000)
	})
	app.AnyFunc("/time/bind", func(ctx eudore.Context) error {
		var data Data
		ctx.Debug(string(ctx.Body()))
		err := ctx.Bind(&data)
		ctx.Info(err, data)
		return err
	})

	client.NewRequest("GET", "/time/text").AddHeader(eudore.HeaderAccept, eudore.MimeText).Do()
	client.NewRequest("GET", "/time/json").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("PUT", "/time/bind").AddHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON).BodyString(`{"time":"12s"}`).Do()
	client.NewRequest("PUT", "/time/bind").AddHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON).BodyString(`{"time":12000000000}`).Do()
	client.NewRequest("PUT", "/time/bind").AddHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON).BodyString(`{"time":"x"}`).Do()

	app.CancelFunc()
	app.Run()
}

func TestUtilGetCast(t *testing.T) {
	app := eudore.NewApp()

	app.Debug(eudore.GetBool(int(1)))
	app.Debug(eudore.GetBool(uint(1)))
	app.Debug(eudore.GetBool(float32(1.0)))
	app.Debug(eudore.GetBool("true"))
	app.Debug(eudore.GetInt(int(123)))
	app.Debug(eudore.GetInt(uint(234)))
	app.Debug(eudore.GetInt(float64(345)))
	app.Debug(eudore.GetInt("456"))
	app.Debug(eudore.GetInt64(int(123)))
	app.Debug(eudore.GetInt64(uint(234)))
	app.Debug(eudore.GetInt64(float64(345)))
	app.Debug(eudore.GetInt64("456"))
	app.Debug(eudore.GetUint(int(123)))
	app.Debug(eudore.GetUint(uint(234)))
	app.Debug(eudore.GetUint(float64(345)))
	app.Debug(eudore.GetUint("456"))
	app.Debug(eudore.GetUint64(int(123)))
	app.Debug(eudore.GetUint64(uint(234)))
	app.Debug(eudore.GetUint64(float64(345)))
	app.Debug(eudore.GetUint64("456"))
	app.Debug(eudore.GetFloat32(int(123)))
	app.Debug(eudore.GetFloat32(uint(234)))
	app.Debug(eudore.GetFloat32(float64(345)))
	app.Debug(eudore.GetFloat32("456"))
	app.Debug(eudore.GetFloat64(int(123)))
	app.Debug(eudore.GetFloat64(uint(234)))
	app.Debug(eudore.GetFloat64(float64(345)))
	app.Debug(eudore.GetFloat64("456"))
	app.Debug(eudore.GetString(int(123)))
	app.Debug(eudore.GetString(uint(234)))
	app.Debug(eudore.GetString(float64(345)))
	app.Debug(eudore.GetString("456"))
	app.Debug(eudore.GetString([]byte("456")))
	app.Debug(eudore.GetString(true))
	app.Debug(eudore.GetString(eudore.NewContextKey("string")))
	app.Debug(eudore.GetBytes("strings"))
	app.Debug(eudore.GetStrings("strings"))
	app.Debug(eudore.GetStrings([]interface{}{"1", "2", "3"}))

	app.CancelFunc()
	app.Run()

}

func TestUtilGetCastString(t *testing.T) {
	app := eudore.NewApp()

	app.Debug(eudore.GetStringBool("true"))
	app.Debug(eudore.GetStringBool("1"))
	app.Debug(eudore.GetStringBool("bool"))
	app.Debug(eudore.GetStringInt("1"))
	app.Debug(eudore.GetStringInt("0", 1))
	app.Debug(eudore.GetStringInt("0", 0))
	app.Debug(eudore.GetStringInt64("1"))
	app.Debug(eudore.GetStringInt64("0", 1))
	app.Debug(eudore.GetStringInt64("0", 0))
	app.Debug(eudore.GetStringUint("1"))
	app.Debug(eudore.GetStringUint("0", 1))
	app.Debug(eudore.GetStringUint("0", 0))
	app.Debug(eudore.GetStringUint64("1"))
	app.Debug(eudore.GetStringUint64("0", 1))
	app.Debug(eudore.GetStringUint64("0", 0))
	app.Debug(eudore.GetStringFloat32("1"))
	app.Debug(eudore.GetStringFloat32("0", 1))
	app.Debug(eudore.GetStringFloat32("0", 0))
	app.Debug(eudore.GetStringFloat64("1"))
	app.Debug(eudore.GetStringFloat64("0", 1))
	app.Debug(eudore.GetStringFloat64("0", 0))

	app.CancelFunc()
	app.Run()
}

func TestUtilGetWarp(t *testing.T) {
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
	app.Info("%#v", warp.GetInterface(""))

	app.Info(warp.GetInt("int"))
	app.Info(warp.GetInt64("int"))
	app.Info(warp.GetUint("int"))
	app.Info(warp.GetUint64("int"))
	app.Info(warp.GetFloat32("int"))
	app.Info(warp.GetFloat64("int"))
	app.Info(warp.GetInt("int8"))
	app.Info(warp.GetInt64("int8"))
	app.Info(warp.GetUint("int8"))
	app.Info(warp.GetUint64("int8"))
	app.Info(warp.GetFloat32("int8"))
	app.Info(warp.GetFloat64("int8"))

	app.Info(warp.GetInt("int1"))
	app.Info(warp.GetInt64("int1"))
	app.Info(warp.GetUint("int1"))
	app.Info(warp.GetUint64("int1"))
	app.Info(warp.GetFloat32("int1"))
	app.Info(warp.GetFloat64("int1"))
	app.Info(warp.GetInt("int1", 3))
	app.Info(warp.GetInt64("int1", 3))
	app.Info(warp.GetUint("int1", 3))
	app.Info(warp.GetUint64("int1", 3))
	app.Info(warp.GetFloat32("int1", 3))
	app.Info(warp.GetFloat64("int1", 3))

	app.Info(warp.GetString("int"))
	app.Info(warp.GetString("string"))
	app.Info(warp.GetString("nil"))
	app.Info(warp.GetString("bytes"))
	app.Info(warp.GetString("int", "default"))
	app.Info(warp.GetString("string", "default"))
	app.Info(warp.GetString("nil", "default"))

	app.Info(warp.GetBytes("int"))
	app.Info(warp.GetBytes("string"))
	app.Info(warp.GetBytes("nil"))
	app.Info(warp.GetBytes("bytes"))

	app.Info(warp.GetStrings("nil"))
	app.Info(warp.GetStrings("string"))
	app.Info(warp.GetStrings("arrayint"))
	app.Info(warp.GetStrings("arraystr"))
	app.Info(warp.GetStrings("arraybyte"))
}
