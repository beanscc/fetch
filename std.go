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

func Get(ctx context.Context, url string) *Fetch {
	return defaultFetch.Get(ctx, url)
}

func Post(ctx context.Context, url string) *Fetch {
	return defaultFetch.Post(ctx, url)
}

func Put(ctx context.Context, url string) *Fetch {
	return defaultFetch.Put(ctx, url)
}

func Delete(ctx context.Context, url string) *Fetch {
	return defaultFetch.Delete(ctx, url)
}

func Head(ctx context.Context, url string) *Fetch {
	return defaultFetch.Head(ctx, url)
}
