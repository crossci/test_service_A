package main

import (
	"sync/atomic"
	"time"
)

type Item struct {
	ItemID    int64   `json:"itemID" bson:"itemID"`
	ItemName  string  `json:"itemName" bson:"itemName"`
	ItemCount int32   `json:"itemCount" bson:"itemCount"`
	Price     float32 `json:"price" bson:"price"`
}

var index int64 = 0

// 获取64位唯一ID
// 总共19位,前两位是小于90的tag，中间13位位当前的毫秒数，后四位是0-9999
// 线程安全	 92 2337203685477 5807
//	        |--|----毫秒-----|----
func GenUID64(tag int64) int64 {
	if tag < 0 || tag > 90 {
		return 0
	}
	ms := GetMilliSecond() * 10000
	head2 := tag * 100000000000000000
	atomic.AddInt64(&index, 1)
	return head2 + ms + index%10000
}

// 获取当前毫秒
func GetMilliSecond() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}
