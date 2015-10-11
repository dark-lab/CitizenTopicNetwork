package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dark-lab/CitizenTopicNetwork/shared/config"
	"time"
)

type timelinesTweets map[string]anaconda.Tweet
type searchTweets map[string]anaconda.Tweet

func GetTwitter(conf *config.Configuration) *anaconda.TwitterApi {

	anaconda.SetConsumerKey(conf.TwitterConsumerKey)
	anaconda.SetConsumerSecret(conf.TwitterConsumerSecret)
	api := anaconda.NewTwitterApi(conf.TwitterAccessToken, conf.TwitterAccessTokenSecret)
	return api
}

func GetFollowersNumber(api *anaconda.TwitterApi, account string) int {
	searchresult, _ := api.GetUsersShow(account, nil)
	return searchresult.FollowersCount
}

func GetFollowingNumber(api *anaconda.TwitterApi, account string) int {
	searchresult, _ := api.GetUsersShow(account, nil)
	return searchresult.FriendsCount
}

func GetFollowers(api *anaconda.TwitterApi, account string) []int64 {
	v := url.Values{}
	v.Set("screen_name", account)
	v.Set("count", "200")
	var User anaconda.User
	var Followers []int64
	pages := api.GetFollowersListAll(v)
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

func GetFollowing(api *anaconda.TwitterApi, account string) []int64 {
	v := url.Values{}
	v.Set("screen_name", account)
	v.Set("count", "5000")
	var Following []int64
	var id int64
	pages := api.GetFriendsIdsAll(v)
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

func GetTimelines(api *anaconda.TwitterApi, account string, since int64) timelinesTweets {
	myTweets := make(timelinesTweets)
	var max_id int64
	var tweet anaconda.Tweet
	searchresult, _ := api.GetUsersShow(account, nil)
	v := url.Values{}
	var timeline []anaconda.Tweet
	//var Tweettime string
	v.Set("user_id", searchresult.IdStr)
	v.Set("count", "1") //getting twitter first tweet
	timeline, _ = api.GetUserTimeline(v)
	max_id = timeline[0].Id // putting it as max_id
	time, _ := timeline[0].CreatedAtTime()

	for time.Unix() >= since { //until we don't exceed our range of interest

		v = url.Values{}
		v.Set("user_id", searchresult.IdStr)
		v.Set("count", "200")
		v.Set("max_id", strconv.FormatInt(max_id, 10))
		timeline, _ = api.GetUserTimeline(v)
		for _, tweet = range timeline {

			time, _ = tweet.CreatedAtTime()
			myTweets[tweet.IdStr] = tweet
			//Tweettime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second())
			//log.Info("\tTweet @ " + Tweettime + " : " + tweet.IdStr)
			max_id = tweet.Id -1
		}
		//log.Info("\tFinished reading timeslice for " + account)
	}
	//log.Info("\tFinished reading timeline for " + account)

	return myTweets

}

func Search(api *anaconda.TwitterApi, since int64, searchString string) searchTweets {
	myTweets := make(searchTweets)
	var max_id int64
	var tweet anaconda.Tweet
	v := url.Values{}
	var Tweettime string
	var myTime time.Time
	v.Set("count", "200")

	searchResult, _ := api.GetSearch(searchString, v)
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

		searchResult, _ := api.GetSearch(searchString, v)

		for _, tweet = range searchResult.Statuses {
			myTweets[tweet.IdStr] = tweet
			fmt.Println(tweet.Text)
			myTime, _ = tweet.CreatedAtTime()

			Tweettime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", myTime.Year(), myTime.Month(), myTime.Day(), myTime.Hour(), myTime.Minute(), myTime.Second())
			user_tweet := tweet.User.IdStr
			log.Info("["+user_tweet+"] Tweet @ " + Tweettime + " : " + tweet.IdStr)
			max_id = tweet.Id -1
		}

	}
	return myTweets

}
