package webhook

import (
	"runtime/debug"
	"strconv"
	"sync/atomic"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/openapi"
	log "github.com/sirupsen/logrus"
)

type Frame struct {
	BotId   uint64
	Echo    string
	Ok      bool
	Openapi openapi.OpenAPI
}

var GlobalId int64 = 0

func SafeGo(fn func()) {
	go func() {
		defer func() {
			e := recover()
			if e != nil {
				log.Errorf("err recovered: %+v", e)
				log.Errorf("%s", debug.Stack())
			}
		}()
		fn()
	}()
}

func GenerateId() int64 {
	return atomic.AddInt64(&GlobalId, 1)
}

func GenerateIdStr() string {
	return strconv.FormatInt(GenerateId(), 10)
}
