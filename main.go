package main

import (
	"fmt"
	rethink "github.com/dancannon/gorethink"
	"github.com/thoj/go-ircevent"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Message struct {
	Sender  string
	Content string
	Channel string
	Url     string
	Posted  time.Time
}

func main() {
	session := InitDB()

	message_chan := make(chan Message)
	url_chan := make(chan Message)

	chans := []string{"#secretyospos", "#cobol"}
	conn := irc.IRC("ilovegifs", "ilovegifs")

	err := conn.Connect("irc.synirc.net:6667")
	if err != nil {
		fmt.Println(err)
		return
	}

	go parseMessages(message_chan, url_chan)
	go addUrl(url_chan, session)

	conn.AddCallback("001", func(e *irc.Event) {
		for _, v := range chans {
			conn.Join(v)
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		message := Message{}
		message.Sender = e.Nick
		message.Content = e.Message()
		message.Channel = strings.Split(e.Raw, " ")[2]

		message_chan <- message
	})

	conn.Loop()
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

func parseMessages(message_chan chan Message, url_chan chan Message) {
	r, _ := regexp.Compile("(http|https):\\/\\/([\\w\\-_]+(?:(?:\\.[\\w\\-_]+)+))([\\w\\-\\.,@?^=%&amp;:/~\\+#]*[\\w\\-\\@?^=%&amp;/~\\+#])?")

	for message := range message_chan {
		url := r.FindString(message.Content)
		if url != "" {
			resp, err := http.Get(url)
			if err == nil {
				contentType := resp.Header.Get("Content-Type")
				if contentType == "image/gif" {
					message.Url = url
					message.Posted = time.Now()
					url_chan <- message
				}
			}
		}
	}
}

func addUrl(url_chan chan Message, session *rethink.Session) {
	for message := range url_chan {
		_, err := rethink.Db("gifs").Table("entries").Insert(message).RunWrite(session)
		if err != nil {
			fmt.Println(err)
		}
	}
}
