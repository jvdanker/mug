package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/jvdanker/mug/lib"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"os/exec"
	"runtime"
	"time"
)

func CreateScreenshot(url string) (string, string, error) {
	b, err := lib.Run(5*time.Second, url)
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

func startChrome() {
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
