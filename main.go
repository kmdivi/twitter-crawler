package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

// APIConf contains api key
type APIConf struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var apiConf APIConf
	{
		apiConfPath := flag.String("conf", "config.json", "API Config File")
		flag.Parse()
		data, errFile := ioutil.ReadFile(*apiConfPath)
		check(errFile)
		errJSON := json.Unmarshal(data, &apiConf)
		check(errJSON)
	}

	anaconda.SetConsumerKey(apiConf.ConsumerKey)
	anaconda.SetConsumerSecret(apiConf.ConsumerSecret)
	api := anaconda.NewTwitterApi(apiConf.AccessToken, apiConf.AccessTokenSecret)

	var track = os.Args[1]

	v := url.Values{}
	v.Set("track", track)
	v.Set("tweet_mode", "extended")

	twitterStream := api.PublicStreamFilter(v)
	for {
		x := <-twitterStream.C
		switch tweet := x.(type) {
		case anaconda.Tweet:
			fmt.Println(tweet.User.CreatedAt)
			fmt.Println(tweet.User.Name)
			fmt.Println("https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr)
			fmt.Printf("\n")
			fmt.Println(tweet.FullText)
			fmt.Println("-----------")
		case anaconda.StatusDeletionNotice:
			// pass
		default:
			fmt.Printf("unknown type(%T) : %v \n", x, x)
		}
	}
}
