package userdetails

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sinha-abhishek/jennie/awshelper"
	"github.com/sinha-abhishek/jennie/confighelper"
	"github.com/sinha-abhishek/jennie/cryptohelper"
	"github.com/sinha-abhishek/jennie/linkedin"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/plus/v1"
)

type User struct {
	UserID            string       `json:"uid"`
	LastLinkedinFetch time.Time    `json:"lastLinkedinFetch"`
	LastEmailScan     time.Time    `json:"lastEmailScan"`
	Token             oauth2.Token `json:"token"`
	Name              string       `json:"name"`
	LinkedinMessage   string       `json:"auto_reply"`
}

var (
	userList = make([]User, 0)
)

func GetUser(userID string) (*User, error) {
	ud, _ := awshelper.FetchUser(userID)
	log.Println("user dunamo=", string(ud))
	user := &User{}
	b, err2 := cryptohelper.Decrypt(ud, userID)
	log.Println(string(b))
	if err2 != nil {
		log.Println("can't decrypt")
		return user, err2
	}
	err := json.Unmarshal(b, user)
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

	srv2, err := plus.New(client)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	p, err := srv2.People.Get("me").Do()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("person=", p.DisplayName)

	savedUser, err3 := GetUser(res.EmailAddress)
	var user *User
	if err3 == nil && savedUser.UserID == res.EmailAddress {
		log.Println("exisitng user ", res.EmailAddress)
		user = savedUser
	} else {
		user = &User{}
	}
	user.UserID = res.EmailAddress
	user.Token = *token
	user.Name = p.DisplayName
	autoConfig, err := confighelper.GetAutoResponseConfig()
	if err != nil {
		return nil, err
	}
	rText := fmt.Sprintf(autoConfig.LinkedinResponse, user.Name, user.Name, user.UserID)
	log.Println("replytext=", rText)
	user.LinkedinMessage = rText
	err = user.Save()
	//userList = append(userList, user)
	err = awshelper.SendUpdateMessage("uid", user.UserID, 800)
	return user, err
}

func (user *User) Save() error {
	userData, _ := json.Marshal(user)
	encData, err := cryptohelper.Encrypt(userData, user.UserID)
	if err != nil {
		log.Println(err)
		return err
	}
	err = awshelper.SaveUser(user.UserID, string(encData))
	if err != nil {
		log.Println("Failed to save user", err)
	}
	return err
}

func PeriodicPuller(ctx context.Context, config *oauth2.Config) {
	t := time.NewTimer(15 * time.Minute) //TODO: fix time period
	for {
		select {
		case <-t.C:
			log.Println("pulling...")
			msgs, err := awshelper.GetUpdateMessages("uid")
			var success []*string
			var uidsSuccess []string
			if err == nil {
				for _, v := range msgs {
					uid := *v.Body
					user, err2 := GetUser(uid)
					if err2 == nil {
						user.DoEmailAutomationForUser(ctx, config)
						uidsSuccess = append(uidsSuccess, uid)
						success = append(success, v.ReceiptHandle)
					}
				}
				awshelper.DeleteMessages(success)
				for _, u := range uidsSuccess {
					awshelper.SendUpdateMessage("uid", u, 800)
				}
			}
			t.Reset(30 * time.Minute)
		}
	}
}

func InitUserList() {
	//TODO:
}

func (user *User) DoEmailAutomationForUser(ctx context.Context, config *oauth2.Config) {
	err := linkedin.SearchMailAndRespond(ctx, config, &user.Token, user.UserID, user.LastLinkedinFetch, user.LinkedinMessage)
	if err == nil {
		user.LastLinkedinFetch = time.Now()
		err = user.Save()
		if err == nil {
			linkedin.ClearRespondedIds(user.UserID)
		}
	}
}
