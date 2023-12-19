package wx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func ListOrders() []Order {
	const timeFormat = "2006-01-02 15:04:05"
	token := GetStableAccessToken()

	t := time.Now()
	body, err := json.Marshal(map[string]interface{}{
		"start_create_time": t.Add(time.Hour * 24 * -15).Format(timeFormat),
		"end_create_time":   t.Format(timeFormat),
		// "status":            0,
		"page":      0,
		"page_size": 1000,
		"source":    1,
	})
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("https://api.weixin.qq.com/product/order/get_list?access_token="+token, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data ListOrdersResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data.Errcode, data.TotalNum)
	return data.Orders
}
