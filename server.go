package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"tgd/cc"
	"tgd/dataloader"
	"tgd/graph"
	"tgd/graph/generated"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err = rc.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	c := generated.Config{Resolvers: &graph.Resolver{}}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	dataloader := dataloader.InitLoader(db, rc, time.Minute)

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json;charset=utf-8")

		if r.Method == "OPTIONS" {
			w.Write([]byte(""))
			return
		}

		cc := cc.New(r, db)

		ctx := cc.Ctx(r.Context())
		ctx = dataloader.Ctx(ctx)

		data, _ := io.ReadAll(r.Body)
		ctx = context.WithValue(ctx, _dataCtxKey, string(data))

		r = r.WithContext(ctx)

		r.Body = io.NopCloser(bytes.NewReader(data))

		srv.ServeHTTP(w, r)
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

const _dataCtxKey = "nzlov@reqData"
