package requests

import (
	"net/http"
	"net/url"
	"sync"
)

var (
	requestPool = sync.Pool{
		New: func() interface{} {
			r, _ := http.NewRequest(http.MethodGet, "", nil)
			return r
		},
	}
)

func Put(r *http.Request) {
	requestPool.Put(r)
}

func Get(method string, uri string) *http.Request {
	r := requestPool.Get().(*http.Request)

	r.Method = method
	r.URL, _ = url.Parse(uri)
	r.Header.Set("Host", r.URL.Host)

	return r
}
