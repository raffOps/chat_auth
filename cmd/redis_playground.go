package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/raffops/chat/pkg/database/redis"
	"log"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("%v", err)
	}
	ctx := context.Background()

	rdb := redis.GetRedisConn(ctx)

	err = rdb.Set(ctx, "teste2", "teste", 0).Err()
	if err != nil {
		log.Fatalf("%v", err)
	}

	//hashMap := map[string]interface{}{
	//	"field3": "value3",
	//	"field4": "value4",
	//}
	//for k, v := range hashMap {
	//	err = rdb.HSet(ctx, "sessionId:12233434", k, v).Err()
	//	if err != nil {
	//		log.Println(err)
	//	}
	//}

	//rdb.SAdd(ctx, "set1", "value1", "value2", "value3")
}
