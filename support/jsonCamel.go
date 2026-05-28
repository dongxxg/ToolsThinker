package support

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"tools-thinker/support/logger"
	"unicode"

	"github.com/pkg/errors"
)

// 去除特殊html转义
func JsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// 在转换的内容中增加指定的前缀
func JsonMarshalPrefix(prefix string, t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteString(prefix)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	result := buffer.Bytes() //因为Encoder会在最后增加\n分隔符
	length := len(result)
	return result[0 : length-1], err
}

/*************************************** 下划线json ***************************************/
type JsonSnakeCase struct {
	Value interface{}
}

func (c JsonSnakeCase) MarshalJSON() ([]byte, error) {
	// Regexp definitions
	var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
	var wordBarrierRegex = regexp.MustCompile(`(\w)([A-Z])`)
	marshalled, err := json.Marshal(c.Value)
	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			return bytes.ToLower(wordBarrierRegex.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)
	return converted, err
}

/*************************************** 驼峰json ***************************************/
type JsonCamelCase struct {
	Value interface{}
}

func UmmarshalJSON(byteStr []byte, v interface{}) error {
	var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)

	converted := keyMatchRegex.ReplaceAllFunc(
		byteStr,
		func(match []byte) []byte {
			matchStr := string(match)
			key := matchStr[1 : len(matchStr)-2]
			resKey := Ucfirst(Case2Camel(key))
			return []byte(`"` + resKey + `":`)
		},
	)

	logger.Info("convert to %s", string(converted))
	return json.Unmarshal(converted, v)
}

func MarshalJSON(c *JsonCamelCase) (*[]byte, error) {
	var keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
	marshalled, err := json.Marshal(c.Value)
	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			matchStr := string(match)
			key := matchStr[1 : len(matchStr)-2]
			resKey := Lcfirst(Case2Camel(key))
			return []byte(`"` + resKey + `":`)
		},
	)
	return &converted, err
}

/*************************************** 其他方法 ***************************************/
// 驼峰式写法转为下划线写法
func Camel2Case(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}
	return buffer.Buffer.String()
}

// 下划线写法转为驼峰写法
func Case2Camel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

// 首字母大写
func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// 首字母小写
func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i interface{}) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Buffer.Write(val)
	case rune:
		b.Buffer.WriteRune(val)
	default:
		break
	}
	return b
}

func (b *Buffer) append(s string) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****error******")
		}
	}()
	b.Buffer.WriteString(s)
	return b
}

// RemoveBOM 移除 UTF-8 BOM(字节顺序标记)
// 如果不移除会导致json.Unmarshal报错:invalid character 'ï' looking for beginning of value
func RemoveBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// ReadJsonFile2Struct 读取json文件到结构体
// fileName 文件名
// t 结构体
func ReadJsonFile2Struct[T any](fileName string, t *T) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return errors.Wrapf(err, "read file:%s fail", fileName)
	}
	data = RemoveBOM(data)
	err = json.Unmarshal(data, t)
	if err != nil {
		return errors.Wrapf(err, "file:%s,json unmarshal fail", fileName)
	}
	return nil
}
