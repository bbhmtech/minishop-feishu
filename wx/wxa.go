package wx

import (
	"encoding/json"
	"time"
)

// key: transactionId
// 订单状态枚举：(1) 待发货；(2) 已发货；(3) 确认收货；(4) 交易完成；(5) 已退款。
func ListShippingInfo(pastDays int) map[string]int {
	token := GetStableAccessToken()
	_, begin := daysAgo(pastDays)
	body := map[string]interface{}{
		"pay_time_range": map[string]int64{
			"begin_time": begin.Unix(),
			// "end_time":
		},
		"page_size": 50,
	}

	shippingInfo := map[string]int{}
	for {
		buf := wxPostJSON("https://api.weixin.qq.com/wxa/sec/order/get_order_list?access_token="+token, body)
		var r map[string]interface{}
		err := json.Unmarshal(buf, &r)
		if err != nil {
			panic(err)
		}

		for _, v := range r["order_list"].([]interface{}) {
			tId, state := v.(map[string]interface{})["transaction_id"].(string), v.(map[string]interface{})["order_state"].(float64)
			shippingInfo[tId] = int(state)
		}

		if !r["has_more"].(bool) {
			break
		}

		body["last_index"] = r["last_index"]
		// r["order"].(map[string]interface{})["order_state"]
	}

	return shippingInfo
}

// logisticsType: 1=快递 2=同城配送 3=虚拟商品 4=用户自提
func UploadShippingInfo(transactionId string, logisticsType int, itemName string) error {
	token := GetStableAccessToken()

	buf := wxPostJSON("https://api.weixin.qq.com/wxa/sec/order/get_order?access_token="+token, map[string]interface{}{
		"transaction_id": transactionId,
	})

	var getOrderResp map[string]interface{}
	err := json.Unmarshal(buf, &getOrderResp)
	if err != nil {
		panic(err)
	}

	buf = wxPostJSON("https://api.weixin.qq.com/wxa/sec/order/upload_shipping_info?access_token="+token, map[string]interface{}{
		"order_key": map[string]interface{}{
			"order_number_type": 2,
			"transaction_id":    transactionId,
		},
		"logistics_type": logisticsType,
		"delivery_mode":  1, // UNIFIED_DELIVERY
		"shipping_list": []map[string]interface{}{
			{
				"item_desc": itemName,
			},
		},
		"upload_time": time.Now().In(WeixinLocation).Format(time.RFC3339Nano),
		"payer": map[string]interface{}{
			"openid": getOrderResp["order"].(map[string]interface{})["openid"],
		},
	})

	var uploadShippingInfoResp map[string]interface{}
	err = json.Unmarshal(buf, &uploadShippingInfoResp)
	if err != nil {
		panic(err)
	}

	return nil
}
