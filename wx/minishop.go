package wx

import (
	"encoding/json"
	"strconv"
	"time"
)

type Order struct {
	OrderID     int64  `json:"order_id"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
	Status      int    `json:"status"`
	OrderDetail struct {
		ProductInfos []struct {
			ProductID int    `json:"product_id"`
			SkuID     int    `json:"sku_id"`
			ThumbImg  string `json:"thumb_img"`
			SalePrice int    `json:"sale_price"`
			SkuCnt    int    `json:"sku_cnt"`
			Title     string `json:"title"`
			SkuAttrs  []struct {
				AttrKey   string `json:"attr_key"`
				AttrValue string `json:"attr_value"`
			} `json:"sku_attrs"`
			OnAftersaleSkuCnt     int    `json:"on_aftersale_sku_cnt"`
			FinishAftersaleSkuCnt int    `json:"finish_aftersale_sku_cnt"`
			SkuCode               string `json:"sku_code"`
			MarketPrice           int    `json:"market_price"`
		} `json:"product_infos"`
		PayInfo struct {
			PayMethod     string `json:"pay_method"`
			PrepayID      string `json:"prepay_id"`
			PrepayTime    string `json:"prepay_time"`
			PayTime       string `json:"pay_time"`
			TransactionID string `json:"transaction_id"`
		} `json:"pay_info"`
		PriceInfo struct {
			ProductPrice int `json:"product_price"`
			OrderPrice   int `json:"order_price"`
			Freight      int `json:"freight"`
		} `json:"price_info"`
		DeliveryInfo struct {
			AddressInfo struct {
				UserName     string `json:"user_name"`
				PostalCode   string `json:"postal_code"`
				ProvinceName string `json:"province_name"`
				CityName     string `json:"city_name"`
				CountyName   string `json:"county_name"`
				DetailInfo   string `json:"detail_info"`
				NationalCode string `json:"national_code"`
				TelNumber    string `json:"tel_number"`
			} `json:"address_info"`
			DeliveryMethod string `json:"delivery_method"`
			ExpressFee     []struct {
				ShippingMethod string `json:"shipping_method"`
			} `json:"express_fee"`
			DeliveryProductInfo []any `json:"delivery_product_info"`
			OfflineDeliveryTime int   `json:"offline_delivery_time"`
			OfflinePickupTime   int   `json:"offline_pickup_time"`
		} `json:"delivery_info"`
		CouponInfo struct {
			CouponID []any `json:"coupon_id"`
		} `json:"coupon_info"`
	} `json:"order_detail"`
	AftersaleDetail struct {
		AftersaleOrderList  []any `json:"aftersale_order_list"`
		OnAftersaleOrderCnt int   `json:"on_aftersale_order_cnt"`
	} `json:"aftersale_detail"`
	Openid  string `json:"openid"`
	ExtInfo struct {
		CustomerNotes string `json:"customer_notes"`
		MerchantNotes string `json:"merchant_notes"`
	} `json:"ext_info"`
	OrderType int `json:"order_type"`
}

type ListOrdersResponse struct {
	Errcode  int     `json:"errcode"`
	Orders   []Order `json:"orders"`
	TotalNum int     `json:"total_num"`
}

func ListOrders(pastDays int) map[string]Order {
	const timeFormat = "2006-01-02 15:04:05"
	token := GetStableAccessToken()

	t := time.Now()
	page := 1
	body := map[string]interface{}{
		"start_create_time": t.Add(time.Hour * 24 * -time.Duration(pastDays)).Format(timeFormat),
		"end_create_time":   t.Format(timeFormat),
		// "status":            0,
		"page":      page,
		"page_size": 50,
		"source":    1,
	}

	r := map[string]Order{}
	for page := 1; ; page++ {
		body["page"] = page
		buf := wxPostJSON("https://api.weixin.qq.com/product/order/get_list?access_token="+token, body)

		var data ListOrdersResponse
		err := json.Unmarshal(buf, &data)
		if err != nil {
			panic(err)
		}

		if len(data.Orders) == 0 || data.TotalNum == len(r) {
			break
		}

		for _, v := range data.Orders {
			r[strconv.FormatInt(v.OrderID, 10)] = v
		}
	}

	return r
}
