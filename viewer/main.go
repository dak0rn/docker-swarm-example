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
	// TODO This should be in a transaction

	log.Info("Fetching stored host information from the database...")
	keys, err := redis.SMembers(REDIS_LIST_KEY).Result()
	if nil != err {
		return handleError(err, "Failed to fetch the data list", ctx)
	}

	var stored []DockerHostInfo

	for _, key := range keys {
		log.WithField("host", key).Info("Fetching host information")
		datakey, err := redis.Get(REDIS_HOST_PREFIX + key).Result()

		if nil != err {
			return handleError(err, fmt.Sprintf("Failed to fetch key: %s", key), ctx)
		}

		info := DockerHostInfo{}
		err = json.Unmarshal([]byte(datakey), &info)
		if nil != err {
			return handleError(err, "Failed to unmarshal JSON", ctx)
		}

		stored = append(stored, info)
	}

	response, err := json.Marshal(stored)
	if nil != err {
		return handleError(err, "Failed to marshal JSON", ctx)
	}

	ctx.SetContentType("application/json")

	ctx.Write(response)

	return nil
}

func collectInfo(ctx *routing.Context) error {
	log.Info("Retrieved data to store...")
	body := ctx.Request.Body()
	data := DockerHostInfo{}
	redis := ctx.Get("redis").(*redis.Client)

	err := json.Unmarshal(body, &data)
	if nil != err {
		return handleError(err, "Failed to parse the request body", ctx)
	}

	// TODO This should be in a transaction

	// Store the server data first
	err = redis.Set(REDIS_HOST_PREFIX+data.ID, data, 0).Err()
	if nil != err {
		return handleError(err, "Failed to write the host data", ctx)
	}

	// Add the new entry to the set of hosts
	err = redis.SAdd(REDIS_LIST_KEY, data.ID).Err()
	if nil != err {
		return handleError(err, "Failed to write the host to the host list", ctx)
	}

	ctx.SetContentType("application/json")
	ctx.Write([]byte(`{"stored": true}`))
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
	router.Post("/collect", collectInfo)

	log.Info("Starting the web server...")
	fasthttp.ListenAndServe(":3000", router.HandleRequest)
}
