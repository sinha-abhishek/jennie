package userdetails

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/sinha-abhishek/jennie/linkedin"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

type User struct {
	UserID            string       `json:"uid"`
	LastLinkedinFetch time.Time    `json:"lastLinkedinFetch"`
	LastEmailScan     time.Time    `json:"lastEmailScan"`
	Token             oauth2.Token `json:"token"`
}

var (
	userList = make([]User, 0)
)

func GetUser(userID string) (*User, error) {
	user := &User{}
	f, err := os.Open(userID + ".txt")
	if err != nil {
		return user, err
	}
	err = json.NewDecoder(f).Decode(user)
	//TODO : remove
	userList = append(userList, *user)
	return user, err
}

func FetchAndSaveUser(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*User, error) {
	client := config.Client(ctx, token)
	srv, err := gmail.New(client)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	u := "me"
	res, err2 := srv.Users.GetProfile(u).Do()
	if err2 != nil {
		log.Println(err2)
		return nil, err2
	}
	user := User{}
	user.UserID = res.EmailAddress
	user.Token = *token
	//TODO: move this to a DB
	f, err3 := os.OpenFile(user.UserID+".txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err3 != nil {
		log.Fatalf("Unable to cache user: %v", err3)
		return nil, err3
	}
	defer f.Close()
	json.NewEncoder(f).Encode(user)
	userList = append(userList, user)
	return &user, nil
}

func PeriodicPuller(ctx context.Context, config *oauth2.Config) {
	t := time.NewTimer(15 * time.Second)
	for {
		select {
		case <-t.C:
			for _, user := range userList {
				user.DoEmailAutomationForUser(ctx, config)
				t.Reset(15 * time.Second)
			}
		}
	}
}

func InitUserList() {
	//TODO:
}

func (user *User) DoEmailAutomationForUser(ctx context.Context, config *oauth2.Config) {
	linkedin.SearchMailAndRespond(ctx, config, &user.Token, user.UserID)
}
