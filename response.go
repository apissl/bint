package bint

import (
	"fmt"
	"reflect"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

const (
	Success = "success"
	Fail    = "fail"
)

// Ret 快速格式化返回
// 支持类型：error、string、int、struct、map、切片，其中：error、string 为错误
// 通常1个入参，如果有2个入参，则参数1为“code”、参数2为错误（error、string）
func (c *Context) Ret(args ...any) error {
	resp := Response{Code: -1, Msg: "abnormal parameter of Ret()"}
	switch len(args) {
	case 0:
		resp = Response{Code: 0, Msg: Success}
	case 1:
		switch reflect.TypeOf(args[0]).Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			resp = Response{Code: 0, Msg: Success, Data: args[0]}
		case reflect.Ptr:
			if err, ok := args[0].(error); ok {
				resp = Response{Code: -1, Msg: err.Error()}
			} else {
				resp = Response{Code: 0, Msg: Success, Data: args[0]}
			}
		case reflect.Int:
			resp = Response{Code: args[0].(int), Msg: Fail}
		case reflect.String:
			resp = Response{Code: -1, Msg: args[0].(string)}
		default:
			resp = Response{Code: -1, Msg: fmt.Sprintf("%v", args[0])}
		}
	case 2:
		if code, ok := args[0].(int); ok {
			switch args[1].(type) {
			case string:
				resp = Response{Code: code, Msg: args[1].(string)}
			case error:
				resp = Response{Code: code, Msg: args[1].(error).Error()}
			default:
				resp = Response{Code: code, Msg: fmt.Sprintf("%v", args[1])}
			}
		}
	}
	return c.Ctx.JSON(resp)
}
