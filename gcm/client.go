package gcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/google"
)

const (
	// FCMSendEndpoint is the endpoint for sending message to the Firebase Cloud Messaging (FCM) server.
	// See more on https://firebase.google.com/docs/cloud-messaging/server
	FCMSendEndpoint = "https://fcm.googleapis.com/v1/projects/cansukepush/messages:send"
)

const (
	// fcmPushPriorityHigh and fcmPushPriorityNormal is priority of a delivery message options
	// See more on https://firebase.google.com/docs/cloud-messaging/concept-options?hl=en#setting-the-priority-of-a-message
	fcmPushPriorityHigh   = "high"
	fcmPushPriorityNormal = "normal"
)

const (
	// maxRegistrationIDs are max number of registration IDs in one message.
	maxRegistrationIDs = 1000

	// maxTimeToLive is max time FCM storage can store messages when the device is offline
	maxTimeToLive = 2419200 // 4 weeks
)

// Client abstracts the interaction between the application server and the
// FCM server. The developer must obtain an API key from the Google APIs
// Console page and pass it to the Client so that it can perform authorized
// requests on the application server's behalf. To send a message to one or
// more devices use the Client's Send methods.
type Client struct {
	ApiKey string
	URL    string
	Http   *http.Client
}

// NewClient returns a new sender with the given URL and apiKey.
// If one of input is empty or URL is malformed, returns error.
// It sets http.DefaultHTTP client for http connection to server.
// If you need our own configuration overwrite it.
func NewClient(urlString, apiKey string) (*Client, error) {
	if len(urlString) == 0 {
		return nil, fmt.Errorf("missing FCM endpoint url")
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("missing API Key")
	}

	if _, err := url.Parse(urlString); err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %s", urlString, err)
	}

	return &Client{
		URL:    urlString,
		ApiKey: apiKey,
		Http:   http.DefaultClient,
	}, nil
}

// Send sends a message to the FCM server without retrying in case of
// service unavailability. A non-nil error is returned if a non-recoverable
// error occurs (i.e. if the response status is not "200 OK").
func (c *Client) Send(msg *Message, acsJsonData []byte) (*Response, error) {
	if err := msg.validate(); err != nil {
		return nil, err
	}

	return c.send(msg, acsJsonData)
}

func (c *Client) send(msg *Message, acsJsonData []byte) (*Response, error) {
	var buf bytes.Buffer

	//oldJsonData, _ := json.Marshal(*msg)
	//fmt.Printf("旧送信JSON(Android):%s\n\n", string(oldJsonData))

	var responses []Response

	acsToken, err := getAcsessToken(acsJsonData)
	if err != nil {
		return nil, err
	}
	for _, token := range msg.RegistrationIDs {
		messageV1 := MessageV1{
			Token:                 token,
			CollapseKey:           msg.CollapseKey,
			Data:                  msg.Data,
			DelayWhileIdle:        msg.DelayWhileIdle,
			TimeToLive:            msg.TimeToLive,
			RestrictedPackageName: msg.RestrictedPackageName,
			DryRun:                msg.DryRun,
		}
		messageV1.Notification.Title = msg.Notification.Title
		messageV1.Notification.Body = msg.Notification.Body
		messageV1.Android.Notification.Tag = msg.Notification.Tag
		messageV1.Android.Notification.ClickAction = msg.Notification.ClickAction
		messageV1.Android.Priority = msg.Priority

		wrappedMsg := WrappedMessage{messageV1}

		//jsonData, err := json.Marshal(wrappedMsg)
		//if err != nil {
		//	fmt.Println("Error marshaling to new JSON:", err)
		//	return nil, err
		//}
		//fmt.Printf("送信JSON(Android):%s\n\n", string(jsonData))

		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(wrappedMsg); err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", c.URL, &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *acsToken))
		req.Header.Add("Content-Type", "application/json")

		resp, err := c.Http.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("invalid status code %d: %s", resp.StatusCode, resp.Status)
		}

		var response Response
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&response); err != nil {
			return nil, err
		}

		// 各レスポンスをスライスに追加
		responses = append(responses, response)
	}

	return &responses[0], err
}

func getAcsessToken(acsJsonData []byte) (*string, error) {
	// // service-account.jsonを取得
	// data, err := ioutil.ReadFile("./serviceAccountKey.json")
	// if err != nil {
	// 	fmt.Printf("failed to read service account file: %v", err)
	// 	return nil, err
	// }

	// OAuth2トークンを取得するために、Googleのクレデンシャルを使用
	creds, err := google.CredentialsFromJSON(context.Background(), acsJsonData, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("error getting credentials: %v", err)
	}

	// トークンの取得
	tokenSource := creds.TokenSource
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %v", err)
	}

	// アクセストークンを表示
	//fmt.Printf("Access Token: %s\n", token.AccessToken)

	return &token.AccessToken, err
}
