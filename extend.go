/*
Package listenerutil http 服务处理函数封装。

主要功能：

* 将interface{}转化为json并响应。

* 自动添加http响应头，包括允许跨域、content-type等。

* 自动将错误转为errmsg字段返回。

* 响应前后hook支持，可用于记录访问日志、响应时间等。

示例：
	import "dcommon/listenerutil"

	func main() {
		http.HandleFunc("/test", listenerutil.ExtendHandler(testHandler))

		// 添加hook，打印访问日志
		listenerutil.AddEndHook(func(w http.ResponseWriter, r *http.Request, result interface{},
				status int, err error, d time.Duration) {

			fmt.Println(r.Method, r.URL, status, d.Seconds(), r.UserAgent())
		})

		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}


	func testHandler(r *http.Request) (interface{}, int, error) {

		// resp, param 需支持 json 解析
		var param Param
		r.ParseBodyParam(&param)

		resp, err := businessFunc(param)

		// err不为空时，会自动响应400
		return resp, http.StatusOK, err
	}

*/
package listenerutil

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type hookManager struct {
	beginHooks []http.HandlerFunc
	endHooks   []EndHookFunc
}

// EndHookFunc 响应后处理方法。
// result : 响应结果;
// status : 响应状态码;
// err    : 错误信息;
// d      : 响应处理时间;
type EndHookFunc func(w http.ResponseWriter, r *http.Request, result interface{}, status int, err error, d time.Duration)

var hookMgr = &hookManager{
	beginHooks: make([]http.HandlerFunc, 0),
	endHooks:   make([]EndHookFunc, 0),
}

func (hookMgr *hookManager) addBeginHook(hookFunc http.HandlerFunc) {
	hookMgr.beginHooks = append(hookMgr.beginHooks, hookFunc)
}

func (hookMgr *hookManager) addEndHook(hookFunc EndHookFunc) {
	hookMgr.endHooks = append(hookMgr.endHooks, hookFunc)
}

// AddBeginHook 添加响应前hook处理方法
func AddBeginHook(hookFunc http.HandlerFunc) {
	hookMgr.addBeginHook(hookFunc)
}

// AddEndHook 添加响应后hook处理方法
func AddEndHook(hookFunc EndHookFunc) {
	hookMgr.addEndHook(hookFunc)
}

func (hookMgr *hookManager) doBeginHooks(w http.ResponseWriter, r *http.Request) {

	for _, hook := range hookMgr.beginHooks {
		hook(w, r)
	}
}

func (hookMgr *hookManager) doEndHooks(w http.ResponseWriter, r *http.Request, result interface{},
	status int, err error, d time.Duration) {

	for _, hook := range hookMgr.endHooks {
		hook(w, r, result, status, err, d)
	}
}

func doWrapResponse(w http.ResponseWriter, response interface{}, status int, err error) {

	data, ok := response.([]byte)
	if !ok {
		result := make(map[string]interface{}, 1)

		if err != nil || status != http.StatusOK {
			if status == http.StatusOK {
				status = http.StatusBadRequest
			}
			if err == nil {
				err = errors.New(http.StatusText(status))
			}
			result["errno"] = status
			result["errmsg"] = err.Error()
		} else {
			result["data"] = response
			result["errno"] = 0
		}

		var err error
		data, err = json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	if _, isGzip := w.(gzipResponseWriter); !isGzip {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	}
	w.WriteHeader(status)
	w.Write(data)
}

// ExtendHandler http处理函数，对http.hadlerFunc的封装。
// 将interface{}解析为json，填到body并响应。
// 自动添加http头。
func ExtendHandler(handler func(*http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		hookMgr.doBeginHooks(w, r)
		result, status, err := handler(r)
		doWrapResponse(w, result, status, err)
		hookMgr.doEndHooks(w, r, result, status, err, time.Now().Sub(beginTime))
	}
}
