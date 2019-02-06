package shell


import (
	"time"
	"bytes"
	"regexp"
	"os/exec"

	"eudore"
)

type Shell struct {
	List 	[]string
}

func NewShell() *Shell {
	s := &Shell{}
	go func() {
		s.Update()
		time.Sleep(10 * time.Second)	
	}()
	
	return s
}

func (s *Shell) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "shell")
	ip := ctx.RemoteAddr()
	for _, i := range s.List {
		if ip == i {
			return true, true
		}
	}
	return false, false
}

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

