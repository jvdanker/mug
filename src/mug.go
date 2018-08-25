package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	var (
		url     = ""
		output  = ""
		verbose = false
		usage   = false
	)

	flag.StringVar(&url, "u", url, "URL")
	flag.StringVar(&output, "o", output, "Output filename")
	flag.BoolVar(&verbose, "v", verbose, "Verbose output")
	flag.BoolVar(&usage, "?", usage, "Display usage")
	flag.Parse()

	if url == "" || output == "" || usage {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Create snapshot of %v to %v\n", url, output)

	var err error

	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []chromedp.Option{chromedp.WithRunnerOptions(runner.Flag("headless", false))}
	if verbose {
		opts = append(opts, chromedp.WithLog(log.Printf))
	}

	// create chrome instance
	c, err := chromedp.New(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}

	//enableNetworkEvents(ctx, c)
	res, err := createSnapshot(ctx, c, url)

	// shutdown chrome
	err = c.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// TODO this doesn't work in headless mode
	// wait for chrome to finish
	//err = c.Wait()
	//if err != nil {
	//	log.Fatal(err)
	//}

	err = ioutil.WriteFile(output, res, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func createSnapshot(ctx context.Context, c *chromedp.CDP, url string) ([]byte, error) {
	err := c.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(3 * time.Second),
		//chromedp.WaitVisible(".content", chromedp.ByQuery),
	})
	if err != nil {
		log.Fatal(err)
	}

	var res []byte
	af := chromedp.ActionFunc(func(ctxt context.Context, h cdp.Executor) error {
		var err error

		root, err := dom.GetDocument().Do(ctxt, h)
		body, err := dom.QuerySelector(root.NodeID, "body").Do(ctxt, h)
		bm, err := dom.GetBoxModel().WithNodeID(body).Do(ctxt, h)
		emulation.SetDeviceMetricsOverride(1400, bm.Height, 1, false).Do(ctxt, h)
		//emulation.SetVisibleSize

		res, err = page.CaptureScreenshot().WithFromSurface(true).Do(ctxt, h)

		return err
	})
	err = c.Run(ctx, af)

	return res, err
}

func enableNetworkEvents(ctx context.Context, c *chromedp.CDP) error {
	af := chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
		return network.Enable().Do(ctx, h)
	})
	return c.Run(ctx, af)
}
