package helpers

import (
	"fmt"
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

type User struct {
	Group     string `json:"name"`
	Reduction int    `json:"count"`
}

func GifCount(session *rethink.Session) []*User {
	users := []*User{}

	table := rethink.Db("gifs").Table("entries")
	userQuery := table.Group("Sender").Count()

	userRows, _ := userQuery.Run(session)
	for userRows.Next() {
		var user User

		err := userRows.Scan(&user)
		if err != nil {
			fmt.Println(err)
		}

		users = append(users, &user)
	}

	return users
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
