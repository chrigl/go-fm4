package main

import (
	"flag"
	"fmt"
	"github.com/chrigl/go-fm4/fm4"
	"io"
	"log"
	"net/http"
	"os"
)

var DownloadFile string

type VerboseDownload struct {
	io.Reader
	total int
}

func (vd *VerboseDownload) Read(p []byte) (int, error) {
	n, err := vd.Reader.Read(p)
	vd.total += n

	if err == nil {
		fmt.Println("Read", n, "bytes")
	}

	return n, err
}

func downloadFile(filename *string, url *string) int64 {
	out, err := os.Create(*filename)
	if err != nil {
		fmt.Println(fmt.Sprint(err))
		panic(err)
	}
	defer out.Close()

	resp, err := http.Get(*url)
	if err != nil {
		fmt.Println(fmt.Sprint(err))
		panic(err)
	}
	defer resp.Body.Close()
	src := VerboseDownload{Reader: resp.Body}

	n, err := io.Copy(out, &src)
	if err != nil {
		fmt.Println(fmt.Sprint(err))
		panic(err)
	}

	return n
}

func printBroadcasts(broadcasts *[]fm4.Broadcasts) {
	for _, x := range *broadcasts {
		fmt.Println(x.DateISO)
		for _, y := range x.Broadcasts {
			fmt.Printf("  * %s: %s - %s\n", y.StartISO, y.ProgramKey, y.Title)
			if y.Description != "" {
				fmt.Printf("    %s\n", y.Description)
			}
		}
	}
}

func initFlags() {
	flag.BoolVar(&fm4.LatestOnly, "l", false, "Only list the latest StreamId")
	flag.StringVar(&fm4.ChannelName, "s", "", "Streams of this channel. e.g. 4TV")
	flag.StringVar(&DownloadFile, "d", "", "Download the latest Stream")
	flag.Parse()
}

func printStreamIds_naive(channelName string, broadcasts *[]fm4.Broadcasts) {
	streams, _ := fm4.SearchBroadcast_naive(channelName, broadcasts)
	streamUrls, nStreamUrls := fm4.FetchStreamIds_naive(channelName, &streams)
	if len(*streamUrls) > 0 {
		if fm4.LatestOnly {
			fm4.PrintStreamId(&(*streamUrls)[nStreamUrls-1])
		} else {
			for _, s := range *streamUrls {
				fm4.PrintStreamId(&s)
			}
		}
	}
}

func processStreamIds(channelName string, broadcasts *[]fm4.Broadcasts, fn func(chan *string, chan bool)) {

	msgChan := make(chan *fm4.Broadcast)
	doneChan := make(chan bool)
	printChan := make(chan *string)
	donePrintChan := make(chan bool)

	go fm4.SearchBroadcast(channelName, broadcasts, msgChan, doneChan)
	go fm4.FetchStreamIds(channelName, msgChan, doneChan, printChan, donePrintChan)

	fn(printChan, donePrintChan)
}

func printStreamIds(channelName string, broadcasts *[]fm4.Broadcasts) {

	fn := func(printChan chan *string, donePrintChan chan bool) {
		emptyString := ""
		var lastStream *string
		lastStream = &emptyString
		for {
			select {
			case msg := <-printChan:
				if fm4.LatestOnly {
					lastStream = msg
				} else {
					fm4.PrintStreamId(msg)
				}
			case <-donePrintChan:
				if fm4.LatestOnly && *lastStream != "" {
					fm4.PrintStreamId(lastStream)
				} else {
					fmt.Println("Unable to find stream")
				}
				return
			}
		}
	}

	processStreamIds(channelName, broadcasts, fn)
}

func downloadLastStream(channelName string, broadcasts *[]fm4.Broadcasts, fileName string) {
	fn := func(printChan chan *string, donePrintChan chan bool) {
		emptyString := ""
		var lastStream *string
		lastStream = &emptyString
		for {
			select {
			case msg := <-printChan:
				lastStream = msg
			case <-donePrintChan:
				if *lastStream == "" {
					fmt.Println("Unable to find stream")
					return
				}
				streamUrl := fm4.GetStreamId(lastStream)
				defer func() {
					if r := recover(); r != nil {
						os.Exit(1)
					}
				}()
				n := downloadFile(&fileName, &streamUrl)
				fmt.Println("Downloaded", n, "bytes")
				return
			}
		}
	}

	processStreamIds(channelName, broadcasts, fn)
}

func main() {
	fm4.BaseUrl = "http://audioapi.orf.at/fm4/json/2.0/broadcasts"
	fm4.LoopBaseUrl = "http://loopstream01.apa.at/?channel=fm4&ua=flash&id="
	initFlags()

	resp, err := fm4.GetBroadcasts()
	if err != nil {
		log.Fatal("Unable to fetch broadcasts ", err)
	}

	if fm4.ChannelName == "" {
		printBroadcasts(&resp)
	} else {
		if DownloadFile != "" {
			downloadLastStream(fm4.ChannelName, &resp, DownloadFile)
		} else {
			printStreamIds(fm4.ChannelName, &resp)
		}
		// printStreamIds_naive(fm4.ChannelName, &resp)
	}
}
