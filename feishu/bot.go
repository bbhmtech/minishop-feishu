package feishu

import (
	"context"
	"encoding/json"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type BotInfo struct {
	ActivateStatus int    `json:"activate_status"`
	AppName        string `json:"app_name"`
	AvatarURL      string `json:"avatar_url"`
	IPWhiteList    []any  `json:"ip_white_list"`
	OpenID         string `json:"open_id"`
}

func GetBotInfo() *BotInfo {
	resp, err := client.Do(context.Background(), &larkcore.ApiReq{
		HttpMethod:                http.MethodGet,
		ApiPath:                   "https://open.feishu.cn/open-apis/bot/v3/info",
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	})
	type BotInfoResp struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Bot  BotInfo `json:"bot"`
	}
	var botInfo BotInfoResp
	err = json.Unmarshal(resp.RawBody, &botInfo)
	if err != nil {
		panic(err)
	}

	if botInfo.Code != 0 {
		panic(botInfo.Msg)
	}

	return &botInfo.Bot
}
