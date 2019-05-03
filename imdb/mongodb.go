package imdb

import "github.com/globalsign/mgo"

func ConnectDB(url string) (*mgo.Session, error) {
	ses, err := mgo.Dial(url)
	return ses, err
}
