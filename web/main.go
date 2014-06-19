package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	rethink "github.com/dancannon/gorethink"
	"github.com/ell/gifbot/helpers"
	"github.com/gorilla/feeds"
	"io"
	"log"
	"net/http"
	"time"
)

var session *rethink.Session

func ChannelRSS(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get(":channel")

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Channel").Eq("#" + channel)).OrderBy("-Date").Limit(20)
	rows, _ := query.Run(session)

	feed := CreateFeed(rows, req)

	atom, _ := feed.ToAtom()

	w.Header().Set("Content-Type", "application/atom+xml")
	io.WriteString(w, atom)
}

func CreateFeed(rows *rethink.ResultRows, req *http.Request) *feeds.Feed {
	channel := req.URL.Query().Get(":channel")
	items := []*feeds.Item{}

	feed := &feeds.Feed{
		Title:       "Recent Gifs for " + channel,
		Link:        &feeds.Link{Href: "http://gifs.boner.io/"},
		Description: "COOL GIFS COOL GIFS COOL GIFS",
		Created:     time.Now(),
	}

	for rows.Next() {
		var m helpers.Message
		err := rows.Scan(&m)
		if err != nil {
			fmt.Println(err)
		}
		item := &feeds.Item{
			Title:       m.Sender + " posted a new gif in " + m.Channel,
			Link:        &feeds.Link{Href: m.Url},
			Description: m.Url,
			Created:     time.Now(),
		}

		items = append(items, item)
	}

	feed.Items = items

	return feed
}

func main() {
	session = helpers.InitDB()

	m := pat.New()

	m.Get("/feed/channel/:channel.atom", http.HandlerFunc(ChannelRSS))

	http.Handle("/", m)

	err := http.ListenAndServe("127.0.0.1:9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
