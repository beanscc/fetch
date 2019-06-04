package fetch

import (
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
