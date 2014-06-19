package main

import (
	"encoding/json"
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

type User struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func GifCount(w http.ResponseWriter, req *http.Request) {
	users := []*User{}

	table := rethink.Db("gifs").Table("entries")
	userQuery := table.Map(rethink.Row.Field("Sender")).Distinct()

	userRows, _ := userQuery.Run(session)
	for userRows.Next() {
		var user string

		err := userRows.Scan(&user)
		if err != nil {
			fmt.Println(err)
		}

		countRow, err := table.Filter(rethink.Row.Field("Sender").Eq(user)).Count().RunRow(session)
		if err != nil {
			fmt.Println(err)
		}

		if !countRow.IsNil() {
			var gifCount int
			err := countRow.Scan(&gifCount)

			if err == nil {
				user := User{
					Name:  user,
					Count: gifCount,
				}

				users = append(users, &user)
			}
		}
	}

	jsonResponse, _ := json.Marshal(users)

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(jsonResponse))

}

func ChannelRSS(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get(":channel")

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Channel").Eq("#" + channel)).OrderBy(rethink.Desc("Posted")).Limit(20)
	rows, _ := query.Run(session)
	feed := CreateFeed(rows, channel)

	atom, _ := feed.ToAtom()

	w.Header().Set("Content-Type", "application/atom+xml")
	io.WriteString(w, atom)
}

func UserRSS(w http.ResponseWriter, req *http.Request) {
	user := req.URL.Query().Get(":user")

	query := rethink.Db("gifs").Table("entries").Filter(rethink.Row.Field("Sender").Eq(user)).OrderBy(rethink.Desc("Posted")).Limit(20)
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

		title := m.Sender + " posted a new gif in " + m.Channel
		if m.Nws {
			title = m.Sender + " posted a new NSFW gif in " + m.Channel
		}

		item := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: m.Url},
			Description: m.Content,
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
	m.Get("/api/count", http.HandlerFunc(GifCount))

	http.Handle("/", m)

	err := http.ListenAndServe("127.0.0.1:9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
