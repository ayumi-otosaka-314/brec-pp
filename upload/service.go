package upload

import "github.com/ayumi-otosaka-314/brec-pp/brec"

type Service interface {
	Receive() chan<- *brec.EventDataFileClose
}
