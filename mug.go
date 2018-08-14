package main

import (
	"context"
	"flag"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	var (
		url    = ""
		output = ""
	)

	flag.StringVar(&url, "u", url, "URL")
	flag.StringVar(&output, "o", output, "Output filename")
	flag.Parse()

	if url == "" || output == "" {
		flag.Usage()
		os.Exit(1)
	}

	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Run(ctxt, chromedp.Tasks{
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
	err = c.Run(ctxt, af)

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(output, res, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
