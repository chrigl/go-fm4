package fm4

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var LoopBaseUrl string
var BaseUrl string
var LatestOnly bool
var ChannelName string

type Broadcast struct {
	Start           uint      `json:"start"`
	Url             string    `json:"url"`
	Title           string    `json:"title"`
	StartISO        time.Time `json:"startISO"`
	Subtitle        string    `json:"subtitle"`
	ProgramKey      string    `json:"programKey"`
	StartOffset     int       `json:"startOffset"`
	IsPublic        bool      `json:"isPublic"`
	EndISO          time.Time `json:"endISO"`
	EndOffset       int       `json:"endOffset"`
	ScheduledOffset int       `json:"scheduledOffset"`
	Description     string    `json:"description"`
	Scheduled       uint      `json:"scheduled"`
	End             uint      `json:"end"`
	IsBroadcasted   bool      `json:"isBroadcasted"`
}

type Broadcasts struct {
	DateISO    time.Time   `json:"dateISO"`
	DateOffset int         `json:"dateOffset"`
	Day        uint        `json:"day"`
	Date       uint        `json:"date"`
	Broadcasts []Broadcast `json:"broadcasts"`
}

type Channel struct {
	ProgramKey      string           `json:"programKey"`
	Title           string           `json:"title"`
	Subtitle        string           `json:"subtitle"`
	Description     string           `json:"description"`
	IsPublic        bool             `json:"isPublic"`
	IsBroadcasted   bool             `json:"isBroadcasted"`
	Scheduled       uint             `json:"scheduled"`
	ScheduledOffset int              `json:"scheduledOffset"`
	Start           uint             `json:"start"`
	StartISO        time.Time        `json:"startISO"`
	StartOffset     int              `json:"startOffset"`
	End             uint             `json:"end"`
	EndISO          time.Time        `json:"endISO"`
	EndOffset       int              `json:"endOffset"`
	Url             string           `json:"url"`
	Items           []ChannelItems   `json:"items"`
	Streams         []ChannelStreams `json:"streams"`
}

type ChannelItems struct {
	Start         uint      `json:"start"`
	StartISO      time.Time `json:"startISO"`
	StartOffset   int       `json:"startOffset"`
	End           uint      `json:"end"`
	EndISO        time.Time `json:"endISO"`
	EndOffset     int       `json:"endOffset"`
	Type          string    `json:"type"`
	IsPublic      bool      `json:"isPublic"`
	IsBroadcasted bool      `json:"isBroadcasted"`
	Duration      uint      `json:"duration"`
	Title         string    `json:"title"`
	Interpreter   string    `json:"interpreter"`
}

type ChannelStreams struct {
	Start        uint      `json:"start"`
	StartISO     time.Time `json:"startISO"`
	StartOffset  int       `json:"startOffset"`
	End          uint      `json:"end"`
	EndISO       time.Time `json:"endISO"`
	EndOffset    int       `json:"endOffset"`
	Alias        string    `json:"alias"`
	Title        string    `json:"title"`
	LoopStreamId string    `json:"loopStreamId"`
}

type urlError struct {
	s          string
	UrlStr     string
	StatusCode int
}

func (e *urlError) Error() string {
	return strconv.Itoa(e.StatusCode) + ": " + e.UrlStr + " - " + e.s
}

func getBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Unable to get broadcasts ", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		urlErr := &urlError{
			s:          "get not successful",
			UrlStr:     url,
			StatusCode: resp.StatusCode,
		}
		return nil, urlErr
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read broadcasts body data ", err)
		return nil, err
	}

	return res, nil
}

func GetBroadcasts() ([]Broadcasts, error) {
	res, err := getBody(BaseUrl)
	if err != nil {
		return nil, err
	}

	var jres []Broadcasts

	err = json.Unmarshal(res, &jres)
	if err != nil {
		log.Println("Unable to decode ", err)
		return nil, err
	}

	return jres, nil
}

func GetChannel(day string, channelName string) (Channel, error) {
	// http://audioapi.orf.at/fm4/json/2.0/broadcasts/20141229/4UL
	res, err := getBody(BaseUrl + "/" + day + "/" + channelName)

	var jres Channel
	if err != nil {
		return jres, err
	}

	err = json.Unmarshal(res, &jres)
	if err != nil {
		log.Println("Unable to decode ", err)
		return jres, err
	}

	return jres, nil
}

func SearchBroadcast_naive(name string, broadcasts *[]Broadcasts) ([]*Broadcast, int) {
	var res []*Broadcast
	cnt := 0
	for _, y := range *broadcasts {
		for ix, x := range y.Broadcasts {
			if x.ProgramKey == name {
				res = append(res, &y.Broadcasts[ix])
				cnt++
			}
		}
	}
	return res, cnt
}

func SearchBroadcast(name string, broadcasts *[]Broadcasts, msg chan *Broadcast, done chan bool) {
	/* Fake pythons yield :P */
	for _, y := range *broadcasts {
		for ix, x := range y.Broadcasts {
			if x.ProgramKey == name {
				// Not using ptr := &x, since range does not give the real pointer.
				msg <- &y.Broadcasts[ix]
			}
		}
	}

	done <- true
}

func getDateString(dt *time.Time) string {
	/* Just returns %Y%d%m
	 * Time is meant literally
	 * Mon Jan 2 15:04:05 -0700 MST 2006
	 */
	return dt.Format("20060102")
}

func FetchStreamIds_naive(channelName string, broadcasts *[]*Broadcast) (*[]string, int) {
	var res []string
	cnt := 0

	for _, x := range *broadcasts {
		if x.IsBroadcasted {
			dateString := getDateString(&x.StartISO)
			chans, err := GetChannel(dateString, channelName)
			if err != nil {
				log.Println("Unable to fetch channel ", err)
				continue
			}
			for _, x := range chans.Streams {
				res = append(res, x.LoopStreamId)
				cnt++
			}
		}
	}

	return &res, cnt
}

func FetchStreamIds(channelName string, msgChan chan *Broadcast, doneChan chan bool, printChan chan *string, donePrintChan chan bool) {
	for {
		select {
		case msg := <-msgChan:
			if msg.IsBroadcasted {
				dateString := getDateString(&msg.StartISO)
				chans, err := GetChannel(dateString, channelName)
				if err != nil {
					log.Println("Unable to fetch channel ", err)
					continue
				}
				for _, x := range chans.Streams {
					printChan <- &x.LoopStreamId
				}
			}
		case <-doneChan:
			donePrintChan <- true
			return
		}
	}
}

func GetStreamId(streamId *string) string {
	return LoopBaseUrl + *streamId
}

func PrintStreamId(streamId *string) {
	fmt.Println(GetStreamId(streamId))
}
