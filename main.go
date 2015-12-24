package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/dark-lab/CitizenTopicNetwork/shared/config"
	"github.com/dark-lab/CitizenTopicNetwork/shared/utils"
	"github.com/gernest/nutz"
	. "github.com/mattn/go-getopt"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("CitizenTopicNetwork")
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortpkg}.%{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

const GENERATED_HASHTAGS = "user_generated_hashtags"
const MATCHING_HASHTAGS = "matching_hashtags"

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
	fmt.Println(">> Exporting gathered data")
	mygraph := Graph{Nodes: []Node{}, Links: []Link{}, Mode: "static", Defaultedgetype: "undirected"}

	for _, account := range conf.TwitterAccounts {
		fmt.Println(">> Generating graph for " + account)
		mygraph = Graph{Nodes: []Node{}, Links: []Link{}, Mode: "static", Defaultedgetype: "undirected"}
		db := nutz.NewStorage(account+".db", 0600, nil)

		nodecount := 0

		myNetwork := db.GetAll(account, MATCHING_HASHTAGS).DataList
		myMatrix := make(map[string][]int)                  // this is the Matrix Hashtags/ Users ID
		myNetworkMatrix := make(map[string]map[string]int8) //so we can extract later data easily
		myMapNetwork := make(map[int]string)                //this will be used to resolve User ID of the graph <-> Twitter id
		var myCSV [][]string
		HashtagsMap := db.GetAll(account, GENERATED_HASHTAGS)
		var Hashtags []string
		Hashtags = append(Hashtags, "UserID") //First column reserved to userid

		for h, _ := range HashtagsMap.DataList {
			Hashtags = append(Hashtags, string(h))
		}
		myCSV = append(myCSV, Hashtags)

		for k, _ := range myNetwork {
			//		ki64, _ := strconv.ParseInt(string(k), 10, 64)
			ki64 := string(k)
			//Column name is ki64
			myUserNetwork := db.GetAll(account, MATCHING_HASHTAGS, k).DataList
			var userOccurrence []string
			userOccurrence = append(userOccurrence, string(k)) //this is the userid
			for _, h := range Hashtags {
				if occurence, ok := myUserNetwork[h]; ok {
					userOccurrence = append(userOccurrence, string(occurence))
				} else {
					userOccurrence = append(userOccurrence, strconv.Itoa(0))
				}
			}
			myCSV = append(myCSV, userOccurrence)

			myNetworkMatrix[ki64] = make(map[string]int8)
			//fmt.Println("User ID is: " + string(k))
			if len(myUserNetwork) > conf.HashtagCutOff { //Cutting off people that just tweeted 1 hashtag

				htagsatisfied := false //cutoff var
				for h, o := range myUserNetwork {
					occur, _ := strconv.Atoi(string(o))
					//fmt.Println("Hastag: " + string(h) + " saw  " + string(o) + " times")
					if occur > conf.HashtagOccurrenceCutOff { // Cutting off people that just tweeted the hashtag once
						myMatrix[string(h)] = append(myMatrix[string(h)], nodecount) // Generating adjacient map
						htagsatisfied = true                                         //cutoff var, setting true to enable node counting
					}
					occurrences, _ := strconv.Atoi(string(o))
					myNetworkMatrix[ki64][string(h)] = int8(occurrences) // convert the db to a map

				}
				if htagsatisfied { //Cutting off also nodes who satisfied the cuttoff above
					myMapNetwork[nodecount] = ki64 //mapping Graph user id with Tweet user id

					mygraph.Nodes = append(mygraph.Nodes, Node{Id: nodecount, Name: string(k), Group: 1})
					nodecount++
				}
			}

		}
		fmt.Println(">> Preparing graph for " + account)
		linkscount := 0
		for hashtag, users := range myMatrix {
			for _, userid := range users {
				for _, userid2 := range users {
					if userid2 != userid {
						if int(myNetworkMatrix[myMapNetwork[userid]][hashtag]) > conf.HashtagOccurrenceCutOff {

							mygraph.Links = append(mygraph.Links, Link{Id: linkscount, Source: userid, Target: userid2, Value: float32(myNetworkMatrix[myMapNetwork[userid]][hashtag])})

							linkscount++
						}
					}
				}
			}
		}
		fmt.Println(">> Writing matrix to csv")
		utils.WriteCSV(myCSV, account+".csv")
		fmt.Println(">> Writing graph to json file")

		//
		// nUniqueMentions, _ := strconv.Atoi(string(unique_mentions.Data))
		// nMentions_to_followed, _ := strconv.Atoi(string(mentions_to_followed.Data))
		// nTweets, _ := strconv.Atoi(string(tweets.Data))
		// nReTweets, _ := strconv.Atoi(string(retweets.Data))

		//mygraph.Nodes = append(mygraph.Nodes, Node{Name: account, Group: group})

		// for k, v := range myUniqueMentions {

		// 	weight, _ := strconv.Atoi(string(v))
		// 	mygraph.Nodes = append(mygraph.Nodes, Node{Name: string(k), Group: group, Thickness: 0.01, Size: 0.01})

		// 	mygraph.Links = append(mygraph.Links, Link{Source: innercount, Target: nodecount, Value: weight})
		// 	innercount++
		// }

		fileJson, _ := ffjson.Marshal(&mygraph)
		err = ioutil.WriteFile(account+".output", fileJson, 0644)
		if err != nil {
			log.Info("WriteFileJson ERROR: " + err.Error())
		}
		out, err := xml.Marshal(mygraph)
		out = append([]byte(`<?xml version="1.0" encoding="UTF-8"?><gexf xmlns="http://www.gexf.net/1.2draft" version="1.2"> <meta lastmodifieddate="2009-03-20"> <creator>dark-lab</creator><description>Gephi file</description> </meta>`), out...)
		out = append(out, []byte(`</gexf>`)...)

		err = ioutil.WriteFile(account+".output.gexf", out, 0644)
		if err != nil {
			log.Info("WriteFileJson ERROR: " + err.Error())
		}
	}

}

func GatherData(configurationFile string) {

	if configurationFile == "" {
		panic("I can't work without a configuration file")
	}

	log.Info("Loading config")
	conf, err := config.LoadConfig(configurationFile)
	if err != nil {
		panic(err)
	}
	crawler := NewTwitterCrawler(&conf)

	myTweets := make(map[string]timelinesTweets)

	for _, account := range conf.TwitterAccounts {
		log.Info("-== Timeline for Account: %#v ==-\n", account)

		if crawler.configuration.Number != 0 {
			myTweets[account] = crawler.GetTimelinesN(account, false, conf.Number, conf.Slices) //false: don't be strict, getting all hashtag in the timeline, also if they are out the interested range

		} else {
			myTweets[account] = crawler.GetTimelines(account, false) //false: don't be strict, getting all hashtag in the timeline, also if they are out the interested range
		}
		log.Info("-== END TIMELINE for %#v ==-\n", account)

	}

	for _, account := range conf.TwitterAccounts {

		GatherDataFromAccount(crawler, account, myTweets[account])
	}

}

func GatherDataFromAccount(crawler *TwitterCrawler, account string, timeLine timelinesTweets) {
	retweetRegex, _ := regexp.Compile(`^RT`) // detecting retweets
	log.Info(">> Depth look on " + account)

	retweets := 0
	db := nutz.NewStorage(account+".db", 0600, nil)
	fmt.Println("-== Account: " + account + " ==-")
	fmt.Println("\tTweets: " + strconv.Itoa(len(timeLine)))
	var SocialNetwork map[string]struct{}
	SocialNetwork = make(map[string]struct{})
	for _, t := range timeLine {
		// detecting hashtags
		for _, tag := range t.Entities.Hashtags {

			if tag.Text != "" {
				fmt.Println("\tFound hashtag: " + tag.Text)
				SocialNetwork[tag.Text] = struct{}{}
			}
		}
		if retweetRegex.MatchString(t.Text) == true {
			retweets++
		}
	}

	fmt.Println("\tRetweets " + strconv.Itoa(retweets) + " retweets")
	fmt.Println("\t" + strconv.Itoa(len(SocialNetwork)) + " hashtags")

	var memory_network map[string]map[string]int
	memory_network = make(map[string]map[string]int)

	//Cycling on hashtags
	for k, _ := range SocialNetwork {

		db.Create(account, k, []byte(""), GENERATED_HASHTAGS)
		db.Create(account, "retweets", []byte(strconv.Itoa(retweets)))

		fmt.Println("\t Searching hashtag: " + k)
		var MyTweetsNetwork searchTweets

		// not searching right before we found an hashtag
		// storing them to be UNIQUE, then in another phase searching deep further
		if crawler.configuration.Number != 0 {
			MyTweetsNetwork = crawler.SearchN("#"+k, crawler.configuration.Number, crawler.configuration.Slices)
		} else {
			MyTweetsNetwork = crawler.Search("#" + k)
		}
		for _, tweet := range MyTweetsNetwork {
			if _, exists := memory_network[tweet.User.IdStr]; exists {
				memory_network[tweet.User.ScreenName][k]++
			} else {
				memory_network[tweet.User.ScreenName] = make(map[string]int)
				memory_network[tweet.User.ScreenName][k]++
			}

		}
	}

	for user, tags := range memory_network {
		for tag, occurrence := range tags {
			db.Create(account, tag, []byte(
				strconv.Itoa(occurrence)), MATCHING_HASHTAGS, user)
		}
	}

}

func FloatToString(input_num float32) string {
	// to convert a float number to a string
	return strconv.FormatFloat(float64(input_num), 'f', 6, 64)
}
