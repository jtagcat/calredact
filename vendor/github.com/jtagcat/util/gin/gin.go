package gin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jtagcat/util/std"
)

// c.Cookie(), without ok var (instead, check if empty)
func Cookie(c *gin.Context, name string) string {
	njom, _ := c.Cookie(name)
	return njom
}

// Do not change this outside of init()
var ErrorPage = "error.html"

func HandlerWithErr(f func(c *gin.Context, g *Context) (status int, errStr string)) func(*gin.Context) {
	return func(ctx *gin.Context) {
		code, err := f(ctx, &Context{ctx})
		if err != "" {
			ctx.HTML(code, ErrorPage, gin.H{
				"err": fmt.Sprintf("%d %s: %s", code, http.StatusText(code), err),
			})
			ctx.Abort()
			return
		}

		if code != 0 {
			ctx.Status(code)
		}
	}
}

type Context struct {
	ctx *gin.Context
}

// Wrapper for HandlerWithErr format
func (w *Context) HTML(status int, name string, obj any) (int, string) {
	w.ctx.HTML(status, name, obj)
	return 0, ""
}

// Wrapper HandlerWithErr format
func (w *Context) Redirect(status int, location string) (int, string) {
	w.ctx.Redirect(status, location)
	return 0, ""
}

// Wrapper HandlerWithErr format
func (w *Context) Data(status int, contentType string, data []byte) (int, string) {
	w.ctx.Data(status, contentType, data)
	return 0, ""
}

// Wrapper HandlerWithErr format
func (w *Context) Cookie(name string) string {
	return Cookie(w.ctx, name)
}

func RunWithContext(ctx context.Context, router *gin.Engine) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	slog.Info("starting server", slog.String("address", srv.Addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("starting server", std.SlogErr(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("stopping server gracefully", slog.String("context", ctx.Err().Error()))
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("stopping server", std.SlogErr(err))
		os.Exit(1)
	}
}
