package goloader

import (
	"errors"
	"github.com/thewinds/mkdoc"
)

type GoLoader struct {
	config *mkdoc.ObjectLoaderConfig
}

func (g *GoLoader) Load(ts *mkdoc.TypeScope) (*mkdoc.Object, error) {
	if g.config == nil {
		return nil, errors.New("config not set")
	}
	g.config.
}

func (g *GoLoader) SetConfig(cfg mkdoc.ObjectLoaderConfig) {
	g.config = &cfg
}

func (g *GoLoader) Lang() string {
	return "go"
}
