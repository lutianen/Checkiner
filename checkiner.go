package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"compress/flate"
	"compress/gzip"

	"github.com/andybalholm/brotli"
)

// Web site
var kTHY_URL string = "https://ssthy.us/user/checkin"
var kCUTECLOUD_RUL = "https://1.cutecloud.net/user/checkin"
var kTHY_COOKIE_FILE = "./THY_cookie.txt"
var kCUTECLOUD_COOKIE_FILE = "./CUTECLOUD_cookie.txt"

// time interval
var kINTEVAL = time.Hour*24 + time.Second*30

var (
	h bool

	web string

	// cookie
	kTHY_COOKIE       string
	kCUTECLOUD_COOKIE string
)

// Display resoponse for JSON
func HandleResponse(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("Read body failed: ", err)
		return err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		fmt.Println("JSON parse failed: ", err)
		return err
	}

	for k, v := range dat {
		fmt.Println(k, ": ", v)
	}
	return nil
}

func THYCheckiner(cookie string) error {
	kHEADERS := map[string]string{
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding":    "gzip, deflate, br",
		"Accept-Language":    "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Content-Length":     "0",
		"Cookie":             cookie,
		"Origin":             "https://ssthy.us",
		"Referer":            "https://ssthy.us/user",
		"Sec-Ch-Ua":          "\"Chromium\";v=\"118\", \"Google Chrome\";v=\"118\", \"Not=A?Brand\";v=\"99\"",
		"Sec-Ch-Ua-Mobile":   "0",
		"Sec-Ch-Ua-Platform": "Linux",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"X-Requested-With":   "XMLHttpRequest",
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", kTHY_URL, nil)
	if err != nil {
		fmt.Println("Error >>> Creating request: ", err)
	}

	// Add header
	for key, value := range kHEADERS {
		req.Header.Set(key, value)
	}

	// Create HTTP client and send requset
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error >>> POST request err: ", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Status Code Error: ", resp.StatusCode)
		return err
	}

	// Debug: response header
	// for k, v := range resp.Header {
	// 	println(k, ":", v[0])
	// }

	// br 压缩
	fmt.Println("\n>>> THY START <<<")
	defer fmt.Println(">>> THY END <<<")

	if resp.Header.Get("Content-Encoding") == "br" {
		reader := brotli.NewReader(resp.Body)
		err := HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("gzip")

		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println("Create gzip reader error: ", err)
			return err
		}
		err = HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "deflate" {
		reader := flate.NewReader(resp.Body)
		err := HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else {
		fmt.Println("Not supported Content-Encoding")
		return err
	}
	return nil
}

func CUTECLOUDCheckiner(cookie string) error {
	kHEADERS := map[string]string{
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding":    "gzip, deflate, br",
		"Accept-Language":    "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7",
		"Content-Length":     "0",
		"Cookie":             cookie,
		"Origin":             "https://1.cutecloud.net",
		"Referer":            "https://1.cutecloud.net/user",
		"Sec-Ch-Ua":          "\"Chromium\";v=\"118\", \"Google Chrome\";v=\"118\", \"Not=A?Brand\";v=\"99\"",
		"Sec-Ch-Ua-Mobile":   "0",
		"Sec-Ch-Ua-Platform": "Linux",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"X-Requested-With":   "XMLHttpRequest",
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", kCUTECLOUD_RUL, nil)
	if err != nil {
		fmt.Println("Error >>> Creating request: ", err)
		return err
	}

	// Add header
	for key, value := range kHEADERS {
		req.Header.Set(key, value)
	}

	// Create HTTP client and send requset
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error >>> POST request err: ", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Status Code Error: ", resp.StatusCode)
		return err
	}

	// Debug: response header
	// for k, v := range resp.Header {
	// 	println(k, ":", v[0])
	// }

	// br 压缩
	fmt.Println("\n>>> CuteCloud START <<<")
	defer fmt.Println(">>> CuteCloud END <<<")
	if resp.Header.Get("Content-Encoding") == "br" {
		reader := brotli.NewReader(resp.Body)
		err := HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("gzip")
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println("Create gzip reader error: ", err)
			return err
		}
		err = HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "deflate" {
		reader := flate.NewReader(resp.Body)
		err := HandleResponse(reader)
		if err != nil {
			fmt.Println("Handle response failed: ", err)
			return err
		}
	} else {
		fmt.Println("Not supported Content-Encoding")
		return err
	}
	return err
}

func parseWeb(web string) map[string]string {
	webs := strings.Split(web, "&")
	webs_map := map[string]string{}
	for _, w := range webs {
		webs_map[w] = w
	}
	return webs_map
}

func Checkiner(web string) error {
	webs := parseWeb(web)
	// Create timer
	THY_timer := time.NewTimer(kINTEVAL)
	CUTECLOUD_timer := time.NewTimer(kINTEVAL)

	for {
		select {

		case <-THY_timer.C:
			if _, ok := webs["THY"]; ok {
				err := THYCheckiner(kTHY_COOKIE)
				if err != nil {
					return err
				}
				THY_timer.Reset(kINTEVAL)
			}
		case <-CUTECLOUD_timer.C:
			if _, ok := webs["CUTECLOUD"]; ok {
				err := CUTECLOUDCheckiner(kCUTECLOUD_COOKIE)
				if err != nil {
					return err
				}
				CUTECLOUD_timer.Reset(kINTEVAL)
			}
		default:
			// Nothing to do
		}
	}
}

// Get cookie from file
func readCookie(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Read cookie file failed: ", err)
		return ""
	}
	return string(buf)
}

func init() {
	flag.BoolVar(&h, "h", false, "help")

	flag.StringVar(&web, "w", "THY", "set target webs (& is split char) support: [THY, CUTECLOUD]")

	flag.Usage = usage

	kTHY_COOKIE = readCookie(kTHY_COOKIE_FILE)
	kCUTECLOUD_COOKIE = readCookie(kCUTECLOUD_COOKIE_FILE)
}

func main() {
	flag.Parse()

	if h || web == "" || (kTHY_COOKIE == "" && kCUTECLOUD_COOKIE == "") {
		flag.Usage()
		return
	}

	err := Checkiner(web)
	if err != nil {
		exec.Command("notify-send", "-u", "critical", "Checkiner", "Checkiner failed").Run()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Checkiner version: checkiner/0.0.1
Usage: nginx [-h] [-w web]

Options:
`)
	flag.PrintDefaults()
}
