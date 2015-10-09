package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/dark-lab/CitizenTopicNetwork/shared/config"
	_ "github.com/dark-lab/CitizenTopicNetwork/shared/utils"
	"github.com/gernest/nutz"
	. "github.com/mattn/go-getopt"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("CitizenTopicNetwork")
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortpkg}.%{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

func main() {
	var c int
	var configurationFile string

	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend2Formatter)
	OptErr = 0
	for {
		if c = Getopt("g:c:h"); c == EOF {
			break
		}
		switch c {
		case 'g':
			configurationFile = OptArg
			GenerateData(configurationFile)
		case 'c':
			configurationFile = OptArg
			GatherData(configurationFile)
		case 'h':
			println("usage: " + os.Args[0] + " [ -r ] [-c config.json -h]")
			os.Exit(1)
		}
	}

}

func GenerateData(configurationFile string) {
	if configurationFile == "" {
		panic("I can't work without a configuration file")
	}

	log.Info("Loading config")
	conf, err := config.LoadConfig(configurationFile)
	if err != nil {
		panic(err)
	}
	//api := GetTwitter(&conf)
	db := nutz.NewStorage(configurationFile+".db", 0600, nil)
	mygraph := Graph{Nodes: []Node{}, Links: []Link{}}
	innercount := 0
	nodecount := 0
	group := 0
	for _, account := range conf.TwitterAccounts {
		tweets := db.Get(account, "tweets")
		from := db.Get(account, "from")
		retweets := db.Get(account, "retweets")
		unique_mentions := db.Get(account, "unique_mentions")
		total_mentions := db.Get(account, "total_mentions")
		followers := db.Get(account, "followers")
		following := db.Get(account, "following")
		followers_followed := db.Get(account, "followers_followed")
		mentions_to_followed := db.Get(account, "mentions_to_followed")

		log.Info("Account: " + account)
		log.Info("from: " + string(from.Data))

		log.Info("Tweets: " + string(tweets.Data))

		log.Info("retweets: " + string(retweets.Data))
		log.Info("unique_mentions: " + string(unique_mentions.Data))

		log.Info("total_mentions: " + string(total_mentions.Data))
		log.Info("followers: " + string(followers.Data))
		log.Info("following: " + string(following.Data))
		log.Info("followers_followed: " + string(followers_followed.Data))
		log.Info("mentions_to_followed: " + string(mentions_to_followed.Data))
		// myUniqueMentions := db.GetAll(account, "map_unique_mentions").DataList
		// nUniqueMentions, _ := strconv.Atoi(string(unique_mentions.Data))
		// nMentions_to_followed, _ := strconv.Atoi(string(mentions_to_followed.Data))
		// nTweets, _ := strconv.Atoi(string(tweets.Data))
		// nReTweets, _ := strconv.Atoi(string(retweets.Data))

		mygraph.Nodes = append(mygraph.Nodes, Node{Name: account, Group: group})

		// for k, v := range myUniqueMentions {

		// 	weight, _ := strconv.Atoi(string(v))
		// 	mygraph.Nodes = append(mygraph.Nodes, Node{Name: string(k), Group: group, Thickness: 0.01, Size: 0.01})

		// 	mygraph.Links = append(mygraph.Links, Link{Source: innercount, Target: nodecount, Value: weight})
		// 	innercount++
		// }
		innercount++
		nodecount = innercount
		nodecount++
		group++
	}
	fileJson, _ := json.MarshalIndent(mygraph, "", "  ")
	err = ioutil.WriteFile(configurationFile+".output", fileJson, 0644)
	if err != nil {
		log.Info("WriteFileJson ERROR: " + err.Error())
	}

}

func GatherData(configurationFile string) {
	var retweets int

	if configurationFile == "" {
		panic("I can't work without a configuration file")
	}

	log.Info("Loading config")
	conf, err := config.LoadConfig(configurationFile)
	if err != nil {
		panic(err)
	}

	myTweets := make(map[string]timelinesTweets)
	api := GetTwitter(&conf)

	retweetRegex, _ := regexp.Compile(`^RT`) // detecting retweets

	for _, account := range conf.TwitterAccounts {
		log.Info("-== Timeline for Account: %#v ==-\n", account)

		myTweets[account] = GetTimelines(api, account, conf.FetchFrom)

		log.Info("-== END TIMELINE for %#v ==-\n", account)

	}

	log.Info("Analyzing && collecting data")

	for i := range myTweets {
		retweets = 0
		db := nutz.NewStorage(i+".db", 0600, nil)
		fmt.Println("-== Account: " + i + " ==-")
		fmt.Println("\tTweets: " + strconv.Itoa(len(myTweets[i])))
		var SocialNetwork map[string]struct{}
		SocialNetwork = make(map[string]struct{})
		for _, t := range myTweets[i] {
			// detecting hashtags
			for _, tag := range t.Entities.Hashtags {

				if tag.Text != "" {
					SocialNetwork[tag.Text] = struct{}{}
				}
			}
			if retweetRegex.MatchString(t.Text) == true {
				retweets++
			}
		}

		fmt.Println("\tRetweets " + strconv.Itoa(retweets) + " retweets")
		var memory_network map[string]map[string]int
		memory_network = make(map[string]map[string]int)
		for k, _ := range SocialNetwork {

			db.Create(i, k, []byte(""), "hashtags")
			db.Create(i, "retweets", []byte(strconv.Itoa(retweets)))

			fmt.Println("\tFound hashtag: " + k)
			MyTweetsNetwork := Search(api, conf.FetchFrom, "#"+k)
			for _, tweet := range MyTweetsNetwork {
				if _, exists := memory_network[tweet.User.IdStr]; exists {
					memory_network[tweet.User.IdStr][k]++
				} else {
					memory_network[tweet.User.IdStr] = make(map[string]int)
					memory_network[tweet.User.IdStr][k]++
				}

			}
			//TODO: convert memory_network to be saved in nutz
			//TODO: not searching right before we found an hashtag, storing them to be UNIQUE, then in another phase searching deep further
		}

	}
}

func FloatToString(input_num float32) string {
	// to convert a float number to a string
	return strconv.FormatFloat(float64(input_num), 'f', 6, 64)
}
