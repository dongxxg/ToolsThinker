package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"tools-thinker/support/logger"
)

const CONTENT_TYPE_MIME_JSON = "application/json"
const CONTENT_TYPE_MIME_FORM = "application/x-www-form-urlencoded"
const TimeOut = 20 * time.Second

func HttpPost(url string, contentType string, data []byte) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), TimeOut)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		logger.Error("HttpPost url:%s error  %s", url, err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("HttpPost url:%s readall error  %s", url, err)
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		logger.Error(
			"HttpPost url:%s , resp code = %d, resp body = %s",
			url,
			resp.StatusCode,
			string(body),
		)
		return "", err
	}
	// logger.Debug("HttpPost url:%s response: %s", url, responseStr)
	return string(body), nil
}

func HttpGetstr(url string) (string, error) {
	body, err := HttpGet(url)
	if err != nil {
		return "", err
	}
	responseStr := string(body)
	// logger.Debug("HttpGetstr url:%s response: %s", url, responseStr)
	return responseStr, nil
}

func HttpGet(url string) ([]byte, error) {
	resp, err := getResponse(url)

	if err != nil {
		fmt.Println(err)
		logger.Error("HttpGet url:%s error %s", url, err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}

// HttpGetFile 下载文件到内存
func HttpGetFile(url string) (content []byte, suffix string, err error) {
	resp, err := getResponse(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	suffix = suffixByResponse(resp)

	content, err = ioutil.ReadAll(resp.Body)
	return
}

// 下载指定url内容到文件
func Down(url string, localFile string) error {
	//这里应该用http.Get方法,然后直接将resposewrite到文件,减少一个内容的拷贝;
	//系统提供的Get方法返回的reader, 跟后续链接的方法应该更好结合;
	begintime := time.Now().Unix()
	defer func() {
		endtime := time.Now().Unix()
		fmt.Printf("down file cost %v s \n", endtime-begintime)
	}()
	res, err := getResponse(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 根据Response设置后缀名，后缀名优先使用Response中提取到的
	suffix := suffixByResponse(res)
	localFile = replaceSuffix(localFile, suffix)

	out, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)

	return err
}

// Get请求获取Response，默认超时时间20s
func getResponse(url string) (*http.Response, error) {
	ctx, _ := context.WithTimeout(context.Background(), TimeOut)
	req, _ := http.NewRequest("GET", url, nil)
	return http.DefaultClient.Do(req.WithContext(ctx))
}

// 根据Response获取文件后缀名
// 参考 https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
func suffixByResponse(res *http.Response) string {
	suffix := ""
	// 先根据content-type
	contentType := res.Header.Get("content-type")
	logger.Debug("contentType is %s while download file", contentType)
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

// 替换文件后缀名
func replaceSuffix(localFile, suffix string) string {
	if len(suffix) == 0 {
		return localFile
	}
	oldSuffix := path.Ext(localFile)
	if len(oldSuffix) > 0 {
		localFile = strings.TrimRight(localFile, oldSuffix)
	}
	return localFile + suffix
}

func HttpPostHeader(url string, header http.Header, data []byte) (string, int, error) {
	return HttpHeader("POST", url, header, data)
}

func HttpPostHeaderWithRespHeader(
	url string,
	header http.Header,
	data []byte,
) (string, http.Header, int, error) {
	return HttpHeaderWithRespHeader("POST", url, header, data)
}
func HttpGetHeader(url string, header http.Header, data []byte) (string, int, error) {
	return HttpHeader("GET", url, header, data)

}
func HttpHeader(method string, url string, header http.Header, data []byte) (string, int, error) {
	ctx, _ := context.WithTimeout(context.Background(), TimeOut)
	req, _ := http.NewRequest(method, url, bytes.NewReader(data))
	req.Header = header
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println(err)
		logger.Error("HttpPost url:%s error  %s", url, err)
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	responseStr := string(body)
	// logger.Debug("HttpPost url:%s response: %s", url, responseStr)
	statusCode := resp.StatusCode
	if statusCode != 200 {
		return responseStr, statusCode, errors.New(
			fmt.Sprintf("url:%s,statusCode : %d", url, statusCode),
		)
	}
	return responseStr, statusCode, nil
}

func HttpHeaderWithRespHeader(
	method string,
	url string,
	header http.Header,
	data []byte,
) (string, http.Header, int, error) {
	ctx, _ := context.WithTimeout(context.Background(), 20000*time.Millisecond)
	req, _ := http.NewRequest(method, url, bytes.NewReader(data))
	req.Header = header
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println(err)
		logger.Error("HttpHeaderWithRespHeader url:%s error  %s", url, err)
		return "", resp.Header, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp.Header, 0, err
	}
	responseStr := string(body)
	// logger.Debug("HttpPost url:%s response: %s", url, responseStr)
	statusCode := resp.StatusCode
	if statusCode != 200 {
		return responseStr, resp.Header, statusCode, errors.New(
			fmt.Sprintf("url:%s,statusCode : %d", url, statusCode),
		)
	}
	return responseStr, resp.Header, statusCode, nil
}

func GetArgsInt(param url.Values, name string) *int {
	nameStrs := param.Get(name)
	if len(nameStrs) == 0 {
		return nil
	} else {
		result, err := strconv.Atoi(nameStrs)
		if err != nil {
			return nil
		}
		return &result
	}
}
