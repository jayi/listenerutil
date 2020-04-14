package listenerutil

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

// GZipHandler http处理函数，对http.hadlerFunc的封装，处理gzip压缩。
// 请求体为gzip压缩时，解压请求体。
// 请求允许接收gzip时，使用gzip压缩响应内容。
func GZipHandler(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// 请求体为压缩过的
			data, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// 解压请求体
				gzr, err := gzip.NewReader(bytes.NewBuffer(data))
				if err != nil {
					// 解压失败使用原始的
					r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
				} else {
					data, err := ioutil.ReadAll(gzr)
					if err == nil {
						r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
					} else {
						fmt.Println("GZipHandler: failed to read body:", err)
					}
				}
			}
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next(gw, r)
	})
}
