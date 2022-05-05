package main

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func CommonResult(w http.ResponseWriter, code string, msg string) {
	Result(w, code, msg, nil)
}

func ResultSuccess(w http.ResponseWriter, obj interface{}) {
	Result(w, "0", "success", obj)
}

func Result(w http.ResponseWriter, code string, msg string, obj interface{}) {
	// 当返回json时, 添加json头
	w.Header().Add("Content-Type", "application/json")
	result := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}
	if obj != nil {
		vi := reflect.ValueOf(obj)
		if vi.Kind() == reflect.Interface {
			if !vi.IsNil() {
				result["data"] = obj
			}
		} else {
			result["data"] = obj
		}
	}
	_ = json.NewEncoder(w).Encode(result)
}
