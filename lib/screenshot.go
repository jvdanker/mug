package lib

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	"log"
	"time"
)

func CreateScreenshot(url string, verbose bool) []byte {
	fmt.Printf("Create snapshot of %v\n", url)

	var err error

	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []chromedp.Option{chromedp.WithRunnerOptions(
		runner.Flag("headless", false),
		runner.Flag("incognito", true),
		runner.Flag("window-size", "800,600"),
		runner.Flag("hide-scrollbars", true),
	)}
	if verbose {
		opts = append(opts, chromedp.WithLog(log.Printf))
	}

	// create chrome instance
	c, err := chromedp.New(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}

	res, err := createSnapshot(ctx, c, url)

	// shutdown chrome
	err = c.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return res
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
