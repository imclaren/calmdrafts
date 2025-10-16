package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Client wraps the Gmail API client
type Client struct {
	service *gmail.Service
}

// Draft represents a Gmail draft with relevant information
type Draft struct {
	ID           string
	MessageID    string
	Subject      string
	To           string
	InternalDate time.Time
	IsEmpty      bool
}

// NewClient creates a new Gmail API client with OAuth2 authentication
func NewClient(ctx context.Context, credentialsPath, tokenPath string) (*Client, error) {
	config, err := getOAuthConfig(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	token, err := getToken(tokenPath, config)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %v", err)
	}

	httpClient := config.Client(ctx, token)
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return &Client{service: service}, nil
}

// getOAuthConfig loads OAuth configuration from credentials file
func getOAuthConfig(credentialsPath string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailModifyScope)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// getToken retrieves a token from file or prompts user to authorize
func getToken(tokenPath string, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := tokenFromFile(tokenPath)
	if err == nil {
		return token, nil
	}

	token, err = getTokenFromWeb(config)
	if err != nil {
		return nil, err
	}

	if err := saveToken(tokenPath, token); err != nil {
		return nil, err
	}

	return token, nil
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code:\n%v\n", authURL)
	fmt.Print("Authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	return token, nil
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// ListDrafts retrieves all drafts from Gmail
func (c *Client) ListDrafts(ctx context.Context) ([]*Draft, error) {
	user := "me"
	drafts := []*Draft{}

	err := c.service.Users.Drafts.List(user).Pages(ctx, func(response *gmail.ListDraftsResponse) error {
		for _, draft := range response.Drafts {
			draftDetail, err := c.service.Users.Drafts.Get(user, draft.Id).Format("full").Do()
			if err != nil {
				fmt.Printf("Error fetching draft %s: %v\n", draft.Id, err)
				continue
			}

			d := &Draft{
				ID:        draft.Id,
				MessageID: draftDetail.Message.Id,
			}

			// Parse internal date
			if draftDetail.Message.InternalDate > 0 {
				d.InternalDate = time.Unix(draftDetail.Message.InternalDate/1000, 0)
			}

			// Extract subject and to fields
			for _, header := range draftDetail.Message.Payload.Headers {
				switch header.Name {
				case "Subject":
					d.Subject = header.Value
				case "To":
					d.To = header.Value
				}
			}

			// Check if draft is empty (no subject, no recipient, no body)
			d.IsEmpty = d.Subject == "" && d.To == "" && isEmpty(draftDetail.Message.Payload)

			drafts = append(drafts, d)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve drafts: %v", err)
	}

	return drafts, nil
}

// isEmpty checks if a message payload has any content
func isEmpty(payload *gmail.MessagePart) bool {
	if payload == nil {
		return true
	}

	// Check if body has data
	if payload.Body != nil && payload.Body.Size > 0 {
		return false
	}

	// Check parts recursively
	for _, part := range payload.Parts {
		if !isEmpty(part) {
			return false
		}
	}

	return true
}

// DeleteDraft deletes a draft by ID
func (c *Client) DeleteDraft(ctx context.Context, draftID string) error {
	user := "me"
	err := c.service.Users.Drafts.Delete(user, draftID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete draft %s: %v", draftID, err)
	}
	return nil
}
