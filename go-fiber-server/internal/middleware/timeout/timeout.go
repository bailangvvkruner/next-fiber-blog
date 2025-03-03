package timeout

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"go-fiber-ent-web-layout/internal/tools"
	"go-fiber-ent-web-layout/pkg/pool"
	"log/slog"
	"time"
)

// NewMiddleware 返回请求超时中间件
// 向请求ctx中set一个WithTimeout的Context
func NewMiddleware(timeout time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		beforeTime := time.Now().UnixMilli()
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer func() {
			cancel()
			afterTime := time.Now().UnixMilli()
			slog.Info(fmt.Sprintf("[method:%s uri:%s] 处理耗时：%dms", c.Method(), c.OriginalURL(), afterTime-beforeTime))
		}()
		c.SetUserContext(ctx)
		ch := make(chan error)
		pool.DoGo(context.Background(), func(ctx context.Context, err any) {
			slog.Error("协程池请求处理异常", "error", err)
			ch <- tools.FiberServerError("请求处理失败")
		}, func() {
			ch <- c.Next()
		})
		select {
		// 如果请求正常完成那么直接返回
		case err := <-ch:
			return err
		// 返回请求超时错误
		case <-ctx.Done():
			slog.ErrorContext(ctx, "请求处理超时", "uri", c.OriginalURL(), "method", c.Method())
			return fiber.ErrRequestTimeout
		}
	}
}
