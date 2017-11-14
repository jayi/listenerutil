package listen

import (
	"net/http"
	"strconv"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var beginHooks []http.HandlerFunc
var endHooks []endHookFunc
var mutex sync.RWMutex

type endHookFunc func(w http.ResponseWriter, r *http.Request, result interface{}, status int, err error, d time.Duration)

func init() {
	beginHooks = make([]http.HandlerFunc, 0)
	endHooks = make([]endHookFunc, 0)
}

func AddBeginHook(hookFunc http.HandlerFunc) {
	mutex.Lock()
	beginHooks = append(beginHooks, hookFunc)
	mutex.Unlock()
}

func AddEndHook(hookFunc endHookFunc) {
	mutex.Lock()
	endHooks = append(endHooks, hookFunc)
	mutex.Unlock()
}

func doBeginHooks(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	for _, hook := range beginHooks {
		hook(w, r)
	}
	mutex.RUnlock()
}

func doEndHooks(w http.ResponseWriter, r *http.Request, result interface{}, status int, err error, d time.Duration) {
	mutex.RLock()
	for _, hook := range endHooks {
		hook(w, r, result, status, err, d)
	}
	mutex.RUnlock()
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
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if _, isGzip := w.(gzipResponseWriter); !isGzip {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	}
	w.WriteHeader(status)
	w.Write(data)
}

func ExtendHandler(handler func(*http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		doBeginHooks(w, r)
		result, status, err := handler(r)
		doWrapResponse(w, result, status, err)
		doEndHooks(w, r, result, status, err, time.Now().Sub(beginTime))
	}
}
