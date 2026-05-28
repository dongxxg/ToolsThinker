package support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"tools-thinker/support/logger"

	"reflect"
	"strings"
	"time"
)

/*Info ...
 */

var started bool

type method struct {
	max, count, failed int
	name               string
	method             interface{}
	save               AttachMentRequest
}
type filter struct {
	name   string
	method Filter
}
type server struct {
	methods map[string]method
	filters []filter
}

const (
	filterPass = 0
	filterNext = -1
)

var service = server{methods: make(map[string]method)}

func Register(name string, function interface{}) {
	RegisterAttachment(name, function, nil)
}
func RegisterAttachment(name string, function interface{}, attach AttachMentRequest) {
	if started {
		return
	}
	logger.Info("register %s", name)
	var allMethod = service.methods
	//检查重复
	allMethod[name] = method{0, 0, 0, name, function, attach}
}
func AddFilter(name string, method Filter) {
	if started {
		return
	}
	service.filters = append(service.filters, filter{name, method})
}

type Result struct {
	Code int         `json:"code"`
	Obj  interface{} `json:"obj"`
	Msg  *string     `json:"msg"`
}

func (a Result) Error() string {
	return ""
}

type Filter func(req *http.Request) (*string, int)

func doFilter(req *http.Request, id string) error {
	method := req.URL.Path
	var filters = service.filters
	for i := 0; i < len(filters); i++ {
		filter := &filters[i]
		if strings.Contains(method, filter.name) {
			msg, code := filter.method(req)
			if msg != nil {
				return Result{Code: code, Msg: msg}
			}
			if code == filterPass {
				return nil
			}
		}
	}
	return nil
}

/*Call ...
 */

func call(name string, req *http.Request, id string) (*reflect.Value, error) {
	var allMethod = service.methods
	var method, ok = allMethod[name]
	if !ok {
		//	return [], nil
	}
	fn := method.method
	argv, _ := construcArgument(fn)
	var begin = time.Now()
	contentType := req.Header.Get("Content-Type")
	var args reflect.Value

	if strings.Contains(contentType, "json") {
		args, _ = getArguments(req.Body, argv)
	} else {
		r, err := req.MultipartReader()
		if err != nil {
			return nil, err
		}
		var json bytes.Buffer
		var files = make(map[string]string)

		for {
			p, err := r.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			filename := p.FileName()
			contentType, hasContentTypeHeader := p.Header["Content-Type"]
			if hasContentTypeHeader {
				if strings.Index(contentType[0], "json") >= 0 {
					maxValueBytes := int64(10 << 20)
					// value, store as string in memory
					n, err := io.CopyN(&json, p, maxValueBytes+1)
					if err != nil && err != io.EOF {
						return nil, err
					}
					maxValueBytes -= n
					if maxValueBytes < 0 {
						return nil, nil
					}
				} else {
					name := p.FormName()
					if name == "" {
						logger.Warn("form name is null")
						continue
					}
					var dir = method.save.GetLocation(name)
					file, err := ioutil.TempFile(dir, filename+"_")
					if err != nil {
						return nil, err
					}
					_, err = io.Copy(file, p)
					if cerr := file.Close(); err == nil {
						err = cerr
					}
					if err != nil {
						os.Remove(file.Name())
						return nil, err
					}
					files[name] = file.Name()
					logger.Info("req[%s] save %s to %s", id, name, file.Name())
				}
			}
		}
		bs := json.Bytes()
		index := bytes.LastIndexByte(bs, byte('}'))
		if index > 0 {
			bs = bs[:index]
		} else {
			//Warn("there is no json value in the request")
		}
		json1 := bytes.NewBuffer(bs)
		if index <= 0 {
			json1.WriteByte('{')
		}
		fileIndex := 0
		for k, v := range files {
			if !(index <= 0 && fileIndex == 0) {
				json1.WriteString(",")
			}
			json1.WriteString("\"")
			json1.WriteString(k)
			json1.WriteString(`":"`)
			json1.WriteString(v)
			json1.WriteByte('"')
			fileIndex++
		}
		json1.WriteByte('}')
		args, err = getArguments(json1, argv)
		if err != nil {
			logger.Warn("decode failed %v", err)
		}
	}
	v := reflect.ValueOf(fn)
	result := v.Call([]reflect.Value{args, reflect.ValueOf(id)})
	var end = time.Now()
	var delta = end.Sub(begin)
	_ = delta
	/*if delta > method.max {

	}*/
	return &result[0], nil
}

func construcArgument(fn interface{}) (reflect.Value, reflect.Value) {
	funcType := reflect.TypeOf(fn)
	index := 0
	argt := funcType.In(index)
	argv := reflect.New(argt)
	subv := argv.Elem()
	for subv.Kind() == reflect.Ptr {
		t := subv.Type()
		fmt.Printf("%v\n", t)
		fmt.Printf("%v\n", t.Elem())
		if subv.CanSet() {
			fmt.Printf("can set \n")
			subv.Set(reflect.New(t.Elem()))
		} else {
			fmt.Printf("can set false\n")
		}
		subv = subv.Elem()
	}
	return argv, subv
}
func getArguments(r io.Reader, argv reflect.Value) (reflect.Value, error) {
	dec := json.NewDecoder(r)
	err := dec.Decode(argv.Interface())
	if err != nil {
		return argv.Elem(), err
	}
	return argv.Elem(), nil
}
