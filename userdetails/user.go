package userdetails

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sinha-abhishek/jennie/cryptohelper"
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
	f, err := ioutil.ReadFile(userID + ".txt")
	if err != nil {
		log.Println("Not found")
		return user, err
	}
	b, err2 := cryptohelper.Decrypt(f, userID)
	log.Println(string(b))
	if err2 != nil {
		log.Println("can't decrypt")
		return user, err2
	}
	err = json.Unmarshal(b, user)
	//TODO : remove
	//userList = append(userList, *user)
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

	userData, _ := json.Marshal(user)
	encData, err4 := cryptohelper.Encrypt(userData, user.UserID)
	if err4 != nil {
		log.Println(err4)
		return nil, err4
	}
	//TODO: move this to a DB
	//TODO:
	err3 := ioutil.WriteFile(user.UserID+".txt", encData, os.ModeExclusive)
	// f, err3 := os.OpenFile(user.UserID+".txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err3 != nil {
		log.Fatalf("Unable to cache user: %v", err3)
		return nil, err3
	}
	// defer f.Close()
	// log.Println(encData)
	// json.NewEncoder(f).Encode(encData)
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
