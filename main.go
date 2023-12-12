package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"template.com/service/src/logger"
	"template.com/service/src/modules/renderengine"
	"template.com/service/src/server"
)

var RenderEngine *renderengine.RenderEngineStruct
var Logger *logger.LoggerEngineStruct

func main() {
	funcname := "main"
	fmt.Println("starting server...")
	defer fmt.Println("shutting down server...")
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf(logger.LogError(funcname, err, "didn't successfully load env variables"))
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// create a channel to capture SIGTERM, SIGINT signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	defer fmt.Println(ctx.Err())
	Logger = logger.NewLoggerEngine(ctx)

	var wg sync.WaitGroup
	ServerInstance := iris.Default()
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := Logger.NewLogThread(); err != nil {
			fmt.Print(err)
			panic(err)
		}
	}()

	ServerInstance.Use(server.BannedIpHandler, iris.Compression)

	// Serve static files
	// "check if request is for static file first"
	ServerInstance.HandleDir("/", iris.Dir("./public/"), iris.DirOptions{
		Compress:  true,
		IndexName: "index.html",
		Cache: iris.DirCacheOptions{
			Enable:          true,
			Encodings:       []string{"gzip"},
			CompressIgnore:  iris.MatchImagesAssets,
			CompressMinSize: 30 * iris.B,
		},
	})

	RenderEngine, err = renderengine.NewRenderer(ctx)
	if err != nil {
		err = fmt.Errorf(logger.LogError(funcname, err, "couldn't build renderengine"))
		fmt.Println(err.Error())
		os.Exit(1)
	}

	visitorsess := sessions.New(sessions.Config{
		Cookie:       "_frontdesk_visitor",
		AllowReclaim: true,
	})
	ServerInstance.Use(visitorsess.Handler())

	wg.Wait()
}
