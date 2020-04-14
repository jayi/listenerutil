# listenerutil
--
    import "dcommon/listenerutil"

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

## Usage

#### func  AddBeginHook

```go
func AddBeginHook(hookFunc http.HandlerFunc)
```
AddBeginHook 添加响应前hook处理方法

#### func  AddEndHook

```go
func AddEndHook(hookFunc EndHookFunc)
```
AddEndHook 添加响应后hook处理方法

#### func  ExtendHandler

```go
func ExtendHandler(handler func(*http.Request) (interface{}, int, error)) http.HandlerFunc
```
ExtendHandler http处理函数，对http.hadlerFunc的封装。 将interface{}解析为json，填到body并响应。
自动添加http头。

#### func  GZipHandler

```go
func GZipHandler(next http.HandlerFunc) http.HandlerFunc
```
GZipHandler http处理函数，对http.hadlerFunc的封装，处理gzip压缩。 请求体为gzip压缩时，解压请求体。
请求允许接收gzip时，使用gzip压缩响应内容。

#### func  ParseBodyParam

```go
func ParseBodyParam(r *http.Request, param interface{}) error
```
ParseBodyParam 按json格式解析请求体body

#### type EndHookFunc

```go
type EndHookFunc func(w http.ResponseWriter, r *http.Request, result interface{}, status int, err error, d time.Duration)
```

EndHookFunc 响应后处理方法。 result : 响应结果; status : 响应状态码; err : 错误信息; d : 响应处理时间;
