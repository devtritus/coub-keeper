package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"context"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	credential := options.Credential{
		Username: "user",
		Password: "pass",
	}
	opts := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(credential)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("vault").Collection("coubs")

	page, totalPages, data := getPage(1)

	for i := page; i < totalPages; i++ {
		page, totalPages, data = getPage(i)
		fmt.Println(i, page, totalPages)

		var batch []mongo.WriteModel

		for _, v := range data {
			entry1 := v.(map[string]interface{})
			entryId := entry1["id"].(float64)
			entry := mongo.NewUpdateOneModel().SetFilter(bson.D{{"id", entryId}}).SetUpdate(bson.D{{"$set", entry1}}).SetUpsert(true)

			batch = append(batch, entry)
		}

		opts := options.BulkWrite().SetOrdered(false)
		res, err := collection.BulkWrite(context.TODO(), batch, opts)
		//filter := bson.D{{"address.market", "Sydney"}}
		//update := bson.D{{"$mul", bson.D{{"price", 1.15}}}}
		//res, err := collection.InsertMany(context.TODO(), data)
		if err != nil {
			log.Fatal("can't insert", err)
		} else {
			fmt.Println(res)
		}
	}
}

func getPage(nextPage int) (page, totalPages int, data []interface{}) {
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
		Value: "7a17afe7de09e5aa0c209921745a9dedff5734f2",
	}

	urlObj, _ := url.Parse("https://coub.com/")

	client.Jar.SetCookies(urlObj, []*http.Cookie{cookie})

	q := req.URL.Query()

	q.Add("all", "true")
	q.Add("order_by", "date")
	q.Add("page", strconv.Itoa(nextPage))

	req.URL.RawQuery = q.Encode()

	fmt.Println(req)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)

		return -1, -1, nil
	} else {
		//fmt.Println(string(content))
		return parseBody(content)
	}
}

func parseBody(content []byte) (page, totalPages int, coubs []interface{}) {
	var data map[string]interface{}
	json.Unmarshal(content, &data)

	page = int(data["page"].(float64))
	totalPages = int(data["total_pages"].(float64))
	coubs = data["coubs"].([]interface{})

	return
}
