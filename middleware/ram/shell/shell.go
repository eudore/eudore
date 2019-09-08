package shell

import (
	"bytes"
	"os/exec"
	"regexp"
	"time"

	"github.com/eudore/eudore"
)

// Shell 定义shell远程ip认证
type Shell struct {
	List []string
}

// NewShell 函数处理一个是shell ram处理者，每10秒会自动更新ip数据。
func NewShell() *Shell {
	s := &Shell{}
	go func() {
		s.Update()
		time.Sleep(10 * time.Second)
	}()

	return s
}

// RamHandle 实现ram.RamHandler接口。
func (s *Shell) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRAM, "shell")
	ip := ctx.RealIP()
	for _, i := range s.List {
		if ip == i {
			return true, true
		}
	}
	return false, false
}

// Update 方法执行linux who命令，获得登录的远程ip。
func (s *Shell) Update() {
	var ips []string
	body, _ := exec.Command("/usr/bin/who").Output()
	rep, _ := regexp.Compile(`.*\(([\d.]+)\)`)
	for _, line := range bytes.Split(body, []byte("\n")) {
		params := rep.FindSubmatch(line)
		if len(params) > 1 {
			ip := string(params[1])
			ips = append(ips, ip)
		}
	}
	s.List = ips
}
