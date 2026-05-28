package http_util

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"time"
	"tools-thinker/support/logger"
	"tools-thinker/support/util"
)

func Post(url string, ops ...Option) ([]byte, error) {
	r := NewRequestParam(ops)
	r.Method = POST
	return getResponseBody(url, r)
}
func Get(url string, ops ...Option) ([]byte, error) {
	r := NewRequestParam(ops)
	r.Method = GET
	return getResponseBody(url, r)
}
func Request(url string, ops ...Option) ([]byte, error) {
	r := NewRequestParam(ops)
	return getResponseBody(url, r)
}

func Down2File(url string, localFile string, ops ...Option) error {
	r := NewRequestParam(ops)
	return down2File(url, localFile, r)
}
func Down2Memory(url string, ops ...Option) ([]byte, string, error) {
	r := NewRequestParam(ops)
	return down2Memory(url, r)
}

func getResponseBody(url string, r *RequestParam) ([]byte, error) {
	var logPrefix = "Request [" + r.Method + "](" + url + ")"
	logger.Debug("%s begin..., and req body: %s", logPrefix, r.GetBodyStr())
	beginTime := time.Now()
	defer func() {
		logger.Debug("%s end and duration: %s", logPrefix, time.Now().Sub(beginTime))
	}()

	// 获取内容
	resp, err := fetchResponse(url, r)
	if err != nil {
		logger.Warn("%s error while fetchResponse as: %s", logPrefix, err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("%s error while read body as: %s", logPrefix, err)
		return nil, err
	}
	logger.Debug("%s resp body is: %s", logPrefix, body)

	// 验证http状态码
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("StatusCode: %d", resp.StatusCode)
		logger.Warn("%s error as %s", logPrefix, err)
		return nil, err
	}
	return body, nil
}

func down2File(url string, localFile string, r *RequestParam) error {
	var logPrefix = "Down2File [" + r.Method + "](" + url + ") to " + localFile
	logger.Debug("%s begin...", logPrefix)
	beginTime := time.Now()
	defer func() {
		logger.Debug("%s end and duration: %s", logPrefix, time.Now().Sub(beginTime))
	}()

	// 获取内容
	res, err := downResponse(url, r)
	if err != nil {
		logger.Error("%s error while fetchResponse as: %s", logPrefix, err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("StatusCode: %d", res.StatusCode)
		logger.Error("%s error as %s", logPrefix, err)
		return err
	}

	// 创建本地文件
	out, err := os.Create(localFile)
	if err != nil {
		logger.Error("%s error while create localFile as: %s", logPrefix, err)
		return err
	}
	defer out.Close()

	// copy到本地文件
	_, err = io.Copy(out, res.Body)
	if err != nil {
		logger.Error("%s error while copy to localFile as: %s", logPrefix, err)
		return err
	}
	return nil
}
func down2Memory(url string, r *RequestParam) ([]byte, string, error) {
	var logPrefix = "Down2Memory [" + r.Method + "](" + url + ")"
	logger.Debug("%s begin...", logPrefix)
	beginTime := time.Now()
	defer func() {
		logger.Debug("%s end and duration: %s", logPrefix, time.Now().Sub(beginTime))
	}()

	// 获取内容
	res, err := downResponse(url, r)
	if err != nil {
		logger.Error("%s error while fetchResponse as: %s", logPrefix, err)
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("StatusCode: %d", res.StatusCode)
		logger.Error("%s error as %s", logPrefix, err)
		return nil, "", err
	}

	// 文件后缀名
	suffix := suffixByResponse(res)

	// 读取文件内容body
	var content []byte
	content, err = ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error("%s error while read body as: %s", logPrefix, err)
		return nil, "", err
	}

	return content, suffix, nil
}

func fetchResponse(url string, r *RequestParam) (*http.Response, error) {
	url = addQuery(url, r.Query)
	ctx, _ := context.WithTimeout(context.Background(), r.Timeout)
	req, _ := http.NewRequest(r.Method, url, r.GetBodyReader())
	req.Header.Set("Content-Type", r.ContentType)
	// 设置其他header参数
	for k, v := range r.header {
		req.Header.Set(k, v)
	}
	return http.DefaultClient.Do(req.WithContext(ctx))
}

func downResponse(url string, r *RequestParam) (*http.Response, error) {
	url = addQuery(url, r.Query)
	ctx, _ := context.WithTimeout(context.Background(), r.Timeout)
	req, _ := http.NewRequest(r.Method, url, r.GetBodyReader())
	return http.DefaultClient.Do(req.WithContext(ctx))
}

// 根据Response获取文件后缀名
// 参考 https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
func suffixByResponse(res *http.Response) string {
	suffix := ""
	// 先根据content-type
	contentType := res.Header.Get("content-type")
	logger.Debug("ContentType is %s while download file", contentType)
	if len(contentType) == 0 {
		logger.Warn("content-type is empty while download file")
	} else {
		switch contentType {
		case "application/vnd.ms-powerpoint":
			suffix = ".ppt"
		case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
			suffix = ".pptx"
		case "application/pdf":
			suffix = ".pdf"
		case "application/msword":
			suffix = ".doc"
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			suffix = ".docx"
		case "application/vnd.ms-excel":
			suffix = ".xls"
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
			suffix = ".xlsx"
		case "video/mp4", "video/mpeg":
			suffix = ".mp4"
		case "audio/mpeg":
			suffix = ".mp3"
		case "image/jpeg":
			suffix = ".jpg"
		case "image/png":
			suffix = ".png"
		case "application/json":
			suffix = ".json"
		default:
			logger.Warn("unknown content-type %s while download file", contentType)
		}
	}

	if len(suffix) > 0 {
		logger.Debug("get file suffix(%s) by content-type(%s)", suffix, contentType)
		return suffix
	}
	// 如果content-type不能确定，再根据Content-Disposition中filename属性
	//Content-Disposition: form-data; name="fieldName"; filename="filename.jpg"
	// content-disposition:attachment; filename=001_1609919687001_lc_2.pdf
	contentDisposition := res.Header.Get("content-disposition")
	logger.Debug("content-disposition is %s while download file", contentDisposition)
	if len(contentDisposition) > 0 {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			logger.Debug("parse content-disposition and params is %s", params)
			if filename := params["filename"]; len(filename) > 0 {
				suffix = path.Ext(filename)
			} else {
				if filename := params["filename*"]; len(filename) > 0 {
					suffix = path.Ext(filename)
				}
			}
		} else {
			logger.Debug("parse content-disposition fail as err is %s", err)
		}
	}

	return suffix
}

func addQuery(url string, query map[string]string) string {
	if query == nil {
		return url
	}

	p, q, _ := util.CutString(url, "?")

	rawQuery, err := url2.ParseQuery(q)
	if err != nil {
		return url
	}

	for k, v := range query {
		rawQuery.Add(k, v)
	}

	return p + "?" + rawQuery.Encode()
}
