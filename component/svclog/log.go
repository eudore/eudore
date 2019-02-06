package svclog

import (
	"fmt"
	"eudore"
	"services/logger/client"
)

var VersionInfo = "eudore logger services " + client.Version()

type (
	SvcLog struct {
		eudore.ComponentName
		Server	string 			`description:"log services server addr."`
		c		*client.Client
		depth	int
	}
)

func init() {
	eudore.RegisterComponent("logger-services", func(arg interface{}) (eudore.Component, error) {
		return NewSvclog(arg)
	})
}

func NewSvclog(arg interface{}) (*SvcLog, error) {
	log := &SvcLog{
		Server:		"localhost:4040",
	}
	if arg != nil {
		c, ok := arg.(eudore.ConfigMap)
		if ok {
			log.Server = c.GetString("server", log.Server)
		}
	}
	con, err := client.NewClient(log.Server) 
	if err != nil {
		return nil, err
	}
	log.c = con
	return log, nil
}

func (l *SvcLog) Debug(args ...interface{}) {
	l.c.Outlog("Logger.Debug",fmt.Sprint(args...), "", nil)
}

func (l *SvcLog) Info(args ...interface{}) {
	l.c.Outlog("Logger.Info",fmt.Sprint(args...), "", eudore.LogFormatFileLineArray(l.depth))
}

func (l *SvcLog) Warning(args ...interface{}) {
	l.c.Outlog("Logger.Warning",fmt.Sprint(args...), "", nil)
}

func (l *SvcLog) Error(args ...interface{}) {
	l.c.Outlog("Logger.Error",fmt.Sprint(args...), "", nil)
}

func (l *SvcLog) Fatal(args ...interface{}) {
	l.c.Outlog("Logger.Fatal",fmt.Sprint(args...), "", nil)
}

func (l *SvcLog) Version() string {
	return VersionInfo
}

func (l *SvcLog) SetOut(addr string) error {
	c, err := client.NewClient(addr) 
	if err != nil {
		return err
	}
	l.c = c
	return nil
}