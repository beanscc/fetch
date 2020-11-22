package fetch

import (
	"context"
	"time"
)

var defaultFetch = New("", &Options{
	Timeout: 15 * time.Second,
	Interceptors: []Interceptor{
		DefaultLogInterceptor,
	},
})

func Get(ctx context.Context, url string, params ...interface{}) *Fetch {
	return defaultFetch.Get(ctx, url, params...)
}

func Post(ctx context.Context, url string, params ...interface{}) *Fetch {
	return defaultFetch.Post(ctx, url, params...)
}

func Put(ctx context.Context, url string, params ...interface{}) *Fetch {
	return defaultFetch.Put(ctx, url, params...)
}

func Delete(ctx context.Context, url string, params ...interface{}) *Fetch {
	return defaultFetch.Delete(ctx, url, params...)
}

func Head(ctx context.Context, url string, params ...interface{}) *Fetch {
	return defaultFetch.Head(ctx, url, params...)
}
