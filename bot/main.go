package main

import (
	"fmt"
	rethink "github.com/dancannon/gorethink"
	"github.com/ell/gifbot/helpers"
	"github.com/thoj/go-ircevent"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var session *rethink.Session

func main() {
	session = helpers.InitDB()

	message_chan := make(chan helpers.Message)
	url_chan := make(chan helpers.Message)

	chans := []string{"#secretyospos", "#cobol"}
	nsfw := []string{"nsfw", "nms", "nws", "nsfl"}
	conn := irc.IRC("ilovegifs", "ilovegifs")

	err := conn.Connect("irc.synirc.net:6667")
	if err != nil {
		fmt.Println(err)
		return
	}

	go parseMessages(message_chan, url_chan)

	conn.AddCallback("001", func(e *irc.Event) {
		for _, v := range chans {
			conn.Join(v)
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		message := strings.Split(e.Message(), " ")
		channel := strings.Split(e.Raw, " ")[2]

		if strings.Contains(message[0], "@") {
			cmd := message[0][1:len(message[0])]
			args := message[1:]

			switch cmd {
			case "top":
				TopFive(conn, args, channel)
			default:
				return
			}
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		message := helpers.Message{}
		message.Sender = e.Nick
		message.Content = e.Message()
		message.Channel = strings.Split(e.Raw, " ")[2]

		for _, w := range nsfw {
			if strings.Contains(strings.ToLower(e.Message()), w) {
				message.Nws = true
			}
		}

		if !strings.Contains(strings.ToLower(e.Nick), "bot") {
			message_chan <- message
		}
	})

	conn.Loop()
}

func TopFive(conn *irc.Connection, args []string, channel string) {
	top := helpers.GifCount(session)
	var top5 []helpers.User

	if len(top) >= 9 {
		top5 = top[0:9]
	} else {
		top5 = top[0:len(top)]
	}

	var entries []string
	for i, user := range top5 {
		entries = append(entries, fmt.Sprintf("#%d %s (%d gifs) ", i+1, user.Group, user.Reduction))
	}

	response := "Top 5 Gif Posters: " + strings.Join(entries, "|")

	conn.Privmsg(channel, response)
}

func parseMessages(message_chan chan helpers.Message, url_chan chan helpers.Message) {
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
					go addUrl(&message)
				}
			}
		}
	}
}

func addUrl(message *helpers.Message) {
	_, err := rethink.Db("gifs").Table("entries").Insert(message).RunWrite(session)
	if err != nil {
		fmt.Println(err)
	}
}
