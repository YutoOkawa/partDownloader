package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func headResponse() int64 {
	resp, err := http.Head("http://cdimage-u-toyama.ubuntulinux.jp/releases/18.04.3/ubuntu-ja-18.04.3-desktop-amd64.iso")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.ContentLength
}

func respBodyBytes(resp *http.Response) ([]byte, error) {
	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

func convertInt64ToString(i int64) string {
	convertedInt64 := strconv.FormatInt(i, 10)
	return convertedInt64
}

func getResponse(client *http.Client, start int64, end int64, respCh chan []byte) <-chan []byte {
	req, err := http.NewRequest("GET", "http://cdimage-u-toyama.ubuntulinux.jp/releases/18.04.3/ubuntu-ja-18.04.3-desktop-amd64.iso", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Range", "bytes="+convertInt64ToString(start)+"-"+convertInt64ToString(end))

	// リクエスト送信
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// レスポンスボディをバイト出力
	body, err := respBodyBytes(resp)
	if err != nil {
		log.Fatal(err)
	}
	respCh <- body
	return respCh
}

func main() {
	// var contents []byte
	// Content-Lengthの取得
	contentLength := headResponse()
	fmt.Println(contentLength)
	ranges := contentLength / http.DefaultMaxIdleConnsPerHost
	fmt.Println(ranges)

	client := http.DefaultClient
	respCh1 := make(chan []byte)
	respCh2 := make(chan []byte)
	go getResponse(client, 0, ranges, respCh1)
	go getResponse(client, ranges+1, ranges*http.DefaultMaxIdleConnsPerHost-1, respCh2)

	var checkCh1 bool = false
	var checkCh2 bool = false

	var contentCh1 []byte
	var contentCh2 []byte

	for !(checkCh1 && checkCh2) {
		select {
		case contentCh1 = <-respCh1:
			fmt.Println("1")
			checkCh1 = true
		case contentCh2 = <-respCh2:
			fmt.Println("2")
			checkCh2 = true
		}
	}

	file, err := os.OpenFile("ubuntu-ja-18.04.3-desktop-amd64_my.iso", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Write(contentCh1)
	file.Write(contentCh2)
}
