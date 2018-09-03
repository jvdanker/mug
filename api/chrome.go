package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"os/exec"
	"runtime"
	"time"
)

func CreateScreenshot(url string) (string, string, error) {
	b, err := run(5*time.Second, url)
	if err != nil {
		return "", "", err
	}

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return "", "", err
	}

	image2 := resize.Resize(100, 0, img, resize.NearestNeighbor)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, image2)
	if err != nil {
		return "", "", err
	}
	b2 := buf.Bytes()

	return "", "data::image/png;base64," + base64.StdEncoding.EncodeToString(b2), nil
}

func run(timeout time.Duration, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Initiate a new RPC connection to the Chrome Debugging Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer domContent.Close()

	lef, err := c.Page.LoadEventFired(ctx)
	if err != nil {
		return nil, err
	}

	// Enable events on the Page domain, it's often preferable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		return nil, err
	}

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs(url)
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		return nil, err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		return nil, err
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	if _, err = lef.Recv(); err != nil {
		return nil, err
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	qsr, err := c.DOM.QuerySelector(ctx, dom.NewQuerySelectorArgs(doc.Root.NodeID, "body"))
	if err != nil {
		return nil, err
	}

	bmr, err := c.DOM.GetBoxModel(ctx, dom.NewGetBoxModelArgs().SetNodeID(qsr.NodeID))
	if err != nil {
		return nil, err
	}

	err = c.Emulation.SetDeviceMetricsOverride(ctx, emulation.NewSetDeviceMetricsOverrideArgs(1024, bmr.Model.Height, 1, false))
	if err != nil {
		return nil, err
	}

	//root, err := dom.GetDocument().Do(ctxt, h)
	//body, err := dom.QuerySelector(root.NodeID, "body").Do(ctxt, h)
	//bm, err := dom.GetBoxModel().WithNodeID(body).Do(ctxt, h)
	//emulation.SetDeviceMetricsOverride(1400, bm.Height, 1, false).Do(ctxt, h)
	//emulation.SetVisibleSize

	// Get the outer HTML for the page.
	//result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
	//	NodeID: &doc.Root.NodeID,
	//})
	//if err != nil {
	//	return err
	//}

	//fmt.Printf("HTML: %s\n", result.OuterHTML)

	// Capture a screenshot of the current page.
	//screenshotName := "screenshot.png"
	screenshotArgs := page.NewCaptureScreenshotArgs().SetFormat("png").SetFromSurface(true)
	screenshot, err := c.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		return nil, err
	}

	//if err = ioutil.WriteFile(screenshotName, screenshot.Data, 0644); err != nil {
	//	return err
	//}

	//fmt.Printf("Saved screenshot: %s\n", screenshotName)

	return screenshot.Data, nil
}

func StartChrome() {
	switch runtime.GOOS {
	case "linux":
		path := "/opt/google/chrome/chrome"
		cmd := exec.Command(path,
			"--remote-debugging-port=9222",
			"--disable-extensions",
			"--disable-default-apps",
			"--disable-sync",
			"--hide-scrollbars",
			"--incognito",
			"--window-size=800,600",
			"--user-data-dir=remote-profile")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return
		} else {
			fmt.Println(string(output))
		}
	case "darwin":
		path := "open"
		cmd := exec.Command(path,
			"-n",
			"-a",
			"Google Chrome",
			"--args",
			"--remote-debugging-port=9222",
			"--disable-extensions",
			"--disable-default-apps",
			"--disable-sync",
			"--hide-scrollbars",
			"--incognito",
			"--window-size=800,600",
			"--user-data-dir=/tmp/Chrome Alt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return
		} else {
			fmt.Println(string(output))
		}
	default:
		panic("Unsupported operating system " + runtime.GOOS)
	}
}
