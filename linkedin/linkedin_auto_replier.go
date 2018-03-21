package linkedin

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sinha-abhishek/jennie/awshelper"
	"github.com/sinha-abhishek/jennie/confighelper"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

var (
	searchString = "from:*@linkedin.com replyto:*@linkedin.com -replyto:*donotreply*"
)

func InitializeLinkedinResponder() {
	//TODO: move this to DynamoDB
	// log.Println("Reading responded file")
	// f, err := os.Open("responded.txt")
	// RespondedIdsMap = make(map[string]bool)
	// if err != nil {
	// 	log.Println("Failed to read responded file")
	// 	return
	// }
	// data, err := ioutil.ReadAll(f)
	// str := string(data)
	// s := strings.Split(str, "\n")
	// for _, v := range s {
	// 	RespondedIdsMap[v] = true
	// }
	// log.Println("Responded = ", RespondedIdsMap)
}

func getTimeOfLastScan(lastFetchTime time.Time) time.Time {
	//TODO: store last responded timestamp and get after that
	if time.Since(lastFetchTime) > 24*7*time.Hour {
		t := time.Now()
		t2 := t.AddDate(0, 0, -7)
		return t2
	}
	return lastFetchTime
}

func SearchMailAndRespond(ctx context.Context, config *oauth2.Config, token *oauth2.Token, uid string, lastFetchTime time.Time) error {
	qtime := getTimeOfLastScan(lastFetchTime).Unix()
	log.Println("token=", token)
	qtimeString := strconv.FormatInt(qtime, 10)
	query := searchString + " after:" + qtimeString
	client := config.Client(ctx, token)
	srv, err := gmail.New(client)
	if err != nil {
		log.Println(err)
		return err
	}
	user := uid
	log.Println(query)
	res, err2 := srv.Users.Messages.List(user).MaxResults(10).Q(query).Do()
	if err2 != nil {
		log.Println(err2)
		//http.Error(w, "Can't get messages", http.StatusInternalServerError)
		return err2
	}
	messages := res.Messages
	for _, msg := range messages {
		log.Println(msg.Id)
		mail, err3 := srv.Users.Messages.Get(user, msg.Id).
			MetadataHeaders("Subject", "Date", "Reply-To").Format("metadata").Do()
		if err3 != nil {
			log.Println(err3)
			return err3
		}
		log.Println(mail.Snippet)
		headerMap := make(map[string]string)
		for _, part := range mail.Payload.Headers {
			// p, _ := base64.StdEncoding.DecodeString(part.Body.Data)
			// log.Println(string(p))
			headerMap[part.Name] = part.Value
		}
		log.Println("Got ", headerMap)
		//TODO: reply to replyto fileds
		//srv.Users.Messages.Send(user, message)
		_, err4 := reply(uid, srv, msg, headerMap)
		if err4 != nil {
			return err4
		}

	}
	return nil
}

func ClearRespondedIds(uid string) {
	// if os.Remove("Responded.txt") == nil {
	// 	RespondedIdsMap = make(map[string]bool)
	// }
	awshelper.ClearRespondedIds(uid)
}

func reply(uid string, srv *gmail.Service, msg *gmail.Message, headerMap map[string]string) (string, error) {
	thId := msg.ThreadId
	respondedDetails, err := awshelper.GetRespondedIdsForUser(uid)
	if err != nil {
		log.Println("Can't read from dynamo")
		return "", err
	}
	if respondedDetails.Uid == uid && respondedDetails.RespondedIds[thId] == true {
		log.Println("Done already for ", thId)
		return "already handle", nil
	}
	var raw string
	raw = "From: " + uid
	raw += "\nTo: " + headerMap["Reply-To"]
	raw += "\nSubject: " + headerMap["Subject"]
	t := time.Now()
	d := t.Format(time.RFC1123Z)
	raw += "\nDate: " + d
	raw += "\nMessage-Id:<CAANbcPMiNyFVQstZH89+isO6iXYBz0bhc9V=M+H9FtiGbu_nCg@mail.gmail.com>"
	autoConfig, err := confighelper.GetAutoResponseConfig()
	if err != nil {
		return "", err
	}
	raw += "\n" + autoConfig.LinkedinResponse
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	s := strings.Replace(encoded, "+", "-", -1)
	s = strings.Replace(s, "/", "_", -1)
	log.Println(raw)
	log.Println(s)
	message := &gmail.Message{}
	message.Raw = s
	message.ThreadId = thId

	res, sendErr := srv.Users.Messages.Send(uid, message).Do()
	log.Println(res.HTTPStatusCode)
	if res.HTTPStatusCode == http.StatusOK {
		log.Println("Replied to threadId", thId)
		AddToRespondedIds(thId, uid)
	}
	return res.Raw, sendErr

}

func AddToRespondedIds(id string, uid string) {
	err := awshelper.AppendRespondedID(id, uid)
	if err != nil {
		return
	}
}
