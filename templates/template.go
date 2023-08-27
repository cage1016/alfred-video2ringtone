package templates

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"
)

//go:generate go-bindata -pkg=template -ignore=.go -nomemcopy  tmpl/...

var engine Engine

type Engine interface {
	init()
	Execute(name string, model interface{}) (string, error)
	ExecuteString(data string, model interface{}) (string, error)
	MustAssetString(name string) string
}

type DefaultEngine struct {
	t *template.Template
}

func funcMap() template.FuncMap {
	return template.FuncMap{}
}
func NewEngine() Engine {
	if engine == nil {
		engine = &DefaultEngine{}
		engine.init()
	}
	return engine
}
func (e *DefaultEngine) init() {
	e.t = template.New("default")
	e.t.Funcs(funcMap())
}

func (e *DefaultEngine) Execute(name string, model interface{}) (string, error) {
	d, err := Asset(fmt.Sprintf("tmpl/%s.tmpl", name))
	if err != nil {
		logrus.Panic(err)
	}
	tmp, err := e.t.Parse(string(d))
	if err != nil {
		logrus.Panic(err)
	}
	ret := bytes.NewBufferString("")
	err = tmp.Execute(ret, model)
	return ret.String(), err
}

func (e *DefaultEngine) ExecuteString(data string, model interface{}) (string, error) {
	tmp, err := e.t.Parse(data)
	if err != nil {
		logrus.Panic(err)
	}
	ret := bytes.NewBufferString("")
	err = tmp.Execute(ret, model)
	return ret.String(), err
}

func (e *DefaultEngine) MustAssetString(name string) string {
	return MustAssetString(name)
}
