package main

/*
通过设置LoggerStd的Writer获取json日志。
LoggerStd每次Write一次是一个完整的json日志的[]byte数据，将json数据匹配put到es。
也可以实现LoggerStdData接口处理日志。
*/

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp(eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Writer: NewSyncWriterElastic("http://localhost:9200", "eudore"),
		Level:  eudore.LogDebug,
		// 时间格式必须为RFC3339Nano，Kibana才能作为索引识别
		TimeFormat: time.RFC3339Nano,
	}))

	for i := 0; i < 100; i++ {
		app.Debug("now is", time.Now().String())
	}
	app.Sync()

	app.CancelFunc()
	app.Run()
}

type syncWriterElastic struct {
	addr  string
	index string
	sync.Mutex
	Datas []byte
}

func NewSyncWriterElastic(addr, index string) eudore.LoggerWriter {
	index = fmt.Sprintf("{\"index\":{\"_index\": \"%s\",\"_type\":\"doc\"}}\n", index)
	return &syncWriterElastic{
		addr:  addr,
		index: index,
	}
}

func (w *syncWriterElastic) Sync() error {
	if len(w.Datas) < 40 {
		return nil
	}
	w.Lock()
	datas := w.Datas
	w.Datas = nil
	w.Unlock()

	req, _ := http.NewRequest("POST", w.addr+"/_bulk", bytes.NewBuffer(datas))
	req.Header.Add("Content-Type", "application/json")
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (w *syncWriterElastic) Write(p []byte) (n int, err error) {
	w.Lock()
	w.Datas = append(w.Datas, p...)
	w.Unlock()
	n = len(p)
	if len(w.Datas) > 3000 {
		err = w.Sync()
	}
	return
}
