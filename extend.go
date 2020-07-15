/*
Package listenerutil http 服务处理函数封装。

主要功能：

* 将interface{}转化为json并响应。

* 自动添加http响应头，包括允许跨域、content-type等。

* 自动将错误转为errmsg字段返回。

* 响应前后hook支持，可用于记录访问日志、响应时间等。

示例：
	import "github.com/jayi/listenerutil"

	func main() {
		http.HandleFunc("/test", listenerutil.ExtendHandler(testHandler))

		// 添加hook，打印访问日志
		listenerutil.AddEndHandleFunc(func(w http.ResponseWriter, r *http.Request, result *handlerResult) {

			fmt.Println(r.Method, r.URL, result.StatusCode, result.Cost.Seconds(), r.UserAgent())
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
	"strings"
	"time"
)

type handlerManager struct {
	beginHooks       []http.HandlerFunc
	endHooks         []EndHandleFunc
	dataFieldName    string
	codeFieldName    string
	msgFieldName     string
	allowCrossOrigin bool
}

// HandleResult 响应结果相关信息
type HandleResult struct {
	Data       interface{}
	StatusCode int
	Err        error
	Cost       time.Duration
}

// EndHookFunc 响应后处理方法。
// result : 响应结果;
// status : 响应状态码;
// err    : 错误信息;
// d      : 响应处理时间;
type EndHookFunc func(w http.ResponseWriter, r *http.Request, result interface{}, status int, err error, d time.Duration)

// EndHandleFunc 新版响应后处理方法。
// result : 响应结果，包含响应数据，状态码，错误信息，响应处理时间等
type EndHandleFunc func(w http.ResponseWriter, r *http.Request, result *HandleResult)

const (
	defaultDataFieldName = "data"
	defaultCodeFieldName = "errno"
	defaultMsgFieldName  = "errmsg"
)

var handlerMgr = &handlerManager{
	beginHooks:       make([]http.HandlerFunc, 0),
	endHooks:         make([]EndHandleFunc, 0),
	dataFieldName:    defaultDataFieldName,
	codeFieldName:    defaultCodeFieldName,
	msgFieldName:     defaultMsgFieldName,
	allowCrossOrigin: true,
}

func (handlerMgr *handlerManager) addBeginHook(hookFunc http.HandlerFunc) {
	handlerMgr.beginHooks = append(handlerMgr.beginHooks, hookFunc)
}

func (handlerMgr *handlerManager) addEndHook(hookFunc EndHookFunc) {
	handlerMgr.endHooks = append(handlerMgr.endHooks, func(w http.ResponseWriter, r *http.Request, result *HandleResult) {
		hookFunc(w, r, result.Data, result.StatusCode, result.Err, result.Cost)
	})
}

func (handlerMgr *handlerManager) addEndHandleFunc(hookFunc EndHandleFunc) {
	handlerMgr.endHooks = append(handlerMgr.endHooks, hookFunc)
}

func (handlerMgr *handlerManager) setDataFieldName(name string) error {
	if len(name) == 0 {
		return errors.New("invalid field name: " + name)
	}
	if name == handlerMgr.codeFieldName || name == handlerMgr.msgFieldName {
		return errors.New("duplicate filed name: " + name)
	}
	handlerMgr.dataFieldName = name
	return nil
}

func (handlerMgr *handlerManager) setCodeFieldName(name string) error {
	if len(name) == 0 {
		return errors.New("invalid field name: " + name)
	}
	if name == handlerMgr.dataFieldName || name == handlerMgr.msgFieldName {
		return errors.New("duplicate filed name: " + name)
	}
	handlerMgr.codeFieldName = name
	return nil
}

func (handlerMgr *handlerManager) setMsgFieldName(name string) error {
	if len(name) == 0 {
		return errors.New("invalid field name: " + name)
	}
	if name == handlerMgr.dataFieldName || name == handlerMgr.codeFieldName {
		return errors.New("duplicate filed name: " + name)
	}
	handlerMgr.msgFieldName = name
	return nil
}

func (handlerMgr *handlerManager) setAllowCrossOrigin(allow bool) {
	handlerMgr.allowCrossOrigin = allow
}

func (handlerMgr *handlerManager) doBeginHooks(w http.ResponseWriter, r *http.Request) {

	for _, hook := range handlerMgr.beginHooks {
		hook(w, r)
	}
}

func (handlerMgr *handlerManager) doEndHooks(w http.ResponseWriter, r *http.Request, result *HandleResult) {

	for _, hook := range handlerMgr.endHooks {
		hook(w, r, result)
	}
}

// AddBeginHook 添加响应前hook处理方法
func AddBeginHook(hookFunc http.HandlerFunc) {
	handlerMgr.addBeginHook(hookFunc)
}

// AddEndHook 添加响应后hook处理方法
func AddEndHook(hookFunc EndHookFunc) {
	handlerMgr.addEndHook(hookFunc)
}

// AddEndHandleFunc 添加响应后hook处理新方法
func AddEndHandleFunc(hookFunc EndHandleFunc) {
	handlerMgr.addEndHandleFunc(hookFunc)
}

// SetDataFieldName 设置响应数据字段key名
func SetDataFieldName(name string) error {
	return handlerMgr.setDataFieldName(name)
}

// SetCodeFieldName 设置响应码字段key名
func SetCodeFieldName(name string) error {
	return handlerMgr.setCodeFieldName(name)
}

// SetMsgFieldName 设置响应错误信息key名
func SetMsgFieldName(name string) error {
	return handlerMgr.setMsgFieldName(name)
}

// SetAllowCrossOrigin 设置是否允许跨域
func SetAllowCrossOrigin(allow bool) {
	handlerMgr.setAllowCrossOrigin(allow)
}

const (
	credentialsTrue               = "true"
	defaultOriginValue            = "*"
	originRequestHeader           = "Origin"
	accessControlRequestHeaders   = "Access-Control-Request-Headers"
	accessControlRequestMethod    = "Access-Control-Request-Method"
	accessControlAllowOrigin      = "Access-Control-Allow-Origin"
	accessControlAllowCredentials = "Access-Control-Allow-Credentials"
	accessControlAllowHeaders     = "Access-Control-Allow-Headers"
	accessControlAllowMethods     = "Access-Control-Allow-Methods"
)

//处理跨域
func doAccessOrigin(w http.ResponseWriter, r *http.Request) {

	origin := r.Header.Get(originRequestHeader)

	if len(strings.TrimSpace(origin)) <= 0 {
		origin = defaultOriginValue
	}

	w.Header().Set(accessControlAllowOrigin, origin)
	w.Header().Set(accessControlAllowCredentials, credentialsTrue)

	if r.Method == http.MethodOptions {
		w.Header().Set(accessControlAllowMethods, r.Header.Get(accessControlRequestMethod))
		w.Header().Set(accessControlAllowHeaders, r.Header.Get(accessControlRequestHeaders))
	}
}

// WrapResponse 将interface{}转为json写入http.ResponseWriter
func WrapResponse(w http.ResponseWriter, response interface{}, status int, err error) {
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
			result[handlerMgr.codeFieldName] = status
			result[handlerMgr.msgFieldName] = err.Error()
		} else {
			result[handlerMgr.dataFieldName] = response
			result[handlerMgr.codeFieldName] = 0
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
	if _, isGzip := w.(gzipResponseWriter); !isGzip {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	}
	w.WriteHeader(status)
	w.Write(data)
}

// ExtendHandler http处理函数，对http.handlerFunc的封装。
// 将interface{}解析为json，填到body并响应。
// 自动添加http头。
func ExtendHandler(handler func(*http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		handlerMgr.doBeginHooks(w, r)
		if handlerMgr.allowCrossOrigin {
			doAccessOrigin(w, r)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		data, status, err := handler(r)
		WrapResponse(w, data, status, err)
		handleResult := &HandleResult{
			Data:       data,
			StatusCode: status,
			Err:        err,
			Cost:       time.Now().Sub(beginTime),
		}
		handlerMgr.doEndHooks(w, r, handleResult)
	}
}
