package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	routing "github.com/qiangxue/fasthttp-routing"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"os"
)

const (
	REDIS_LIST_KEY    string = "viewer--datalist"
	REDIS_HOST_PREFIX string = "viewer--key--"
)

// Reads the given flag from the environment
// Panics, if it is unset
func readenv(name string) string {
	value, has := os.LookupEnv(name)
	if !has {
		panic(fmt.Sprintf("Environment variable %s not set", name))
	}

	return value
}

// Returns a fasthttp-routing middleware function that
// sets the given redis client in the request context
func withRedis(client *redis.Client) routing.Handler {
	return func(ctx *routing.Context) error {
		ctx.Set("redis", client)
		ctx.Next()
		return nil
	}
}

func handleError(err error, msg string, ctx *routing.Context) error {
	log.WithField("error", err).Error(msg)
	ctx.SetStatusCode(500)
	ctx.SetBody([]byte(`{"error": "unknown"}`))

	return err
}

func serveHTML(ctx *routing.Context) error {
	ctx.SetContentType("text/html")
	ctx.Write(htmlPage)
	return nil
}

func serveCollectedInfo(ctx *routing.Context) error {
	redis := ctx.Get("redis").(*redis.Client)

	keys, err := redis.SMembers(REDIS_LIST_KEY).Result()
	if nil != err {
		return handleError(err, "Failed to fetch the data list", ctx)
	}

	var stored []string

	for _, key := range keys {
		datakey, err := redis.Get(REDIS_HOST_PREFIX + key).Result()
		if nil != err {
			return handleError(err, fmt.Sprintf("Failed to fetch key: %s", key), ctx)
		}

		stored = append(stored, datakey)
	}

	response, err := json.Marshal(stored)
	if nil != err {
		return handleError(err, "Failed to marshal JSON", ctx)
	}

	ctx.SetContentType("application/json")

	ctx.Write(response)

	return nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
		FullTimestamp:    true,
	})

	log.Info("Starting up...")

	redisAddr := readenv("REDIS_ADDR")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	router := routing.New()

	router.Use(withRedis(redisClient))
	router.Get("/", serveHTML)
	router.Get("/collected", serveCollectedInfo)

	log.Info("Starting the web server...")
	fasthttp.ListenAndServe(":3000", router.HandleRequest)
}
