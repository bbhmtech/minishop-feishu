package feishu

import (
	"context"
	"encoding/json"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func SendTextMessage(receiveIdType, receiveId, content string) *larkim.CreateMessageResp {
	c, err := json.Marshal(struct {
		Text string `json:"text"`
	}{content})
	if err != nil {
		panic(err)
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveId).
			MsgType("text").
			Content(string(c)).
			Build()).
		Build()
	resp, err := client.Im.Message.Create(context.Background(), req)
	if err != nil {
		panic(err)
	}

	return resp
}
