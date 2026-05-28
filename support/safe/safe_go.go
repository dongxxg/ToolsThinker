package safe

import (
	"runtime/debug"
	logger "tools-thinker/support/logger"
)

// Safego 给要运行的方法加上defer recover，捕获panic并打印错误日志和堆栈信息
// funVar 将要运行的方法
// name 方法的说明，打印错误日志时会带上该信息
func Safego(funVar func(), name string) {
	safego(funVar, name, nil, nil, nil)
}

// SafegoResolve 方法正常退出时，运行回调方法res
func SafegoResolve(funVar func(), name string, res func()) {
	safego(funVar, name, res, nil, nil)
}

// SafegoReject 方法内部发生panic时，运行回调方法rej
func SafegoReject(funVar func(), name string, rej func()) {
	safego(funVar, name, nil, rej, nil)
}

// SafegoFinally 无论是否发生panic，方法结束时都要运行回调方法fin
func SafegoFinally(funVar func(), name string, fin func()) {
	safego(funVar, name, nil, nil, fin)
}

func safego(funVar func(), name string, res func(), rej func(), fin func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("%s panic as: %s\n%s", name, err, debug.Stack())
			if rej != nil {
				rej()
			}
		} else {
			if res != nil {
				res()
			}
		}
		if fin != nil {
			fin()
		}
	}()
	funVar()
}
