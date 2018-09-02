package main

import (
	"context"
	"fmt"
	"github.com/jvdanker/mug/api"
	"github.com/jvdanker/mug/handler"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	var stop = make(chan os.Signal, 1)
	var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	var worker = api.NewWorker()

	signal.Notify(stop, os.Interrupt)

	h := &http.Server{Addr: ":8080", Handler: nil}

	handlers := handler.NewHandlers(stop, worker)
	handlers.AddHandler("/updates", handlers.HandleGetUpdates)
	handlers.AddHandler("/shutdown", handlers.HandleShutdown)
	handlers.AddHandler("/list", handlers.HandleListRequests)
	handlers.AddHandler("/init/", handlers.HandleInitRequests)
	handlers.AddHandler("/pdiff/", handlers.HandlePDiffRequest)
	handlers.AddHandler("/scan", handlers.HandleScanAllRequests)
	handlers.AddHandler("/screenshot/reference/get/", handlers.HandleGetReferenceScreenshot)
	handlers.AddHandler("/screenshot/scan/", handlers.HandleGetScanScreenshot)
	handlers.AddHandler("/url/add", handlers.HandleAddUrl)
	handlers.AddHandler("/url/scan/", handlers.HandleScanRequests)
	handlers.AddHandler("/url/", handlers.HandleDeleteUrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go api.StartChrome()
	go stopHandler(h, cancel, stop, logger)
	go worker.Worker(ctx, wg)

	logger.Printf("Listening on http://0.0.0.0:8080\n")
	if err := h.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}

	wg.Wait()
}

func stopHandler(h *http.Server, cancel context.CancelFunc, stop <-chan os.Signal, logger *log.Logger) {
	fmt.Println("Press ctrl+c to interrupt...")

	<-stop

	fmt.Println("Shutting down...")
	h.Shutdown(context.Background())
	cancel()
	logger.Println("Server gracefully stopped")
}
