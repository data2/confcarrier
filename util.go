// @description package main
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func ToJsonString(response interface{}) string {
	json1, _ := json.Marshal(response)
	return string(json1)
}

func ToInterface(s string) Record {
	var r Record
	json.Unmarshal([]byte(s), &r)
	return r
}

func md5go(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
