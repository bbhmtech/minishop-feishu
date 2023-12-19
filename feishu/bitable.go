package feishu

import (
	"context"
	"fmt"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

func ListRecords(appToken, tableId, filter string) *larkbitable.ListAppTableRecordResp {
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		PageSize(500).
		Filter(filter).
		Build()

	resp, err := client.Bitable.AppTableRecord.List(context.Background(), req)
	if err != nil {
		panic(err)
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		panic(resp.Msg)
	}

	// fmt.Println(larkcore.Prettify(resp))
	return resp
}

func IterRecords(appToken, tableId, filter string) *larkbitable.ListAppTableRecordIterator {
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		Filter(filter).
		PageSize(50).
		Build()

	iter, err := client.Bitable.AppTableRecord.ListByIterator(context.Background(), req)
	if err != nil {
		panic(err)
	}

	return iter
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

	// fmt.Println(larkcore.Prettify(resp))
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

	// fmt.Println(larkcore.Prettify(resp))
	return resp
}
