package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	"github.com/yiffyi/minishop-feishu/feishu"
	"github.com/yiffyi/minishop-feishu/wx"
)

var localTimeLocation *time.Location

var (
	orderStatusMapping = map[int]string{
		10:  "待付款",
		15:  "等待成团",
		16:  "待接单",
		17:  "待核销",
		20:  "待发货",
		21:  "部分发货",
		30:  "待收货",
		100: "完成",
		181: "取消-自动",
		190: "取消-超卖",
		200: "取消-售后",
		250: "取消-超时",
	}
	shippingMethodMap = map[string]string{
		"ShippingMethod_Express":  "快递",
		"ShippingMethod_SameCity": "同城配送",
		"ShippingMethod_Pickup":   "自提",
	}
)

var (
	tableId, appToken string
)

func mapWxOrderToFeishuRecord(o wx.Order) map[string]interface{} {
	createTime, err := time.ParseInLocation("2006-01-02 15:04:05", o.CreateTime, localTimeLocation)
	if err != nil {
		panic(err)
	}
	productName := []string{}
	for _, v := range o.OrderDetail.ProductInfos {
		a := v.Title
		for _, attr := range v.SkuAttrs {
			a += "," + attr.AttrValue
		}
		a += " x " + strconv.Itoa(v.SkuCnt)
		productName = append(productName, a)
	}

	return map[string]interface{}{
		"订单号":  strconv.FormatInt(o.OrderID, 10),
		"支付单号": o.OrderDetail.PayInfo.TransactionID,
		"下单时间": createTime.UnixMilli(),
		"订单状态": orderStatusMapping[o.Status],
		"商品名":  strings.Join(productName, ";"),
		"客户备注": o.ExtInfo.CustomerNotes,
		"商家备注": o.ExtInfo.MerchantNotes,
		"客户姓名": o.OrderDetail.DeliveryInfo.AddressInfo.UserName,
		"收货地址": o.OrderDetail.DeliveryInfo.AddressInfo.DetailInfo,
		"配送方式": shippingMethodMap[o.OrderDetail.DeliveryInfo.ExpressFee[0].ShippingMethod],
	}

}

func pullMinishopOrders(operatorOpenId string) {
	// fmt.Println(wx.ListOrders())

	// key: recordId
	existingRecordsMap := map[string](map[string]interface{}){}
	// key: orderId
	newRecordsMap := map[string](map[string]interface{}){}
	for _, v := range wx.ListOrders() {
		newRecordsMap[strconv.FormatInt(v.OrderID, 10)] = mapWxOrderToFeishuRecord(v)
		// fmt.Println()
	}

	for _, v := range feishu.ListRecords("Hca7bQbAQay3y8siHD8cmtKmnZc", "tblRPH2xhfQKWrJG").Data.Items {
		orderId := *v.StringField("订单号")
		if r, exists := newRecordsMap[orderId]; exists {
			existingRecordsMap[*v.RecordId] = r
			delete(newRecordsMap, orderId)
		}
	}

	newRecords := []map[string]interface{}{}
	for _, v := range newRecordsMap {
		newRecords = append(newRecords, v)
	}

	// existingRecords := []map[string]interface{}{}
	// for _, v := range existingRecordsMap {
	// 	existingRecords = append(existingRecords, v)
	// }

	if len(newRecords) > 0 {
		feishu.AddRecords(appToken, tableId, newRecords)
	}
	feishu.UpdateRecords(appToken, tableId, existingRecordsMap)

	feishu.SendTextMessage("open_id", operatorOpenId, fmt.Sprint("同步完成，本次新增", len(newRecords), "条，更新", len(existingRecordsMap), "条。"))
}

func main() {

	var err error
	localTimeLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	feishu.InitClient(os.Getenv("FEISHU_APPID"), os.Getenv("FEISHU_APPSECRET"))
	wx.InitClient(os.Getenv("WEIXIN_APPID"), os.Getenv("WEIXIN_APPSECRET"))

	tableId, appToken = os.Getenv("FEISHU_TABLEID"), os.Getenv("FEISHU_APPTOKEN")

	handler := dispatcher.NewEventDispatcher(os.Getenv("FEISHU_VERIFICATION"), os.Getenv("FEISHU_EVENTENCCODE")).
		OnP2BotMenuV6(func(ctx context.Context, event *larkapplication.P2BotMenuV6) error {
			if *event.Event.EventKey == "pullMinishopOrders" {
				feishu.SendTextMessage("open_id", *event.Event.Operator.OperatorId.OpenId, "正在处理……")
				go pullMinishopOrders(*event.Event.Operator.OperatorId.OpenId)
			}
			return nil
		})

	http.HandleFunc("/feishu/event", httpserverext.NewEventHandlerFunc(handler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	err = http.ListenAndServe(os.Getenv("LISTEN_ADDR"), nil)
	if err != nil {
		panic(err)
	}
}
