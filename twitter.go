package main

import (
	"fmt"
	"net/url"
	"strconv"

	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Masterminds/cookoo/log"
	"github.com/dark-lab/CitizenTopicNetwork/shared/config"
)

type timelinesTweets map[string]anaconda.Tweet
type searchTweets map[string]anaconda.Tweet

type TwitterCrawler struct {
	api           *anaconda.TwitterApi
	configuration *config.Configuration
}

func NewTwitterCrawler(conf *config.Configuration) *TwitterCrawler {
	return &TwitterCrawler{api: GetTwitter(conf), configuration: conf}
}

func GetTwitter(conf *config.Configuration) *anaconda.TwitterApi {

	anaconda.SetConsumerKey(conf.TwitterConsumerKey)
	anaconda.SetConsumerSecret(conf.TwitterConsumerSecret)
	return anaconda.NewTwitterApi(conf.TwitterAccessToken, conf.TwitterAccessTokenSecret)
}

func (c *TwitterCrawler) GetFollowersNumber(account string) int {
	searchresult, _ := c.api.GetUsersShow(account, nil)
	return searchresult.FollowersCount
}

func (c *TwitterCrawler) GetFollowingNumber(account string) int {
	searchresult, _ := c.api.GetUsersShow(account, nil)
	return searchresult.FriendsCount
}

func (c *TwitterCrawler) GetFollowers(account string) []int64 {
	v := url.Values{}
	v.Set("screen_name", account)
	v.Set("count", "200")
	var User anaconda.User
	var Followers []int64
	pages := c.api.GetFollowersListAll(v)
	counter := 0
	for page := range pages {
		//Print the current page of followers
		for _, User = range page.Followers {
			counter++
			Followers = append(Followers, User.Id)
			log.Debug("["+strconv.Itoa(counter)+"] Getting another Follower", User.Id)
		}
	}
	return Followers
}

func (c *TwitterCrawler) GetFollowing(account string) []int64 {
	v := url.Values{}
	v.Set("screen_name", account)
	v.Set("count", "5000")
	var Following []int64
	var id int64
	pages := c.api.GetFriendsIdsAll(v)
	counter := 0
	for page := range pages {
		//Print the current page of "Friends"
		for _, id = range page.Ids {
			counter++
			Following = append(Following, id)
			log.Debug("["+strconv.Itoa(counter)+"] Getting another Following", id)
		}
	}
	return Following
}

func (c *TwitterCrawler) GetTimelines(account string, since int64, strictrange bool) timelinesTweets {
	myTweets := make(timelinesTweets)
	var max_id int64
	var tweet anaconda.Tweet
	searchresult, _ := c.api.GetUsersShow(account, nil)
	v := url.Values{}
	var timeline []anaconda.Tweet
	//var Tweettime string
	v.Set("user_id", searchresult.IdStr)
	v.Set("count", "1") //getting twitter first tweet
	timeline, _ = c.api.GetUserTimeline(v)
	max_id = timeline[0].Id // putting it as max_id
	time, _ := timeline[0].CreatedAtTime()

	for time.Unix() >= since { //until we don't exceed our range of interest

		v = url.Values{}
		v.Set("user_id", searchresult.IdStr)
		v.Set("count", "200")
		v.Set("max_id", strconv.FormatInt(max_id, 10))
		timeline, _ = c.api.GetUserTimeline(v)
		for _, tweet = range timeline {

			time, _ = tweet.CreatedAtTime()
			if strictrange && time.Unix() >= since {
				myTweets[tweet.IdStr] = tweet
			} else {
				myTweets[tweet.IdStr] = tweet
			}
			//Tweettime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second())
			//log.Info("\tTweet @ " + Tweettime + " : " + tweet.IdStr)
			max_id = tweet.Id - 1
		}
		//log.Info("\tFinished reading timeslice for " + account)
	}
	//log.Info("\tFinished reading timeline for " + account)

	return myTweets

}

func (c *TwitterCrawler) Search(since int64, searchString string) searchTweets {
	myTweets := make(searchTweets)
	var max_id int64
	var tweet anaconda.Tweet
	v := url.Values{}
	var Tweettime string
	var myTime time.Time
	v.Set("count", "200")

	searchResult, _ := c.api.GetSearch(searchString, v)
	for _, tweet := range searchResult.Statuses {
		myTweets[tweet.IdStr] = tweet
		fmt.Println(tweet.Text)
		myTime, _ = tweet.CreatedAtTime()

	}
	max_id = searchResult.Metadata.MaxId // putting it as max_id
	for myTime.Unix() >= since {         //until we don't exceed our range of interest
		v = url.Values{}
		v.Set("count", "200")
		v.Set("max_id", strconv.FormatInt(max_id, 10))

		searchResult, _ := c.api.GetSearch(searchString, v)

		for _, tweet = range searchResult.Statuses {
			if myTime.Unix() >= since {
				myTweets[tweet.IdStr] = tweet
			}
			fmt.Println(tweet.Text)
			myTime, _ = tweet.CreatedAtTime()

			Tweettime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", myTime.Year(), myTime.Month(), myTime.Day(), myTime.Hour(), myTime.Minute(), myTime.Second())
			user_tweet := tweet.User.IdStr
			log.Info("[" + user_tweet + "] Tweet @ " + Tweettime + " : " + tweet.IdStr)
			max_id = tweet.Id - 1
		}

	}
	return myTweets

}
