package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"monk/database"
	"monk/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddToCart(c *gin.Context) {

	var productcart []models.ProductUser

	email, _ := c.Get("email")
	if email == "" {
		log.Println("user email is empty")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
		return
	}

	if err := c.BindJSON(&productcart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{}
	filter := bson.D{primitive.E{Key: "email", Value: email}}

	err := database.Client.UserCollection.FindOne(ctx, filter).Decode(&user)
	for _, products := range productcart {
		count := 0
		for index, product := range user.UserCart.Products {
			if product.ProductID == products.ProductID {
				user.UserCart.Products[index].Quantity += products.Quantity
				count++
			}
		}
		if count == 0 {
			user.UserCart.Products = append(user.UserCart.Products, products)
		}
	}

	_, err = database.Client.UserCollection.ReplaceOne(ctx, filter, user)
	if err != nil {
		return
	}

	c.IndentedJSON(200, "Successfully Added to the cart")
}

func RemoveFromCart(c *gin.Context) {

	var productcart []models.ProductUser

	email, _ := c.Get("email")
	if email == "" {
		log.Println("user email is empty")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
		return
	}
	if err := c.BindJSON(&productcart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{}
	filter := bson.D{primitive.E{Key: "email", Value: email}}

	err := database.Client.UserCollection.FindOne(ctx, filter).Decode(&user)
	for _, products := range productcart {

		for index, product := range user.UserCart.Products {
			if product.ProductID == products.ProductID {
				user.UserCart.Products[index].Quantity -= products.Quantity
				if user.UserCart.Products[index].Quantity <= 0 {
					fmt.Println("index", index)
					remove(user.UserCart.Products, index)
				}
			}
		}
	}
	fmt.Println(user.UserCart)

	_, err = database.Client.UserCollection.ReplaceOne(ctx, filter, user)
	if err != nil {
		return
	}
	c.IndentedJSON(200, "Successfully removed from cart")
}

func remove(slice []models.ProductUser, s int) []models.ProductUser {
	return append(slice[:s], slice[s+1:]...)
}
func CheckOutCart(c *gin.Context) {
	email, _ := c.Get("email")
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid email"})
		c.Abort()
		return
	}
	var filledcart models.User
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}).Decode(&filledcart)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(500, "not id found")
		return
	}
	order := models.Order{}
	order.OrderCart = filledcart.UserCart.Products
	order.OrderedAT = time.Now()
	for _, val := range filledcart.UserCart.Products {
		order.Price = order.Price + (*&val.Quantity * *val.TotalPrice)
	}
	order.PaymentMethod.COD = true
	order.PlacedBy = email.(string)
	ordRes, err := database.Client.OrderCollection.InsertOne(ctx, order)
	if err != nil || ordRes.InsertedID == nil {
		c.JSON(http.StatusBadRequest, "cant place order to checkout cart")
		return
	}
	filledcart.UserCart = models.Cart{}
	_, err = database.Client.UserCollection.UpdateOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}, filledcart)
	if err != nil {
		c.JSON(http.StatusBadRequest, "order placed but the cart is not empty pls clear the cart")
		return
	}
	c.JSON(http.StatusOK, "checked out cart and order is placed")

}

func GetItemsFromCart(c *gin.Context) {
	email, _ := c.Get("email")
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
		c.Abort()
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var filledcart models.User
	err := database.Client.UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}).Decode(&filledcart)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(500, "not id found")
		return
	}
	c.JSON(http.StatusOK, filledcart.UserCart)
	ctx.Done()
}

func GetOrders(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orders []models.Order
	rederRes, err := database.Client.UserCollection.Find(ctx, bson.M{})
	err = rederRes.All(ctx, &orders)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(500, "not id found")
		return
	}
	c.JSON(http.StatusOK, orders)

}
func GetOrdersForUser(c *gin.Context) {
	email, _ := c.Get("email")
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
		c.Abort()
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orders []models.Order
	err := database.Client.OrderCollection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}).Decode(&orders)
	if err != nil {
		c.JSON(http.StatusBadGateway, "unable to retrive orders for  user")
		return
	}
	c.JSON(http.StatusOK, orders)

}
