package env

import (
	"log"
	"sync/atomic"
)

var conf atomic.Value // *Conf

func init() {
	err := Reload()
	if err != nil {
		log.Fatalf("env.init.Reload(%v)", err)
	}
}

func Reload() error {
	newConf, err := ParseConfFromEnv()
	if err != nil {
		return err
	}
	conf.Store(newConf)
	return nil
}

func Get() *Conf {
	return conf.Load().(*Conf)
}
