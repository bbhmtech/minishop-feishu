package feishu

import (
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

var client *lark.Client

func InitClient(appId, appSecret string) {
	client = lark.NewClient(appId, appSecret, lark.WithLogLevel(larkcore.LogLevelDebug))

}
