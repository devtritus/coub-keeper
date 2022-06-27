package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func main() {
	getPage()
	fmt.Println("Keeper was closed")
}

func getPage() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("Got error while creating cookie jar")
	}

	client := &http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest(http.MethodGet, "https://coub.com/api/v2/timeline/likes", nil)
	if err != nil {
		log.Fatal(err)
	}

	cookie := &http.Cookie{
		Name:  "remember_token",
		Value: "7a17afe7de09e5aa0c209921745a9dedff5724f2",
	}

	urlObj, _ := url.Parse("https://coub.com/")

	client.Jar.SetCookies(urlObj, []*http.Cookie{cookie})

	q := req.URL.Query()

	q.Add("all", "true")
	q.Add("order_by", "date")
	q.Add("page", "1")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(responseBody))
	}

	fmt.Println(resp.Status)

}
