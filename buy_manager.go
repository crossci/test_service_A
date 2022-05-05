package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"net/url"
	"test1/mongo"
	"time"
)

type Order struct {
	OrderID   int64
	ItemID    int64
	Count     int32
	CreateAt  string
	ErrorCode int32
}
type BuyTask struct {
	ItemID int64
	Count  int32
	Order  chan *Order
}
type BuyManager struct {
	exit     chan struct{}
	buy_task chan *BuyTask
}

var buyManager *BuyManager

func GetBuyManager() *BuyManager {
	if buyManager == nil {
		buyManager = &BuyManager{}

	}
	return buyManager
}
func (bm *BuyManager) Init() {
	bm.buy_task = make(chan *BuyTask, 1000)
	go bm.run()
}
func (bm *BuyManager) run() {
	for {
		select {
		case data, ok := <-bm.buy_task: //
			if !ok {
				break
			}
			bm.handleBuy(data)
		}
	}
}
func (bm *BuyManager) Buy(itemID int64, count int32) *BuyTask {
	bt := &BuyTask{
		ItemID: itemID,
		Count:  count,
		Order:  make(chan *Order, 1),
	}
	bm.buy_task <- bt
	return bt
}
func (bm *BuyManager) handleBuy(t *BuyTask) {
	item := bm.get_item(t.ItemID)
	if item != nil {
		if item.ItemCount >= t.Count {
			item.ItemCount -= t.Count

			order := &Order{
				OrderID:   GenUID64(2),
				ItemID:    item.ItemID,
				Count:     t.Count,
				CreateAt:  time.Now().Format("2006-01-02 15:04:05"),
				ErrorCode: 0,
			}
			errorCode := request_gen_order(order, item.Price)
			order.ErrorCode = errorCode
			if errorCode == 0 {
				bm.update_item(item)
			}
			t.Order <- order
		} else {
			//数量不足
			order := &Order{
				ErrorCode: 2,
			}
			t.Order <- order
		}
	} else {
		//找不到商品
		order := &Order{
			ErrorCode: 1,
		}
		t.Order <- order
	}
}

func (bm *BuyManager) get_item(itemID int64) *Item {
	filter := bson.D{{"itemID", itemID}}
	coll := mongo.GetMongoManager().GetCollection("item")
	if coll != nil {
		result := coll.FindOne(context.Background(), filter)
		if result.Err() == nil {
			ret := &Item{}
			err := result.Decode(ret)
			if err == nil {
				return ret
			}
		}
	}
	return nil
}
func (bm *BuyManager) update_item(item *Item) {
	filter := bson.D{{"itemID", item.ItemID}}
	coll := mongo.GetMongoManager().GetCollection("item")
	if coll != nil {
		updateMap := make(map[string]interface{})
		updateMap["itemCount"] = item.ItemCount
		update := bson.M{"$set": updateMap}
		coll.UpdateOne(context.Background(), filter, update)
	}
}

var pay_url = "http://127.0.0.1:8001/"

type gen_order_ret struct {
	ErrorCode int32 `json:"code"`
}

func request_gen_order(order *Order, price float32) int32 {
	payload := make(map[string]interface{})
	header := make(map[string]string)
	payload["orderID"] = order.OrderID
	payload["itemID"] = order.ItemID
	payload["count"] = order.Count
	payload["pay"] = float32(order.Count) * price
	params := url.Values{}
	geturl := fmt.Sprintf("%s%s", pay_url, "genOrder")
	data, err := post("PUT", geturl, &params, payload, header)
	if err != nil {
		return -1
	}
	ret := &gen_order_ret{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return -1
	}
	return ret.ErrorCode
}

func post(method, _url string, params *url.Values, payload map[string]interface{}, headers map[string]string) ([]byte, error) {
	Url, err := url.Parse(_url)
	if err != nil {
		return nil, err
	}
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	client := &http.Client{
	}
	bytesData, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, urlPath, bytes.NewReader(bytesData))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %s \n", urlPath, err))
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(" %s %s \n", urlPath, err))
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
