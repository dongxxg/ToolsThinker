package support

import (
	"runtime/debug"
	"testing"
	"tools-thinker/support/logger"
)

func TestPanicLog(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			logger.Debug("%s ", logger.PanicLog(e, debug.Stack()))
		}
	}()
	// TT
	var f func() = nil
	f()
}
