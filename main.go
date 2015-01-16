package main

import (
	"flag"
	"fm4"
	"fmt"
	"log"
)

func printBroadcasts(broadcasts *[]fm4.Broadcasts) {
	for _, x := range *broadcasts {
		fmt.Println(x.DateISO)
		for _, y := range x.Broadcasts {
			fmt.Printf("  * %s (%s) - %s\n", y.Title, y.ProgramKey, y.StartISO)
			if y.Description != "" {
				fmt.Printf("    %s\n", y.Description)
			}
		}
	}
}

func initFlags() {
	flag.BoolVar(&fm4.LatestOnly, "l", false, "Only list the latest StreamId")
	flag.StringVar(&fm4.ChannelName, "s", "", "Streams of this channel. e.g. 4TV")
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

func printStreamIds(channelName string, broadcasts *[]fm4.Broadcasts) {
	msgChan := make(chan *fm4.Broadcast)
	doneChan := make(chan bool)
	printChan := make(chan *string)
	donePrintChan := make(chan bool)

	go fm4.SearchBroadcast(fm4.ChannelName, broadcasts, msgChan, doneChan)
	go fm4.FetchStreamIds(fm4.ChannelName, msgChan, doneChan, printChan, donePrintChan)

	func() {
		var lastStream *string
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
				}
				return
			}
		}
	}()
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
		// printStreamIds(fm4.ChannelName, &resp)
		printStreamIds_naive(fm4.ChannelName, &resp)
	}
}
