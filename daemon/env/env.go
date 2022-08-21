package env

import (
	"github.com/caarlos0/env"
)

type Conf struct {
	// 레디스 연결 관련 설정은 tsd configmap의 설정을 공유함
	RediAddr string `env:"REDI_ADDR"`
	RediAuth string `env:"REDI_AUTH"`
}

func ParseConfFromEnv() (*Conf, error) {
	conf := &Conf{}
	err := env.Parse(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
