package controllers

import (
	"context"
	"fmt"
	"log"
	"monk/database"
	"monk/models"
	generate "monk/tokens"
	"net/http"
	"time"

	"github.com/go-playground/validator"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var Validate = validator.New()

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login Or Passowrd is Incorerct"
		valid = false
	}
	return valid, msg
}
func SignUp(c *gin.Context) {

	user := models.User{}
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validationErr := Validate.Struct(user)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	count, countErr := database.Client.UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if countErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": countErr})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is already in use"})
		return
	}

	count, countErr = database.Client.UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	if countErr != nil {
		log.Panic(countErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": countErr})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
		return
	}
	password := hashPassword(*user.Password)
	user.Password = &password
	user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.FirstName, *user.LastName, user.UserID)
	user.Token = &token
	user.RefershToken = &refreshtoken
	user.UserCart = models.Cart{}
	user.AddressDetails = []models.Address{}
	user.OrderDetails = []models.Order{}

	_, insertErr := database.Client.UserCollection.InsertOne(ctx, user)
	if insertErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not created"})
		return
	}
	c.JSON(http.StatusCreated, "Successfully Signed Up!!")
}

func Login(c *gin.Context) {
	user, founduser := models.User{}, models.User{}
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validationErr := Validate.Struct(user)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	res := database.Client.UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
	if res != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "login or password incorrect"})
		return
	}
	passwordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
	if !passwordIsValid {
		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}
	token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.FirstName, *founduser.LastName, founduser.UserID)
	err := generate.UpdateAllTokens(token, refreshToken, founduser.UserID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusAccepted, "User loged successfully")
}

func GetUserDetails(c *gin.Context) {
	email, err := c.Get("email")
	if !err {
		c.JSON(http.StatusBadRequest, "failed to get query param")
		return
	}
	founduser := models.User{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	res := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&founduser)
	if res != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, founduser)
}

func AddAddress(c *gin.Context) {
	email, _ := c.Get("email")
	fmt.Print(" email", email)
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
		c.Abort()
		return
	}
	user := models.User{}
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if len(user.AddressDetails) >= 2 || err != nil {
		c.JSON(http.StatusBadGateway, "couldt add address home address already present")
		return
	}
	user.AddressDetails = append(user.AddressDetails, editaddress)
	fmt.Println(user.AddressDetails)
	res, err := database.Client.UserCollection.ReplaceOne(ctx, bson.M{"email": user.Email}, user)
	fmt.Println(res)
	fmt.Println(err)
	if err != nil || (res != nil && res.ModifiedCount == 0) {
		c.JSON(http.StatusBadGateway, "couldt add address home address ")
		return
	}
	c.JSON(http.StatusOK, "added home address")
}

func EditWorkAddress(c *gin.Context) {
	email, _ := c.Get("email")
	fmt.Print(" email", email)
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
		c.Abort()
		return
	}
	user := models.User{}
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if len(user.AddressDetails) == 1 || err != nil {
		c.JSON(http.StatusBadGateway, "no work address avaliable for user")
		return
	}
	user.AddressDetails[1] = editaddress
	_, err = database.Client.UserCollection.ReplaceOne(ctx, bson.M{"email": user.Email}, user)

	if err != nil {
		c.JSON(http.StatusBadGateway, "couldt modify work address")
		return
	}
	c.JSON(http.StatusOK, "modified work address")
}

func EditHomeAddress(c *gin.Context) {
	email, _ := c.Get("email")
	fmt.Print(" email", email)
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
		c.Abort()
		return
	}
	user := models.User{}
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if len(user.AddressDetails) == 0 || err != nil {
		c.JSON(http.StatusBadGateway, "no home address avaliable for user")
		return
	}
	user.AddressDetails[0] = editaddress
	_, err = database.Client.UserCollection.ReplaceOne(ctx, bson.M{"email": user.Email}, user)

	if err != nil {
		c.JSON(http.StatusBadGateway, "couldt modify home address")
		return
	}
	c.JSON(http.StatusOK, "modified home address")
}

func DeleteHomeAddress(c *gin.Context) {
	email, _ := c.Get("email")
	fmt.Print(" email", email)
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
		c.Abort()
		return
	}
	user := models.User{}
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if len(user.AddressDetails) == 0 || err != nil {
		c.JSON(http.StatusBadGateway, "no home address avaliable for user")
		return
	}
	user.AddressDetails[0] = models.Address{}
	_, err = database.Client.UserCollection.ReplaceOne(ctx, bson.M{"email": user.Email}, user)

	if err != nil {
		c.JSON(http.StatusBadGateway, "couldt modify home address")
		return
	}
	c.JSON(http.StatusOK, "modified home address")
}

func DeleteWorkAddress(c *gin.Context) {
	email, _ := c.Get("email")
	fmt.Print(" email", email)
	if email == "" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid code"})
		c.Abort()
		return
	}
	user := models.User{}
	var editaddress models.Address
	if err := c.BindJSON(&editaddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if len(user.AddressDetails) == 1 || err != nil {
		c.JSON(http.StatusBadGateway, "no work address avaliable for user")
		return
	}
	user.AddressDetails[1] = models.Address{}
	_, err = database.Client.UserCollection.ReplaceOne(ctx, bson.M{"email": user.Email}, user)

	if err != nil {
		c.JSON(http.StatusBadGateway, "couldt modify work address")
		return
	}
	c.JSON(http.StatusOK, "modified work address")
}
