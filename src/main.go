package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// setWebMap sets website and cookie path
// map: webMap[website]cookie_path
func setWebMap(web string, path string) map[string]string {
	webMap := make(map[string]string)
	webs := strings.Split(web, kDELEMITER)
	paths := strings.Split(path, kDELEMITER)

	for idx, w := range webs {
		webMap[w] = paths[idx]
	}

	return webMap
}

func checkinRun(webs map[string]string) (string, error) {
	checkers := make([]*Checkin, 0)

	for webName, webCfg := range webs {
		// NewCheckiner()
		fmt.Println(webName, webCfg)
		if webName[0] == 'T' {
			THY_checker := NewCheckiner(webName, LOGIN_HEADER_ACCEPT, LOGIN_HEADER_CONTENT_TYPE, LOGIN_HEADER_METHOD, kTHY_URL_LOGIN, CHECKIN_HEADER_METHOD, kTHY_URL_CHECKIN, webs[webName])

			checkers = append(checkers, THY_checker)
		} else {
			CUTECLOUD_checker := NewCheckiner(webName, kCUTECLOUD_LOGIN_HEADER_ACCEPT, LOGIN_HEADER_CONTENT_TYPE, LOGIN_HEADER_METHOD, kCUTECLOUD_URL_LOGIN, CHECKIN_HEADER_METHOD, kCUTECLOUD_URL_CHECKIN, webs[webName])

			checkers = append(checkers, CUTECLOUD_checker)
		}
	}

	// Create channel
	ch := make(chan struct{})

	// Timer
	go func(ch chan<- struct{}) {
		// Create timer
		timer := time.NewTicker(kINTEVAL)
		defer func() {
			timer.Stop()
			close(ch)
		}()
		for {
			if _, ok := <-timer.C; !ok {
				log.Println("Timer error")
				return
			}
			ch <- struct{}{}
		}
	}(ch)

	// Checkin
	for {
		if _, ok := <-ch; ok {
			// fmt.Println("It's time to checkin")
			wg := sync.WaitGroup{}
			wg.Add(len(checkers))

			curr_day := time.Time.Day(time.Now())

			for _, checker := range checkers {
				fmt.Printf("curr day: %v %v", curr_day, checker)
				go func(checker *Checkin) {
					defer wg.Done()

					if checker.LastDay != curr_day {
						if checker.Flag_checkined {
							return
						}

						// thy
						log.Printf("%s last_day: %d, curr_day: %d\n", checker.Whoami, checker.LastDay, curr_day)
						if _, ok := webs[checker.Whoami]; ok {
							var err error = nil
							if checker.Whoami[0] == 'T' {
								err = checker.Checkin(CHECKIN_HEADER_ACCEPT, HEADER_CONTENT_LENGTH, kTHY_URL_ORIGIN)
							} else {
								err = checker.Checkin(CHECKIN_HEADER_ACCEPT, HEADER_CONTENT_LENGTH, kCUTECLOUD_URL_ORIGIN)
							}
							if err != nil {
								notifySend("Checkiner", "critical", checker.Whoami+" Check in Failed: "+err.Error())
								checker.Flag_checkined = false
								return
							}
							checker.Flag_checkined = true
							checker.LastDay = curr_day
						} else {
							log.Printf("%s does not exist\n", checker.Whoami)
						}
					} else {
						// fmt.Printf("Checkined tody: %v", curr_day)
						log.Printf("Checkined tody: %v", curr_day)
					}
				}(checker)
			}

			wg.Wait()
		}
	}
}

func init() {
	flag.BoolVar(&h, "h", false, "help")

	flag.StringVar(&web, "w", `THY@THY1@CUTECLOUD`, "set target webs ("+kDELEMITER+" is split char) support: [THY, CUTECLOUD]")
	flag.StringVar(&path, "p", "/home/tianen/go/src/Checkiner/config/THY@/home/tianen/go/src/Checkiner/config/THY_0@/home/tianen/go/src/Checkiner/config/CUTECLOUD",
		"set target webs cookie ("+kDELEMITER+" is split char) support: [THY, CUTECLOUD]")
	flag.Float64Var(&interval, "i", 10, "set checkin interval (minute)")
	flag.StringVar(&kLOG_FILE, "l", "./checkiner.log", "set log file path")

	flag.Usage = usage
}

func main() {
	flag.Parse()
	webs = setWebMap(web, path)
	kINTEVAL = time.Duration(float64(time.Minute) * interval)

	if h || web == "" || path == "" || interval <= 0 {
		flag.Usage()
		return
	}

	// Logger
	log_file, err := os.OpenFile(kLOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Open log file error: ", err)
		return
	}
	defer log_file.Close()

	// Set log information to log file
	log.SetOutput(log_file)

	// Welcome
	notifySend("Checkiner", "normal", "Welcome to enjoy your time with Checkiner")

	// It's time to checkin
	for {
		who, err := checkinRun(webs)
		// Checkiner failed
		if err != nil {
			notifySend("Checkiner", "critical", who+" Check in Failed: "+err.Error())
		}
	}
}

func usage() {
	log.Printf(`Checkiner version: checkiner/1.3.0
Usage: checkiner [-h] [-w web]

Example: checkiner -i 120 -w THY@CUTECLOUD -p /home/tianen/go/src/Checkiner/config/THY@/home/tianen/go/src/Checkiner/config/CUTECLOUD

Options:
`)
	flag.PrintDefaults()
}
