package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"test1/mongo"
	"time"
)

func main() {
	mongo.GetMongoManager().Init("mongodb://127.0.0.1:27017", "", "", "rts")
	GetBuyManager().Init()
	http.HandleFunc("/buy", handle_buy)
	http.HandleFunc("/list", handle_list)
	http.HandleFunc("/orders", handle_list_orders)
	http.ListenAndServe(":8000", nil)
}

type buy_param struct {
	ItemID int64 `json:"itemID" bson:"itemID"`
	Count  int32 `json:"count" bson:"count"`
}

func test() {
	coll := mongo.GetMongoManager().GetCollection("item")
	if coll != nil {
		item := &Item{
			ItemID:    GenUID64(1),
			ItemName:  "火龙果",
			ItemCount: 100,
			Price:     7,
		}
		coll.InsertOne(context.Background(), item)
	}
}
func handle_buy(w http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	c := &buy_param{}
	err = json.Unmarshal(body, c)
	if err != nil {
		return
	}
	if c.Count > 0 {
		bt := GetBuyManager().Buy(c.ItemID, c.Count)
		timer := time.NewTimer(time.Duration(time.Millisecond * 900))
		select {
		case order := <-bt.Order:
			if order.ErrorCode != 0 {
				CommonResult(w, "1", "购买失败")
			} else {
				ResultSuccess(w, order)
			}
			return
		case <-timer.C:
			CommonResult(w, "1", "购买超时")
			return
		}

	} else {
		CommonResult(w, "1", "数量必须大于0")
	}
}
func handle_list(w http.ResponseWriter, request *http.Request) {
	coll := mongo.GetMongoManager().GetCollection("item")
	items := make([]*Item, 0)
	if coll != nil {
		filter := bson.D{}
		cursor, err := coll.Find(context.Background(), filter)
		if err != nil {
			return
		}
		for cursor.Next(context.Background()) {
			u := &Item{}
			err := cursor.Decode(u)
			if err == nil {
				items = append(items, u)
			}
		}
	}
	ResultSuccess(w, items)
}
func handle_list_orders(w http.ResponseWriter, request *http.Request) {
	coll := mongo.GetMongoManager().GetCollection("order")
	items := make([]*map[string]interface{}, 0)
	if coll != nil {
		filter := bson.D{}
		cursor, err := coll.Find(context.Background(), filter)
		if err != nil {
			return
		}
		for cursor.Next(context.Background()) {
			u := make(map[string]interface{})
			err := cursor.Decode(u)
			if err == nil {
				items = append(items, &u)
			}
		}
	}
	ResultSuccess(w, items)
}

type b_ret struct {
	Param int32 `json:"param" bson:"param"`
}

func request_serviceB(param int32) int32 {
	//payload := make(map[string]interface{})
	//header := make(map[string]string)
	//payload["param"] = param
	//params := url.Values{}
	//geturl := "http://127.0.0.1:8001/test"
	//data, err := post("PUT", geturl, &params, payload, header)
	//if err != nil {
	//	return -1
	//}
	//ret := &a_param{}
	//err = json.Unmarshal(data, &ret)
	//if err != nil {
	//	return -1
	//}
	//return ret.Param
	return 0
}
