package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	rethink "github.com/dancannon/gorethink"
	"github.com/gorilla/feeds"
	"io"
	"log"
	"net/http"
	"time"
)

var session *rethink.Session

type Message struct {
	Sender  string
	Content string
	Channel string
	Url     string
	Posted  time.Time
}

func ChannelRSS(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get(":channel")

	feed := &feeds.Feed{
		Title:       "Recent Gifs for " + channel,
		Link:        &feeds.Link{Href: "http://gifs.boner.io/"},
		Description: "COOL GIFS COOL GIFS COOL GIFS",
		Author:      &feeds.Author{"elgruntox", "mumphster@gmail.com"},
		Created:     time.Now(),
	}

	items := []*feeds.Item{}

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Channel").Eq("#" + channel)).OrderBy("Date").Limit(20)
	rows, _ := query.Run(session)

	for rows.Next() {
		var m Message
		err := rows.Scan(&m)
		if err != nil {
			fmt.Println(err)
			return
		}
		item := &feeds.Item{
			Title:       m.Sender + " posted a new gif in " + m.Channel,
			Link:        &feeds.Link{Href: m.Url},
			Description: m.Url,
			Author:      &feeds.Author{m.Sender, "bill.gates@microsoft.com"},
			Created:     time.Now(),
		}

		items = append(items, item)
	}

	feed.Items = items

	atom, _ := feed.ToAtom()

	w.Header().Set("Content-Type", "application/atom+xml")
	io.WriteString(w, atom)
}

func main() {
	session = InitDB()

	m := pat.New()

	m.Get("/feed/:channel.atom", http.HandlerFunc(ChannelRSS))

	http.Handle("/", m)

	err := http.ListenAndServe("127.0.0.1:9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func InitDB() *rethink.Session {
	session, err := rethink.Connect(rethink.ConnectOpts{
		Address:     "localhost:28015",
		Database:    "gifs",
		MaxIdle:     10,
		IdleTimeout: time.Second * 10,
	})

	if err != nil {
		log.Println(err)
	}

	err = rethink.DbCreate("gifs").Exec(session)
	if err != nil {
		log.Println(err)
	}

	_, err = rethink.Db("gifs").TableCreate("entries").RunWrite(session)
	if err != nil {
		log.Println(err)
	}

	return session
}
