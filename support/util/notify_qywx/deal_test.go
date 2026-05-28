package notify_qywx

import (
	"fmt"
	"testing"
	"tools-thinker/support"
)

func TestSendNotify(t *testing.T) {
	notifyStr := fmt.Sprintf(" 环境：%s \n 时间： %s \n 会议：%s \n 事件：%s", "dev", support.FormatDate(support.NowMs(), support.DATE_FORMAT1), "test", "自测")
	SendNotify(notifyStr, "1f6ceb7f-9335-4de3-b95a-8a739faa2ab4")
}
