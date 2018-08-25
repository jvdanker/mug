package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jvdanker/mug/lib"
)

func main() {
	data, err := lib.Run(5*time.Second, "https://www.govt.nz")
	if err != nil {
		log.Fatal(err)
	}

	screenshotName := "screenshot.png"
	if err = ioutil.WriteFile(screenshotName, data, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Saved screenshot: %s\n", screenshotName)
}
