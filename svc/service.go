package svc

import (
	"github.com/tespkg/bytes-be/config"
	"os"
)

type Service interface {
	Name() string
	Load(config config.Config) error
	Run(readyChan chan bool)
	Stop(signal os.Signal)
}
