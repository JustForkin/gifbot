package helpers

import (
	rethink "github.com/dancannon/gorethink"
	"log"
	"time"
)

type Message struct {
	Sender  string
	Content string
	Channel string
	Url     string
	Posted  time.Time
	Nws     bool
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
