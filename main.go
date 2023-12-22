package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func concatProductNames(o wx.Order) string {
	productName := []string{}
	for _, v := range o.OrderDetail.ProductInfos {
		a := v.Title
		for _, attr := range v.SkuAttrs {
			a += "," + attr.AttrValue
		}
		a += " x " + strconv.Itoa(v.SkuCnt)
		productName = append(productName, a)
	}
	return strings.Join(productName, ";")
}

func mapWxOrderToFeishuRecord(o wx.Order, shippingState int) map[string]interface{} {
	createTime, err := time.ParseInLocation(wx.WeixinTimeFormat, o.CreateTime, wx.WeixinLocation)
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
		"订单号":    strconv.FormatInt(o.OrderID, 10),
		"支付单号":   o.OrderDetail.PayInfo.TransactionID,
		"下单时间":   createTime.UnixMilli(),
		"订单状态":   orderStatusMapping[o.Status],
		"商品名":    concatProductNames(o),
		"客户备注":   o.ExtInfo.CustomerNotes,
		"商家备注":   o.ExtInfo.MerchantNotes,
		"客户姓名":   o.OrderDetail.DeliveryInfo.AddressInfo.UserName,
		"收货地址":   o.OrderDetail.DeliveryInfo.AddressInfo.DetailInfo,
		"预留电话":   o.OrderDetail.DeliveryInfo.AddressInfo.TelNumber,
		"配送方式":   shippingMethodMap[o.OrderDetail.DeliveryInfo.ExpressFee[0].ShippingMethod],
		"发货信息上传": shippingState >= 2,
	}

}

func pullMinishopOrders() (created, updated int) {
	// fmt.Println(wx.ListOrders())

	shippingInfo := wx.ListShippingInfo(15)

	// key: recordId
	existingRecordsMap := make(map[string](map[string]interface{}), 50)
	// key: orderId
	newRecordsMap := map[string](map[string]interface{}){}
	for k, v := range wx.ListOrders(15) {
		tId := v.OrderDetail.PayInfo.TransactionID
		newRecordsMap[k] = mapWxOrderToFeishuRecord(v, shippingInfo[tId])
		// fmt.Println()
	}

	iter := feishu.IterRecords(appToken, tableId, "today()-15 <= CurrentValue.[下单时间]")
	for {
		hasNext, v, err := iter.Next()
		if err != nil {
			panic(err)
		}

		if !hasNext {
			break
		}

		orderId := *v.StringField("订单号")
		if r, exists := newRecordsMap[orderId]; exists {
			existingRecordsMap[*v.RecordId] = r
			delete(newRecordsMap, orderId)

			if len(existingRecordsMap) == 50 {
				feishu.UpdateRecords(appToken, tableId, existingRecordsMap)
				updated += 50
				existingRecordsMap = make(map[string](map[string]interface{}), 50)
			}
		}
	}

	if len(existingRecordsMap) > 0 {
		updated += len(existingRecordsMap)
		feishu.UpdateRecords(appToken, tableId, existingRecordsMap)
	}

	newRecords := []map[string]interface{}{}
	for _, v := range newRecordsMap {
		newRecords = append(newRecords, v)
	}

	// existingRecords := []map[string]interface{}{}
	// for _, v := range existingRecordsMap {
	// 	existingRecords = append(existingRecords, v)
	// }

	for i := 0; i < len(newRecords); i += 50 {
		feishu.AddRecords(appToken, tableId, newRecords[i:min(len(newRecords), i+50)])
	}
	created += len(newRecords)
	return
}

func pushShippingInfo() int {
	shippingInfo := wx.ListShippingInfo(15)
	updated := 0
	for _, v := range wx.ListOrders(15) {
		tId := v.OrderDetail.PayInfo.TransactionID

		if v.Status > 20 && v.Status <= 100 {
			if r, exists := shippingInfo[tId]; r < 2 && exists {
				logisticsType := 0
				if v.OrderDetail.DeliveryInfo.ExpressFee[0].ShippingMethod == "ShippingMethod_SameCity" {
					logisticsType = 2
				} else if v.OrderDetail.DeliveryInfo.ExpressFee[0].ShippingMethod == "ShippingMethod_Pickup" {
					logisticsType = 4
				}
				err := wx.UploadShippingInfo(tId, logisticsType, concatProductNames(v))
				if err != nil {
					panic(err)
				}
				updated++
			}
		}
	}
	if updated > 0 {
		pullMinishopOrders()
	}
	return updated
}

func setDelivered() int {
	orders := wx.ListOrders(15)
	updated := 0

	iter := feishu.IterRecords(appToken, tableId, `today()-15 <= CurrentValue.[下单时间]`)
	for {
		hasNext, v, err := iter.Next()
		if err != nil {
			panic(err)
		}

		if !hasNext {
			break
		}

		orderId := *v.StringField("订单号")
		if v.StringField("实际情况") != nil {
			actualState := *v.StringField("实际情况")
			actuallyDelivered := actualState == "单次自提" || actualState == "单次配送" || actualState == "7日包-配送" || actualState == "7日包-自提" || actualState == "投妥/已自提"
			if r := orders[orderId].Status; r == 20 && actuallyDelivered {
				wx.DeliverOrder(orderId)
				updated++
			}
		}
	}

	return updated
}

func main() {
	// configure clients and handlers
	feishu.InitClient(os.Getenv("FEISHU_APPID"), os.Getenv("FEISHU_APPSECRET"))
	wx.InitClient(os.Getenv("WEIXIN_APPID"), os.Getenv("WEIXIN_APPSECRET"))

	tableId, appToken = os.Getenv("FEISHU_TABLEID"), os.Getenv("FEISHU_APPTOKEN")

	handler := dispatcher.NewEventDispatcher(os.Getenv("FEISHU_VERIFICATION"), os.Getenv("FEISHU_EVENTENCCODE")).
		OnP2BotMenuV6(func(ctx context.Context, event *larkapplication.P2BotMenuV6) error {
			feishu.SendTextMessage("open_id", *event.Event.Operator.OperatorId.OpenId, "正在处理……")
			if *event.Event.EventKey == "pullMinishopOrders" {
				go func() {
					created, updated := pullMinishopOrders()
					feishu.SendTextMessage("open_id", *event.Event.Operator.OperatorId.OpenId, fmt.Sprint("同步完成，创建", created, "条，同步", updated, "条"))
				}()
			}

			if *event.Event.EventKey == "pushShippingInfo" {
				go func() {
					updated := pushShippingInfo()
					feishu.SendTextMessage("open_id", *event.Event.Operator.OperatorId.OpenId, fmt.Sprint("同步完成，同步", updated, "条"))
				}()
			}

			if *event.Event.EventKey == "setDelivered" {
				go func() {
					updated := setDelivered()
					feishu.SendTextMessage("open_id", *event.Event.Operator.OperatorId.OpenId, fmt.Sprint("同步完成，同步", updated, "条"))
				}()
			}
			return nil
		})

	// register handlers
	http.HandleFunc("/feishu/event", httpserverext.NewEventHandlerFunc(handler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))
	http.HandleFunc("/wx/shippingInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req struct {
			TransactionId string `json:"transactionId"`
			LogisticsType int    `json:"logisticsType"`
			ItemName      string `json:"itemName"`
		}
		err = json.Unmarshal(buf, &req)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = wx.UploadShippingInfo(req.TransactionId, req.LogisticsType, req.ItemName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// pushShippingInfo()
	// pullMinishopOrders()
	// fmt.Println(setDelivered())
	// listen and serve
	err := http.ListenAndServe(os.Getenv("LISTEN_ADDR"), nil)
	if err != nil {
		panic(err)
	}
}
