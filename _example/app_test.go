package eudore_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestAppNew2(*testing.T) {
	app := eudore.NewApp(
		context.Background(),
		eudore.NewConfigEudore(nil),
		eudore.NewRouterStd(nil),
		eudore.NewLoggerStd(nil),
		eudore.NewServerStd(nil),
		eudore.Binder(eudore.BindDefault),
		eudore.Renderer(eudore.RenderDefault),
		eudore.DefaultValidater,
		6666,
	)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})
	app.Listen(":8088")

	app.CancelFunc()
	app.Run()
}

func TestAppListen2(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})

	eudore.Set(app.Server, "", eudore.ServerStdConfig{
		ReadTimeout:  eudore.TimeDuration(12 * time.Second),
		WriteTimeout: eudore.TimeDuration(4 * time.Second),
	})
	eudore.Set(app.Server, "readtimeout", 12*time.Second)

	app.Listen(":8088")
	app.Listen(":8088")
	app.Listen(":8089")
	app.ListenTLS(":8087", "", "")
	app.ListenTLS(":8088", "", "")
	app.Listen("localhost")
	app.ListenTLS("localhost", "", "")
	app.CancelFunc()
	app.Run()
}

func TestAppServerLogger2(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		panic(9999)
	})
	app.Listen(":8088")
	http.Get("http://127.0.0.1:8088")

	app.CancelFunc()
	app.Run()
}

func TestAppInitListener2(t *testing.T) {
	app := eudore.NewApp()
	app.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.Path)
		app.ServeHTTP(w, r)
	}))
	app.Listen(":8088")
	app.CancelFunc()
	app.Run()
}

func TestAppLogger2(*testing.T) {
	app := eudore.NewApp()
	app.Debug(0)
	app.Info(1)
	app.Warning(2)
	app.Error(3)
	app.Fatal(4)
	app.CancelFunc()
	app.Run()
}

func TestAppLoggerf2(*testing.T) {
	app := eudore.NewApp()
	app.Debugf("0")
	app.Infof("1")
	app.Warningf("2")
	app.Errorf("3")
	app.Fatalf("4")
	app.CancelFunc()
	app.Run()
}

func TestAppListenTLS2(*testing.T) {
	createKey()
	app := eudore.NewApp()

	app.ListenTLS(":8088", "testcert.pem", "testkey.pem")
	app.ListenTLS(":8088", "testkey.pem", "testcert.pem")
	app.CancelFunc()
	app.Run()

	os.Remove("testcert.pem")
	os.Remove("testkey.pem")
}

func createKey() {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)

	// 设置 SSL证书的属性用途
	certificate509 := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"eudore"},
			OrganizationalUnit: []string{"eudore"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	// 生成指定位数密匙
	pk, _ := rsa.GenerateKey(rand.Reader, 1024)

	// 生成 SSL公匙
	derBytes, _ := x509.CreateCertificate(rand.Reader, &certificate509, &certificate509, &pk.PublicKey, pk)
	certOut, _ := os.Create("testcert.pem")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// 生成 SSL私匙
	keyOut, _ := os.Create("testkey.pem")
	pem.Encode(keyOut, &pem.Block{Type: "RAS PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	keyOut.Close()
}
