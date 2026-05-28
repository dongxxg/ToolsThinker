package support

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"tools-thinker/support/logger"

	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GetConfig(conf *string, config interface{}) {
	if conf == nil {
		conf = getPath()
	}
	logger.Info("get config from %s\n", *conf)
	content, err := ioutil.ReadFile(*conf)
	if err != nil {
		logger.Error("open %s failed,%v\n", *conf, err)
		return
	}
	err = json.Unmarshal(content, config)
	if err != nil {
		logger.Error("parse config file failed %s", err)
	}
}

func GetConfigFromServer(conf *string, configServer map[string]string, config interface{}) {
	if conf == nil {
		conf = getPath()
	}
	var count = 0
	for {
		count++
		var params = make(map[string]string)
		params["branch"] = configServer["branch"]
		params["path"] = *conf

		var url = "http://" + configServer["host"] + "/config/getConfig?" + GetQueryString(params, nil)
		logger.Info("get config from %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			logger.Info("get config failed %v\n", err)
			if count > 5 {
				logger.Fatal("get config failed, exit\n")
			}
			logger.Info("get config retry in %v s\n", count)
			time.Sleep(time.Duration(count) * time.Second)
			continue
		}
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respByte := buf.Bytes()

		err = json.Unmarshal(respByte, config)
		if err != nil {
			logger.Info("parse config file failed %v", err)
			if count > 5 {
				logger.Fatal("parse config failed, exit\n")
			}
			logger.Info("get config retry in %v s\n", count)
			time.Sleep(time.Duration(count) * time.Second)
			continue
		}
		break
	}
}
func getPath() *string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		logger.Error("%s\n", err)
	}
	path, err := filepath.Abs(file)
	if err != nil {
		logger.Fatal("%s\n", err)
	}
	dirs := strings.Split(path, string(filepath.Separator))

	prefix := os.Getenv("PLASO_DIR")
	if prefix == "" {
		if len(dirs) < 2 {
			prefix = path
		} else {
			prefix = filepath.Join(string(filepath.Separator), dirs[1], dirs[2]) //dirs[0]=="" skip it;
		}
	}

	prefix = filepath.Join(prefix, "conf", dirs[len(dirs)-1]+".conf")
	return &prefix
}
