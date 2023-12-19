package wx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var _appId, _appSecret string

var wxAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

func GetStableAccessToken() string {
	if wxAccessToken.ExpiresAt.After(time.Now()) {
		return wxAccessToken.AccessToken
	}

	body, err := json.Marshal(map[string]interface{}{
		"grant_type": "client_credential",
		"appid":      _appId,
		"secret":     _appSecret,
	})
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("https://api.weixin.qq.com/cgi-bin/stable_token", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &wxAccessToken)
	if err != nil {
		panic(err)
	}

	wxAccessToken.ExpiresAt = time.Now().Add(time.Second * time.Duration(wxAccessToken.ExpiresIn))

	return wxAccessToken.AccessToken
}

func InitClient(appId, appSecret string) {
	_appId, _appSecret = appId, appSecret
}
