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
)

var session *rethink.Session

func ChannelRSS(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get(":channel")

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Channel").Eq("#" + channel)).OrderBy("-Date").Limit(20)
	rows, _ := query.Run(session)

	feed := CreateFeed(rows, channel)

	atom, _ := feed.ToAtom()

	w.Header().Set("Content-Type", "application/atom+xml")
	io.WriteString(w, atom)
}

func UserRSS(w http.ResponseWriter, req *http.Request) {
	user := req.URL.Query().Get(":user")

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Sender").Eq(user)).OrderBy("-Date").Limit(20)
	rows, _ := query.Run(session)
	feed := CreateFeed(rows, user)

	atom, _ := feed.ToAtom()

	w.Header().Set("Content-Type", "application/atom+xml")
	io.WriteString(w, atom)
}

func CreateFeed(rows *rethink.ResultRows, subject string) *feeds.Feed {
	items := []*feeds.Item{}

	feed := &feeds.Feed{
		Title:       "Recent " + subject + " Gifs",
		Link:        &feeds.Link{Href: "http://gifs.boner.io/"},
		Description: "COOL GIFS COOL GIFS COOL GIFS",
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
			Created:     m.Posted,
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
	m.Get("/feed/user/:user.atom", http.HandlerFunc(UserRSS))

	http.Handle("/", m)

	err := http.ListenAndServe("127.0.0.1:9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
