package main

import (
	"context"
	"fmt"
	"github.com/jvdanker/mug/api"
	mh "github.com/jvdanker/mug/http"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

var stop = make(chan os.Signal, 1)
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
var worker = api.NewWorker()

func main() {
	// start chrome
	// /opt/google/chrome/chrome --remote-debugging-port=9222 --user-data-dir=remote-profile

	signal.Notify(stop, os.Interrupt)

	h := &http.Server{Addr: ":8080", Handler: nil}

	handlers := mh.NewHandlers(stop, worker)
	mh.Handle("/shutdown", handlers.HandleShutdown)
	mh.Handle("/list", handlers.HandleListRequests)
	mh.Handle("/init/", handlers.HandleInitRequests)
	mh.Handle("/pdiff/", handlers.HandlePDiffRequest)
	mh.Handle("/scan", handlers.HandleScanAllRequests)
	mh.Handle("/screenshot/reference/get/", handlers.HandleGetReferenceScreenshot)
	mh.Handle("/screenshot/scan/", handlers.HandleGetScanScreenshot)
	mh.Handle("/url/add", handlers.HandleAddUrl)
	mh.Handle("/url/delete/", handlers.HandleDeleteUrl)
	mh.Handle("/url/scan/", handlers.HandleScanRequests)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(h *http.Server, cancel context.CancelFunc) {
		fmt.Println("Press ctrl+c to interrupt...")

		<-stop

		fmt.Println("Shutting down...")
		h.Shutdown(context.Background())
		cancel()
		logger.Println("Server gracefully stopped")
	}(h, cancel)

	var wg sync.WaitGroup
	wg.Add(1)

	//go startChrome()
	go worker.Worker(ctx, wg)

	logger.Printf("Listening on http://0.0.0.0:8080\n")
	if err := h.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}

	wg.Wait()
}
