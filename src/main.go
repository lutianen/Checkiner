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
	// Create checkiner
	THY_checker := NewCheckiner(kTHY_WHOAMI, kTHY_LOGIN_HEADER_ACCEPT, kTHY_LOGIN_HEADER_CONTENT_TYPE, kTHY_LOGIN_HEADER_METHOD, kTHY_URL_LOGIN, kTHY_CHECKIN_HEADER_METHOD, kTHY_URL_CHECKIN, webs[kTHY_WHOAMI])

	CUTECLOUD_checker := NewCheckiner(kCUTECLOUD_WHOAMI, kCUTECLOUD_LOGIN_HEADER_ACCEPT, kCUTECLOUD_LOGIN_HEADER_CONTENT_TYPE, kCUTECLOUD_LOGIN_HEADER_METHOD, kCUTECLOUD_URL_LOGIN, kCUTECLOUD_CHECKIN_HEADER_METHOD, kCUTECLOUD_URL_CHECKIN, webs[kCUTECLOUD_WHOAMI])

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
			wg.Add(2)

			curr_day := time.Time.Day(time.Now())
			if last_day != curr_day {
				go func() {
					defer wg.Done()
					if THY_checker.Flag_checkined {
						return
					}

					// thy
					log.Printf("%s last_day: %d, curr_day: %d\n", kTHY_WHOAMI, last_day, curr_day)
					if _, ok := webs[kTHY_WHOAMI]; ok {
						err := THY_checker.Checkin(kTHY_CHECKIN_HEADER_ACCEPT, kTHY_HEADER_CONTENT_LENGTH, kTHY_URL_ORIGIN)
						if err != nil {
							// return kTHY_WHOAMI, err
							notifySend("Checkiner", "critical", kTHY_WHOAMI+" Check in Failed: "+err.Error())
							THY_checker.Flag_checkined = false
							return
						}
						THY_checker.Flag_checkined = true
					} else {
						log.Printf("%s does not exist\n", kTHY_WHOAMI)
					}
				}()

				go func() {
					defer wg.Done()
					if CUTECLOUD_checker.Flag_checkined {
						return
					}

					// Cutecloud
					log.Printf("%s last_day: %d, curr_day: %d\n", kCUTECLOUD_WHOAMI, last_day, curr_day)
					if _, ok := webs[kCUTECLOUD_WHOAMI]; ok {
						err := CUTECLOUD_checker.Checkin(kCUTECLOUD_CHECKIN_HEADER_ACCEPT, kCUTECLOUD_HEADER_CONTENT_LENGTH, kCUTECLOUD_URL_ORIGIN)
						if err != nil {
							// return kCUTECLOUD_WHOAMI, err
							notifySend("Checkiner", "critical", kCUTECLOUD_WHOAMI+" Check in Failed: "+err.Error())

							CUTECLOUD_checker.Flag_checkined = false
							return
						}
						CUTECLOUD_checker.Flag_checkined = true
					} else {
						log.Printf("%s does not exist\n", kCUTECLOUD_WHOAMI)
					}
				}()

				wg.Wait()
				if THY_checker.Flag_checkined && CUTECLOUD_checker.Flag_checkined {
					last_day = curr_day
					THY_checker.Flag_checkined, CUTECLOUD_checker.Flag_checkined = false, false
				}
			} else {
				log.Printf("last_day: %d, curr_day: %d\n", last_day, curr_day)
				last_day = curr_day
			}
		}
	}
}

func init() {
	flag.BoolVar(&h, "h", false, "help")

	flag.StringVar(&web, "w", "", "set target webs ("+kDELEMITER+" is split char) support: [THY, CUTECLOUD]")
	flag.StringVar(&path, "p", "", "set target webs cookie ("+kDELEMITER+" is split char) support: [THY, CUTECLOUD]")
	flag.Float64Var(&interval, "i", 30, "set checkin interval (minute)")
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

	// Init last day
	kLAST_DAYS = make(map[string]int)
	kLAST_DAYS[kTHY_WHOAMI] = -1
	kLAST_DAYS[kCUTECLOUD_WHOAMI] = -1

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
