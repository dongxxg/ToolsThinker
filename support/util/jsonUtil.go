package util

import (
	"encoding/json"
	"tools-thinker/support/logger"
)

func ConvertToJsonStr(v interface{}) string {
	r, e := json.Marshal(v)
	if e != nil {
		return ""
	}
	return string(r)
}

func MustJsonMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return string([]byte{0})
	}
	return string(b)
}

func MustJsonMarshalByte(v interface{}, defaultValue string) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(defaultValue)
	}
	return b
}

func MustJsonUnmarshalByte(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		logger.Error("json unmarshal err:%s", err.Error())
		return err
	}
	return nil
}

func MustJsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
func Prettify(i interface{}) string {
	resp, _ := json.Marshal(i)
	return string(resp)
}

func PrettifyFormat(i interface{}) string {
	resp, _ := json.MarshalIndent(i, "", " ")
	return string(resp)
}
