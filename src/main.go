package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func parseWeb(web string, path string) map[string]string {
	webs := strings.Split(web, kDELEMITER)
	paths := strings.Split(path, kDELEMITER)

	webs_map := map[string]string{}
	for idx, w := range webs {
		webs_map[w] = paths[idx]
	}

	return webs_map
}

func CheckinRun(webs map[string]string) (string, error) {
	// Create timer
	timer := time.NewTimer(kINTEVAL)
	defer timer.Stop()

	CUTECLOUD_checker := NewCheckiner(kCUTECLOUD_WHOAMI, kCUTECLOUD_LOGIN_HEADER_ACCEPT, kCUTECLOUD_LOGIN_HEADER_CONTENT_TYPE, kCUTECLOUD_LOGIN_HEADER_METHOD, kCUTECLOUD_URL_LOGIN, kCUTECLOUD_CHECKIN_HEADER_METHOD, kCUTECLOUD_URL_CHECKIN, webs[kCUTECLOUD_WHOAMI])

	THY_checker := NewCheckiner(kTHY_WHOAMI, kTHY_LOGIN_HEADER_ACCEPT, kTHY_LOGIN_HEADER_CONTENT_TYPE, kTHY_LOGIN_HEADER_METHOD, kTHY_URL_LOGIN, kTHY_CHECKIN_HEADER_METHOD, kTHY_URL_CHECKIN, webs[kTHY_WHOAMI])

	for {
		curr_day := time.Time.Day(time.Now())
		select {
		case <-timer.C:
			if last_day != curr_day {
				// FIXME Use channel to communicate, or others will be run in the same time
				// thy
				log.Printf("%s last_day: %d, curr_day: %d\n", kTHY_WHOAMI, last_day, curr_day)
				if _, ok := webs[kTHY_WHOAMI]; ok {
					err := THY_checker.Checkin(kTHY_CHECKIN_HEADER_ACCEPT, kTHY_HEADER_CONTENT_LENGTH, kTHY_URL_ORIGIN)
					if err != nil {
						return kTHY_WHOAMI, err
					}
					last_day = curr_day
				} else {
					log.Printf("%s does not exist\n", kTHY_WHOAMI)
				}

				// Cutecloud
				log.Printf("%s last_day: %d, curr_day: %d\n", kCUTECLOUD_WHOAMI, last_day, curr_day)
				if _, ok := webs[kCUTECLOUD_WHOAMI]; ok {
					err := CUTECLOUD_checker.Checkin(kCUTECLOUD_CHECKIN_HEADER_ACCEPT, kCUTECLOUD_HEADER_CONTENT_LENGTH, kCUTECLOUD_URL_ORIGIN)
					if err != nil {
						return kCUTECLOUD_WHOAMI, err
					}
					last_day = curr_day
				} else {
					log.Printf("%s does not exist\n", kCUTECLOUD_WHOAMI)
				}

				timer.Reset(kINTEVAL)
			} else {
				log.Printf("last_day: %d, curr_day: %d\n", last_day, curr_day)
				timer.Reset(kINTEVAL)
				last_day = curr_day
			}
			// default: // Fix bug: Takes up a lot of CPU
			// Nothing to do
		}
	}
	return "", nil // NOTE It should not be here
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
	webs = parseWeb(web, path)
	kINTEVAL = time.Minute * time.Duration(interval)

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
		who, err := CheckinRun(webs)
		// Checkiner failed
		if err != nil {
			notifySend("Checkiner", "critical", who+" Check in Failed: "+err.Error())
		}
	}
}

func usage() {
	log.Printf(`Checkiner version: checkiner/1.2.1
Usage: checkiner [-h] [-w web]

Example: checkiner -i 120 -w THY@CUTECLOUD -p /home/tianen/go/src/Checkiner/config/THY@/home/tianen/go/src/Checkiner/config/CUTECLOUD

Options:
`)
	flag.PrintDefaults()
}
