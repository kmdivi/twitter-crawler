package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"

	"github.com/ChimeraCoder/anaconda"
)

var reMD5 = regexp.MustCompile("[a-fA-F0-9]{32}")
var reSHA1 = regexp.MustCompile("[a-fA-F0-9]{40}")
var reSHA256 = regexp.MustCompile("[a-fA-F0-9]{64}")

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

// CheckHash gets hash value from tweet.
func CheckHash(tweettext string) [][]string {
	if reSHA256.MatchString(tweettext) == true {
		hashSHA256 := reSHA256.FindAllStringSubmatch(tweettext, -1)
		return hashSHA256
	} else if reSHA1.MatchString(tweettext) == true {
		hashSHA1 := reSHA1.FindAllStringSubmatch(tweettext, -1)
		return hashSHA1
	} else if reMD5.MatchString(tweettext) == true {
		hashMD5 := reMD5.FindAllStringSubmatch(tweettext, -1)
		return hashMD5
	}
	return nil
}

func writeToCSV(hash []string) error {
	file, err := os.OpenFile("hashlist.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, v := range hash {
		_, err = file.WriteString(v + "\n")
	}

	if err != nil {
		return err
	}
	return nil
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

	v := url.Values{}
	v.Set("tweet_mode", "extended")
	for _, a := range os.Args {
		v.Set("track", a)
	}

	twitterStream := api.PublicStreamFilter(v)
	for {
		x := <-twitterStream.C
		switch tweet := x.(type) {
		case anaconda.Tweet:
			fmt.Println(tweet.CreatedAt)
			fmt.Println(tweet.User.Name)
			fmt.Println(tweet.Text)
			if tweet.RetweetedStatus != nil {
				fmt.Println("***")
				fmt.Println(tweet.RetweetedStatus.Text)
				fmt.Println("***")
			}
			fmt.Println("https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr)
			fmt.Printf("\n")
			fmt.Println(tweet.FullText)
			fmt.Println("-----------")

			hash := CheckHash(tweet.FullText)
			if hash != nil {
				fmt.Println(hash)
				for _, v := range hash {
					err := writeToCSV(v)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		case anaconda.StatusDeletionNotice:
			// pass
		default:
			fmt.Printf("unknown type(%T) : %v \n", x, x)
		}
	}
}
