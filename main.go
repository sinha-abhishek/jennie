package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sinha-abhishek/jennie/userdetails"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var config *oauth2.Config
var ctx context.Context

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func authorizeGmail(w http.ResponseWriter, r *http.Request) {
	config.RedirectURL = "http://localhost:8000/jennie/onauthcallback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

func onAuthDone(w http.ResponseWriter, r *http.Request) {
	log.Println("Got callback")
	codes, ok := r.URL.Query()["code"]
	if !ok || len(codes) < 1 {
		log.Println("code not recieved")
		http.Error(w, "Code not recieved", http.StatusBadRequest)
	}
	code := codes[0]
	log.Println("Code=" + code)
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	log.Println(tok)
	user, err2 := userdetails.FetchAndSaveUser(ctx, config, tok)
	if err2 != nil {
		user.DoEmailAutomationForUser(ctx, config)
	}

	// client := config.Client(ctx, tok)
	// srv, err := gmail.New(client)
	// user := "me"
	// res, err2 := srv.Users.Messages.List(user).MaxResults(10).Q("from:*@linkedin.com replyto:*@linkedin.com newer_than:7d").Do()
	// if err2 != nil {
	// 	log.Println(err2)
	// 	http.Error(w, "Can't get messages", http.StatusInternalServerError)
	// 	return
	// }
	// messages := res.Messages
	// for _, msg := range messages {
	// 	log.Println(msg.Id)
	// 	mail, err3 := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
	// 	if err3 != nil {
	// 		log.Println(err3)
	// 		http.Error(w, "Can't get messages", http.StatusInternalServerError)
	// 		return
	// 	}
	// 	//log.Println(mail.Payload.Body.Data)
	// 	log.Println(mail.Snippet)
	// 	//log.Println(mail.Payload.H)
	// 	data := ""
	// 	for _, part := range mail.Payload.Parts {
	// 		// p, _ := base64.StdEncoding.DecodeString(part.Body.Data)
	// 		// log.Println(string(p))
	// 		data += part.Body.Data
	// 	}
	// 	d, _ := base64.StdEncoding.DecodeString(data)
	// 	log.Println(string(d))
	// }
	// //log.Println(res)
	//
	// w.Write(([]byte)("Got success"))
	//linkedin.SearchMailAndRespond(ctx, config, tok)

	w.Write(([]byte)("Got success"))
}

func doTaskForUser(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	log.Println("host=", host)
	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) < 1 {
		log.Println("id not recieved")
		http.Error(w, "id not recieved", http.StatusBadRequest)
		return
	}
	id := ids[0]
	user, err := userdetails.GetUser(id)
	log.Println("token=", user.Token)
	if err != nil {
		log.Println("id not recieved")
		http.Redirect(w, r, "http://"+host+"/jennie/authorize", http.StatusTemporaryRedirect)
		//http.Error(w, "id not recieved", http.StatusBadRequest)
	} else {
		//linkedin.SearchMailAndRespond(ctx, config, token)
		user.DoEmailAutomationForUser(ctx, config)
		log.Println("I have user ", user)
	}

}

func startServer() {
	http.HandleFunc("/jennie/authorize", authorizeGmail)
	http.HandleFunc("/jennie/onauthcallback", onAuthDone)
	http.HandleFunc("/jennie/user", doTaskForUser)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Println(err)
		panic(1)
	}
}

func main() {
	ctx = context.Background()
	var err1 error
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err1 = google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailComposeScope, gmail.GmailModifyScope, gmail.GmailReadonlyScope)
	if err1 != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	go userdetails.PeriodicPuller(ctx, config)
	startServer()
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels. %v", err)
	}
	if len(r.Labels) > 0 {
		fmt.Print("Labels:\n")
		for _, l := range r.Labels {
			fmt.Printf("- %s\n", l.Name)
		}
	} else {
		fmt.Print("No labels found.")
	}

}
