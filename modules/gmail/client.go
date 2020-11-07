/*
* This butt-ugly code is direct from Google itself
* https://developers.google.com/calendar/quickstart/go
 */

package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/wtfutil/wtf/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

/* -------------------- Exported Functions -------------------- */

type GmailMessage struct {
	date    string
	snippet string
	from    string
	subject string
	id      string
	payload *gmail.MessagePart
}

type GmailClient struct {
	service  *gmail.Service
	settings *Settings
}

func NewClient(settings *Settings) *GmailClient {
	ctx := context.Background()

	secretPath, _ := utils.ExpandHomeDir(settings.secretFile)

	b, err := ioutil.ReadFile(secretPath)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &GmailClient{
		service:  srv,
		settings: settings,
	}
}

func (client *GmailClient) Fetch() ([]*GmailMessage, error) {
	query := client.settings.searchQuery
	messageResponse, err := client.service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return nil, err
	}

	gmailMessages := []*GmailMessage{}
	messageLimit := int(client.settings.mailCount)
	if len(messageResponse.Messages) < messageLimit {
		messageLimit = len(messageResponse.Messages)
	}
	for i := 0; i < messageLimit; i++ {
		id := messageResponse.Messages[i].Id
		msg, err := client.service.Users.Messages.Get("me", id).Do()
		if err != nil {
			return nil, err
		}
		date := ""
		from := ""
		subject := ""
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				date = h.Value
			} else if h.Name == "From" {
				from = h.Value
			} else if h.Name == "Subject" {
				subject = h.Value
			}
		}
		gmailMessages = append(gmailMessages, &GmailMessage{
			date:    date,
			snippet: msg.Snippet,
			from:    from,
			subject: subject,
			id:      id,
			payload: msg.Payload,
		})
	}

	return gmailMessages, err
}

func (client *GmailClient) Archive(message *GmailMessage) error {
	if message == nil {
		return nil
	}

	removeLabelRequest := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}
	_, err := client.service.Users.Messages.Modify("me", message.id, removeLabelRequest).Do()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (client *GmailClient) Trash(message *GmailMessage) error {
	if message == nil {
		return nil
	}

	_, err := client.service.Users.Messages.Trash("me", message.id).Do()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

/* -------------------- Unexported Functions -------------------- */

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
