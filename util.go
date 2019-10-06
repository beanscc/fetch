package fetch

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Retry 重试
// 当 fn() 返回错误不是 nil 时，按指定 interval 时间间隔后，重试 fn();
// 若 fn() == nil, 则不进行重试
// n 表示最大重试次数；若重试都失败，则返回最后一次的错误；若重试期间成功，则返回 nil
func Retry(interval time.Duration, maxRetry int, fn func(n int) error) error {
	var err error
	for i := 1; i <= maxRetry; i++ {
		err = fn(i)
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}

	return err
}

// DrainBody reads all of b to memory and then returns bytes of b and a ReadCloser
// yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadCloser have identical error-matching behavior.
func DrainBody(b io.ReadCloser) (rb []byte, nopb io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return nil, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
