package wx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var _appId, _appSecret string

const WeixinTimeFormat = "2006-01-02 15:04:05"

var WeixinLocation *time.Location

var wxAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

func wxPostJSON(url string, body interface{}) []byte {
	buf, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var p struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		panic(err)
	}

	if p.ErrCode != 0 {
		panic(p.ErrMsg)
	}

	return buf
}

func GetStableAccessToken() string {
	if wxAccessToken.ExpiresAt.After(time.Now()) {
		return wxAccessToken.AccessToken
	}

	buf := wxPostJSON("https://api.weixin.qq.com/cgi-bin/stable_token", map[string]interface{}{
		"grant_type": "client_credential",
		"appid":      _appId,
		"secret":     _appSecret,
	})

	err := json.Unmarshal(buf, &wxAccessToken)
	if err != nil {
		panic(err)
	}

	wxAccessToken.ExpiresAt = time.Now().Add(time.Second * time.Duration(wxAccessToken.ExpiresIn))

	return wxAccessToken.AccessToken
}

func daysAgo(pastDays int) (now, past time.Time) {
	now = time.Now().In(WeixinLocation) // important: timezone must be Beijing Time UTC+8
	past = now.Add(time.Hour * 24 * -time.Duration(pastDays))
	return
}

func InitClient(appId, appSecret string) {
	_appId, _appSecret = appId, appSecret

	var err error
	WeixinLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
}
