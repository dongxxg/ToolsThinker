package media

import (
	"fmt"
	"os"
	"testing"
	"tools-thinker/support/logger"
)

func TestM3u8Generator(t *testing.T) {
	var pathList = []string{
		"./testdata/a1.m3u8",
		"./testdata/a2.m3u8",
		"./testdata/a3.m3u8",
		"./testdata/testa4.m3u8",
	}
	g := NewM3u8Generator()
	for _, path := range pathList {
		content := getM3u8Content(path)
		g.AppendM3u8(string(content), "atest")
	}
	m3u8 := g.End()
	fmt.Println(string(m3u8.GetContent()))
}

func getM3u8Content(m3u8Path string) []byte {
	content, err := os.ReadFile(m3u8Path)
	if err == nil {
		return content
	} else {
		logger.Error("testfailed,%v", err)
		return []byte{}
	}
}
