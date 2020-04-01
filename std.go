package fetch

import "context"

var defaultFetch = New("", Debug(true))

func Get(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Get(ctx, refPath)
}

func Post(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Post(ctx, refPath)
}

func Put(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Put(ctx, refPath)
}

func Delete(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Delete(ctx, refPath)
}

func Head(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Head(ctx, refPath)
}

func Patch(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Patch(ctx, refPath)
}

func Trace(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Trace(ctx, refPath)
}

func Options(ctx context.Context, refPath string) *Fetch {
	return defaultFetch.Options(ctx, refPath)
}
