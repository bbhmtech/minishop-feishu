package feishu

import (
	"context"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

func ListRecords(appToken, tableId string) *larkbitable.ListAppTableRecordResp {
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		PageSize(500).
		Build()

	resp, err := client.Bitable.AppTableRecord.List(context.Background(), req)
	if err != nil {
		panic(err)
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		panic(resp.Msg)
	}

	fmt.Println(larkcore.Prettify(resp))
	return resp
}

func AddRecords(appToken, tableId string, records []map[string]interface{}) *larkbitable.BatchCreateAppTableRecordResp {
	reqRecords := []*larkbitable.AppTableRecord{}
	for _, v := range records {
		reqRecords = append(reqRecords, larkbitable.
			NewAppTableRecordBuilder().
			Fields(v).
			Build())
	}
	req := larkbitable.NewBatchCreateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		Body(
			larkbitable.NewBatchCreateAppTableRecordReqBodyBuilder().
				Records(reqRecords).
				Build(),
		).
		Build()

	resp, err := client.Bitable.AppTableRecord.BatchCreate(context.Background(), req)
	if err != nil {
		panic(err)
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		panic(resp.Msg)
	}

	fmt.Println(larkcore.Prettify(resp))
	return resp
}

func UpdateRecords(appToken, tableId string, records map[string](map[string]interface{})) *larkbitable.BatchUpdateAppTableRecordResp {
	reqRecords := []*larkbitable.AppTableRecord{}
	for k, v := range records {
		reqRecords = append(reqRecords, larkbitable.
			NewAppTableRecordBuilder().
			RecordId(k).
			Fields(v).
			Build())
	}
	req := larkbitable.NewBatchUpdateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		Body(
			larkbitable.NewBatchUpdateAppTableRecordReqBodyBuilder().
				Records(reqRecords).
				Build(),
		).
		Build()

	resp, err := client.Bitable.AppTableRecord.BatchUpdate(context.Background(), req)
	if err != nil {
		panic(err)
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		panic(resp.Msg)
	}

	fmt.Println(larkcore.Prettify(resp))
	return resp
}
