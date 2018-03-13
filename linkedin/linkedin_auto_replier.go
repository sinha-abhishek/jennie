package linkedin

import (
	"context"
	"log"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

var (
	searchString = "from:*@linkedin.com replyto:*@linkedin.com -replyto:*donotreply*"
)

func getTimeOfLastScan() time.Time {
	//TODO: store last responded timestamp and get after that
	t := time.Now()
	t2 := t.AddDate(0, 0, -2)
	return t2
}

func SearchMailAndRespond(ctx context.Context, config *oauth2.Config, token *oauth2.Token, uid string) {
	qtime := getTimeOfLastScan().Unix()
	qtimeString := strconv.FormatInt(qtime, 10)
	query := searchString + " after:" + qtimeString
	client := config.Client(ctx, token)
	srv, err := gmail.New(client)
	if err != nil {
		log.Println(err)
		return
	}
	user := uid
	log.Println(query)
	res, err2 := srv.Users.Messages.List(user).MaxResults(10).Q(query).Do()
	if err2 != nil {
		log.Println(err2)
		//http.Error(w, "Can't get messages", http.StatusInternalServerError)
		return
	}
	messages := res.Messages
	for _, msg := range messages {
		log.Println(msg.Id)
		mail, err3 := srv.Users.Messages.Get(user, msg.Id).
			MetadataHeaders("Subject", "Date", "Reply-To").Format("metadata").Do()
		if err3 != nil {
			log.Println(err3)
			return
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

	}

}
