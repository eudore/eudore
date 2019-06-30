package eslogger

import (
	"context"
	"fmt"
	"github.com/olivere/elastic"

	"github.com/eudore/eudore"
)

const (
	ComponentLoggerElasticName    = "logger-elastic"
	ComponentLoggerElasticVersion = "logger-elastic v1.0, output log entry to elasticsearch."
)

type (
	LoggerConfig struct {
		Addr  string
		Index string
	}
	Logger struct {
		*LoggerConfig
		client *elastic.Client
		index  *elastic.IndexService
	}
)

func init() {
	eudore.RegisterComponent(ComponentLoggerElasticName, func(arg interface{}) (eudore.Component, error) {
		return NewLogger(arg)
	})
}

func NewLogger(arg interface{}) (eudore.Logger, error) {
	c, ok := arg.(*LoggerConfig)
	if !ok {
		c = &LoggerConfig{
			Addr:  "http://localhost:9200",
			Index: "eudore",
		}
	}
	// check elastic
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(c.Addr))
	if err != nil {
		return nil, err
	}
	// check index
	exists, err := client.IndexExists(c.Index).Do(context.Background())
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(c.Index).Do(context.Background())
		if err != nil {
			return nil, err
		}
		if !createIndex.Acknowledged {
			return nil, fmt.Errorf("createIndex %s Acknowledged is false.", c.Index)
		}
	}
	return &Logger{
		LoggerConfig: c,
		client:       client,
		index:        client.Index().Index(c.Index).Type("doc"),
	}, nil
}

func (l *Logger) Handle(e eudore.Entry) {
	_, err := l.index.BodyJson(e).Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}

func (l *Logger) WithField(key string, value interface{}) eudore.LogOut {
	return eudore.NewEntryStd(l).WithField(key, value)
}

func (l *Logger) WithFields(fields eudore.Fields) eudore.LogOut {
	return eudore.NewEntryStd(l).WithFields(fields)
}

func (l *Logger) Debug(args ...interface{}) {
	eudore.NewEntryStd(l).Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	eudore.NewEntryStd(l).Info(args...)
}

func (l *Logger) Warning(args ...interface{}) {
	eudore.NewEntryStd(l).Warning(args...)
}

func (l *Logger) Error(args ...interface{}) {
	eudore.NewEntryStd(l).Error(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	eudore.NewEntryStd(l).Fatal(args...)
}

func (l *LoggerConfig) GetName() string {
	return ComponentLoggerElasticName
}

func (l *LoggerConfig) Version() string {
	return ComponentLoggerElasticVersion
}
