package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	kHEADERS = map[string]string{
		"Accept-Encoding":    "gzip, deflate, br",
		"Accept-Language":    "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Sec-Ch-Ua":          `"Chromium";v="118", "Google Chrome";v="118", "Not=A?Brand";v="99"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": "Linux",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"X-Requested-With":   "XMLHttpRequest",
	}

	// Web site

	// https://portal.ssthy.us/
	// kTHY_URL_ORIGIN  string = "https://ssthy.us"
	kTHY_URL_ORIGIN  string = "https://portal.ssthy.us"
	kTHY_URL_LOGIN   string = "https://portal.ssthy.us/auth/login"
	kTHY_URL_CHECKIN string = "https://portal.ssthy.us/user/checkin"

	kCUTECLOUD_URL_ORIGIN  string = "https://www.cute-cloud.top"
	kCUTECLOUD_URL_LOGIN   string = "https://www.cute-cloud.top/auth/login"
	kCUTECLOUD_URL_CHECKIN string = "https://www.cute-cloud.top/user/checkin"

	CHECKIN_HEADER_ACCEPT     string = "application/json, text/javascript, */*; q=0.01"
	LOGIN_HEADER_ACCEPT       string = "application/json, text/javascript, */*; q=0.01"
	LOGIN_HEADER_CONTENT_TYPE string = "application/x-www-form-urlencoded; charset=UTF-8"
	LOGIN_HEADER_METHOD       string = "POST"
	CHECKIN_HEADER_METHOD     string = "POST"
	HEADER_CONTENT_LENGTH     string = "0"

	kCUTECLOUD_LOGIN_HEADER_ACCEPT string = "*/*;"

	kDELEMITER string = "@"

	//>>>> flags
	h bool
	// THY
	web string
	// /home/username/...
	path string
	// web : path
	webs map[string]string

	// time interval
	kINTEVAL time.Duration = time.Minute
	interval float64

	// The log file path
	kLOG_FILE string = "./checkiner.log"
	// flags <<<<
)

/**
 * TAG Read config file
 *	email: first line
 *	passwd: second line
 */
func readConfigFromFile(path string) (string, string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		// fmt.Println("Read config file error: ", err)
		log.Println("Read config file error: ", err)
		return "", "", err
	}
	buf_str := strings.Split(string(buf), "\n")
	email, passwd := buf_str[0], buf_str[1]

	return email, passwd, nil
}

func notifySend(title string, level string, body string) {
	exec.Command("notify-send", "-u", level, title, body).Run()
	log.Println(title, body)
}
