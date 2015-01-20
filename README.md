# go-fm4

Just a little command line interface for browsing the austrian radio station
fm4. I just wrote it to get a feeling for [Go](https://golang.org).

This program does not do anything useful, but is able to browse the entire playlist, as well as listing stream urls.

## Installation

You may want to set up a valid `GOPATH`:

```
$ mkdir my_go_env
$ cd my_go_env
$ export GOPATH=$PWD
```

Install and build `go-fm4`
```
$ go get github.com/chrigl/go-fm4
$ go build github.com/chrigl/go-fm4
```

## Usage
Browse the entire playlist:
```
$ ./go-fm4
2015-01-13 06:00:00 +0100 CET
  * 2015-01-13 22:00:17 +0100 CET: 4HS - High Spirits
  * 2015-01-14 00:00:00 +0100 CET: 4CZ - Chez Hermes
  * 2015-01-14 01:01:51 +0100 CET: 4SL - Sleepless
    03 - 05: WH von Teilen der FM4 Soundpark-Sendung vom vergangenen Sonntag
2015-01-14 06:00:00 +0100 CET
  * 2015-01-14 05:59:59 +0100 CET: 4MO - Morning Show
...
```

And list all streams of e.g. 4HB:
```
$ ./go-fm4 -s 4HB
http://loopstream01.apa.at/?channel=fm4&ua=flash&id=2015-01-14_2000_tl_54_7DaysWed17__18161.mp3
http://loopstream01.apa.at/?channel=fm4&ua=flash&id=2015-01-14_2000_tl_54_7DaysWed17__18161.mp3
...
```

Or just the latest one:
```
$ ./go-fm4 -s 4HB -l
http://loopstream01.apa.at/?channel=fm4&ua=flash&id=2015-01-20_1959_tl_54_7DaysTue18__18437.mp3
```

It is up to you, to download this file :)
