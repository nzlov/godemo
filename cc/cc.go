package cc

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

const _ctxKey = "nzlov@cc"

func For(ctx context.Context) *Context {
	return ctx.Value(_ctxKey).(*Context)
}

type Context struct {
	db *gorm.DB
	r  *http.Request
}

func New(r *http.Request, db *gorm.DB) *Context {
	return &Context{
		r:  r,
		db: db,
	}
}

func (c *Context) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, c)
}

func (c *Context) DB() *gorm.DB {
	return c.db
}
