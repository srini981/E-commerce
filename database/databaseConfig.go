package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client MongoClient = DBSET()

type MongoClient struct {
	Client            *mongo.Client
	UserCollection    *mongo.Collection
	ProductCollection *mongo.Collection
	CouponCollection  *mongo.Collection
	OrderCollection   *mongo.Collection
}

func DBSET() MongoClient {
	fmt.Println("dbcon")
	var mongoclient MongoClient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongoadmin:secret@localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	mongoclient.Client = client
	mongoclient.UserCollection = client.Database("E-commerence").Collection("Users")
	mongoclient.ProductCollection = client.Database("E-commerence").Collection("products")
	mongoclient.CouponCollection = client.Database("E-commerence").Collection("coupons")
	mongoclient.OrderCollection = client.Database("E-commerence").Collection("orders")
	return mongoclient
}
