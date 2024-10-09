package controllers

import (
	"context"
	"fmt"
	"monk/database"
	"monk/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateProduct(c *gin.Context) {
	product := models.Product{}
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validationErr := Validate.Struct(product)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}
	product.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_, insertErr := database.Client.ProductCollection.InsertOne(ctx, product)
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}
	c.JSON(http.StatusCreated, "Successfully created product !!")
}

func DeleteProduct(c *gin.Context) {
	id, _ := c.Params.Get("ID")
	objID, _ := primitive.ObjectIDFromHex(id)
	product := models.Product{}
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validationErr := Validate.Struct(product)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}
	product.ID = primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_, insertErr := database.Client.ProductCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}
	c.JSON(http.StatusCreated, "Successfully deleted product !!")

}

func GetAllProduct(c *gin.Context) {
	products := []models.Product{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	productRes, err := database.Client.ProductCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}
	err = productRes.All(ctx, &products)
	c.JSON(http.StatusOK, products)
}

func GetProduct(c *gin.Context) {
	products := models.Product{}
	id, _ := c.Params.Get("ID")
	objID, err := primitive.ObjectIDFromHex(id)
	fmt.Println(id)
	fmt.Println(err)
	fmt.Println(objID)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	productRes := database.Client.ProductCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&products)
	if productRes != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func UpdateProduct(c *gin.Context) {
	product := models.Product{}
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validationErr := Validate.Struct(product)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}
	id, _ := c.Params.Get("ID")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {

	}
	product.ID = objID
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_, insertErr := database.Client.ProductCollection.ReplaceOne(ctx, bson.M{"_id": objID}, product)
	fmt.Println(insertErr)
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
		return
	}
	c.JSON(http.StatusCreated, "Successfully created product !!")

}
