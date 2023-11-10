package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

type Checkiner_t struct {
	Whoami string

	Login_header_accpet       string
	Login_header_content_type string
	Login_header_method       string
	Login_url                 string

	Checkin_header_method string
	Checkin_url           string

	// private
	email  string
	passwd string
}

func NewCheckiner(whoami string, login_header_accpet string, login_header_content_type string, login_header_method string, login_url string, checkin_header_method string, checkin_url string, config_file_path string) *Checkiner_t {
	email, passwd, err := readConfigFromFile(config_file_path)
	if err != nil {
		// fmt.Println("Read config file error: ", err)
		log.Fatal("Read config file error: ", err)
		return nil
	}

	return &Checkiner_t{
		Whoami:                    whoami,
		Login_header_accpet:       login_header_accpet,
		Login_header_content_type: login_header_content_type,
		Login_header_method:       login_header_method,
		Login_url:                 login_url,

		Checkin_header_method: checkin_header_method,
		Checkin_url:           checkin_url,

		// TAG  Set email and passwd by reading config file
		email:  email,
		passwd: passwd,
	}
}

func (this *Checkiner_t) setRequestHeader(req *http.Request) {
	header := map[string]string{
		"Accept":             this.Login_header_accpet,
		"Content-Type":       this.Login_header_content_type,
		"Referer":            this.Login_url,
		"Sec-Ch-Ua":          kHEADERS["Sec-Ch-Ua"],
		"Sec-Ch-Ua-Mobile":   kHEADERS["Sec-Ch-Ua-Mobile"],
		"Sec-Ch-Ua-Platform": kHEADERS["Sec-Ch-Ua-Platform"],
		"User-Agent":         kHEADERS["User-Agent"],
		"X-Requested-With":   kHEADERS["X-Requested-With"],
	}
	// Add header
	for key, value := range header {
		req.Header.Set(key, value)
	}
}

func (this *Checkiner_t) setRequestBody(req *http.Request) {
	data := []byte("email=" + this.email + "&passwd=" + this.passwd)
	req.Body = io.NopCloser(bytes.NewBuffer(data))
}

func (this *Checkiner_t) handleLoginResponse(resp *http.Response, cookie *string) error {
	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Status Code Error: ", resp.StatusCode)
		return errors.New("Status Code: " + string(rune(resp.StatusCode)))
	}

	// Handle response body
	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// fmt.Println(string(buffer))
	buf := map[string]any{}
	json.Unmarshal(buffer, &buf)
	for k, v := range buf {
		if k == "ret" {
			fmt.Println(k, ":", v.(float64))
		} else if k == "msg" {
			// fmt.Println(k, ":", v.(string))
			notifySend("Checkiner", "normal", ">>> "+this.Whoami+" "+v.(string))
		} else {
			// fmt.Println("Unknown key: ", k)
			notifySend("Checkiner", "critical", "Unknown key: "+k)
		}
	}

	// TAG Get the lastest cookie
	for k, v := range resp.Header {
		// fmt.Println(k, ":", v[0])
		if k == "Set-Cookie" {
			for _, val := range v {
				// fmt.Println(val)
				str := strings.Split(val, ";")
				*cookie += (str[0] + "; ")
			}
		}
	}
	return nil
}

// Display resoponse for JSON
func (this *Checkiner_t) handleResponse(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		// fmt.Println("Read body failed: ", err)
		log.Println("Read body failed: ", err)
		return err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		// fmt.Println("JSON parse failed: ", err)
		log.Println("JSON parse failed: ", err)
		return err
	}

	for k, v := range dat {
		fmt.Println(k, ": ", v)
		// log.Println(k, ": ", v)
	}
	notifySend("Checkiner", "normal", ">>> "+this.Whoami+" checkin success: "+dat["msg"].(string))
	return nil
}

func (this *Checkiner_t) login() (string, error) {
	cookie := ""

	// Create request
	req, err := http.NewRequest(this.Login_header_method, this.Login_url, nil)
	if err != nil {
		// fmt.Println(">>> "+this.Whoami+" Creating request: ", err)
		log.Println(">>> "+this.Whoami+" Creating request failed: ", err)
		return cookie, err
	}
	this.setRequestHeader(req)
	this.setRequestBody(req)

	// Create HTTP client and send requset
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// fmt.Println(">>> "+this.Whoami+" "+this.Login_header_method+" request err: ", err)
		log.Println(">>> "+this.Whoami+" "+this.Login_header_method+" request err: ", err)
		return cookie, err
	}
	defer resp.Body.Close()
	err = this.handleLoginResponse(resp, &cookie)

	if err != nil {
		return cookie, err
	}

	return cookie, nil
}

func (this *Checkiner_t) Checkin(header_accpet string, header_content_length string, url_orign string) error {
	cookie, err := this.login()
	if err != nil {
		// fmt.Println(">>> "+this.Whoami+" Login error: ", err)
		log.Println(">>> "+this.Whoami+" Login error: ", err)
		return err
	}

	header := map[string]string{
		"Accept":             header_accpet,
		"Accept-Encoding":    kHEADERS["Accept-Encoding"],
		"Accept-Language":    kHEADERS["Accept-Language"],
		"Content-Length":     header_content_length,
		"Cookie":             cookie,
		"Origin":             url_orign,
		"Referer":            this.Checkin_url,
		"Sec-Ch-Ua":          kHEADERS["Sec-Ch-Ua"],
		"Sec-Ch-Ua-Mobile":   kHEADERS["Sec-Ch-Ua-Mobile"],
		"Sec-Ch-Ua-Platform": kHEADERS["Sec-Ch-Ua-Platform"],
		"Sec-Fetch-Dest":     kHEADERS["Sec-Fetch-Dest"],
		"Sec-Fetch-Mode":     kHEADERS["Sec-Fetch-Mode"],
		"Sec-Fetch-Site":     kHEADERS["Sec-Fetch-Site"],
		"User-Agent":         kHEADERS["User-Agent"],
		"X-Requested-With":   kHEADERS["X-Requested-With"],
	}

	// Create HTTP request
	req, err := http.NewRequest(this.Checkin_header_method, this.Checkin_url, nil)
	if err != nil {
		// fmt.Println(">>> "+this.Whoami+" Creating request: ", err)
		log.Println(">>> "+this.Whoami+" Creating request failed: ", err)
		return err
	}

	// Add header
	for key, value := range header {
		req.Header.Set(key, value)
	}

	// Create HTTP client and send requset
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// fmt.Println(">>> "+this.Whoami+" POST request err: ", err)
		log.Println(">>> "+this.Whoami+" POST request err: ", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// fmt.Println(">>> "+this.Whoami+" Status Code Error: ", resp.StatusCode)
		log.Println(">>> "+this.Whoami+" Status Code Error: ", resp.StatusCode)
		return err
	}

	// Debug: response header
	// for k, v := range resp.Header {
	// 	println(k, ":", v[0])
	// }

	// br 压缩
	// Cookie Expired
	if resp.Header.Get("Content-Type") == "text/html; charset=UTF-8" {
		return errors.New("cookie Expired")
	}

	if resp.Header.Get("Content-Encoding") == "br" {
		reader := brotli.NewReader(resp.Body)
		err := this.handleResponse(reader)
		if err != nil {
			// fmt.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			log.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		// fmt.Println("gzip")
		log.Println("gzip")
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			// fmt.Println(">>> "+this.Whoami+" Create gzip reader error: ", err)
			log.Println(">>> "+this.Whoami+" Create gzip reader error: ", err)
			return err
		}
		err = this.handleResponse(reader)
		if err != nil {
			// fmt.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			log.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			return err
		}
	} else if resp.Header.Get("Content-Encoding") == "deflate" {
		reader := flate.NewReader(resp.Body)
		err := this.handleResponse(reader)
		if err != nil {
			// fmt.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			log.Println(">>> "+this.Whoami+" Handle response failed: ", err)
			return err
		}
	} else {
		// fmt.Println("Not supported Content-Encoding")
		log.Println("Not supported Content-Encoding")
		return err
	}
	return nil
}
