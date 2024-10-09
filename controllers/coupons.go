package controllers

import (
	"context"
	"monk/database"
	"monk/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateCoupons(c *gin.Context) {
	coupon := models.Coupens{}
	if err := c.BindJSON(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not marshal json " + err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := database.Client.CouponCollection.InsertOne(ctx, coupon)
	if res.InsertedID == nil || err != nil {
		c.IndentedJSON(http.StatusBadGateway, "failed to create coupon"+err.Error())
		return
	}
	c.IndentedJSON(200, "Successfully Created coupon")
}

func GetAllCoupons(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rescoupons := []models.Coupens{}
	coupons, err := database.Client.CouponCollection.Find(ctx, bson.M{})
	err = coupons.All(ctx, &rescoupons)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "failed to fetch coupon"+err.Error())
		return
	}
	c.JSON(200, rescoupons)
}

func GetCouponByID(c *gin.Context) {
	couponid, _ := c.Params.Get("ID")
	coupon := models.Coupens{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := database.Client.CouponCollection.FindOne(ctx, bson.M{"_id": couponid}).Decode(&coupon)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "failed to fetch coupon by ID "+couponid+err.Error())
	}
	c.JSON(200, coupon)
}

func DeleteCouponByID(c *gin.Context) {
	couponid, _ := c.Params.Get("ID")
	coupon := models.Coupens{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := database.Client.CouponCollection.DeleteOne(ctx, bson.M{"_id": couponid})
	if err != nil || res.DeletedCount == 0 {
		c.IndentedJSON(http.StatusInternalServerError, "failed to Delete coupon "+couponid+err.Error())
	}
	c.JSON(200, coupon)
}

func FetchAndApplyAllCouponsByCart(c *gin.Context) {
	user, founduser := models.User{}, models.User{}
	rescoupons := []models.Coupens{}
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse json object" + err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	GetUser(ctx, *user.Email)
	res := database.Client.UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
	if res != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user details"})
		return
	}
	productIDs := []string{}
	for _, val := range founduser.UserCart.Products {
		productIDs = append(productIDs, val.ProductID.Hex())
	}
	rescoupon, err := database.Client.CouponCollection.Find(ctx, bson.M{"details.productCoupon.ProductID": bson.M{"$all": productIDs}})
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to fetch coupons for porducts"+err.Error())
		return
	}
	err = rescoupon.All(ctx, &rescoupons)
	cartCoupons := []models.Coupens{}
	rescoupon, err = database.Client.CouponCollection.Find(ctx, bson.M{"type": "cart"})
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to fetch coupons for porducts"+err.Error())
		return
	}
	err = rescoupon.All(ctx, &cartCoupons)
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to parse coupons for products"+err.Error())
		return
	}
	total := 0
	rescoupons = append(rescoupons, cartCoupons...)
	for index, product := range founduser.UserCart.Products {
		for _, coupon := range rescoupons {
			if coupon.Type == "product" && coupon.Details.ProductCoupon.ProductID == product.ProductID {
				founduser.UserCart.Products[index] = ApplyProductCoupon(product, coupon)
				total = total + founduser.UserCart.Products[index].FinalPrice
			} else if coupon.Type == "Cart" {
				founduser.UserCart.TotalPrice = total
				founduser.UserCart = ApplyCartCoupon(founduser.UserCart, coupon)
			} else if coupon.Type == "BxBy" && coupon.Details.ProductCoupon.ProductID == product.ProductID {
				founduser.UserCart.Products[index] = ApplyBXYXCoupon(founduser.UserCart.Products[index], coupon)
				total = total + founduser.UserCart.Products[index].FinalPrice
			}

		}
	}

	_, err = database.Client.UserCollection.UpdateByID(ctx, founduser.ID, founduser)
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to updated cart with coupon discounts")
		return
	}
	c.JSON(200, founduser.UserCart)

}

func GetUser(ctx context.Context, email string) models.User {
	founduser := models.User{}
	res := database.Client.UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&founduser)
	if res != nil {
		return models.User{}
	}
	return founduser
}

func GetCouponsForProductsOrUser(ctx context.Context, productIds []string, id string) []models.Coupens {
	rescoupons := []models.Coupens{}
	opts := options.Find().SetSort(bson.D{{"type", -1}})

	rescoupon, _ := database.Client.CouponCollection.Find(ctx, bson.M{id: bson.M{"$all": productIds}}, opts)
	_ = rescoupon.All(ctx, &rescoupons)
	return rescoupons
}

func ApplyCouponsToCart(c *gin.Context) {
	user, founduser := models.User{}, models.User{}
	rescoupons := []models.Coupens{}
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	founduser = GetUser(ctx, *user.Email)
	productIDs := []string{}
	for _, val := range founduser.UserCart.Products {
		productIDs = append(productIDs, val.ProductID.Hex())
	}
	coupons := GetCouponsForProductsOrUser(ctx, productIDs, "pid")
	rescoupons = append(rescoupons, coupons...)
	total := 0
	for index, product := range founduser.UserCart.Products {
		for _, coupon := range rescoupons {
			if coupon.Type == "product" && coupon.Details.ProductCoupon.ProductID == product.ProductID {
				founduser.UserCart.Products[index] = ApplyProductCoupon(product, coupon)
				total = total + founduser.UserCart.Products[index].FinalPrice
			} else if coupon.Type == "Cart" {
				founduser.UserCart.TotalPrice = total
				founduser.UserCart = ApplyCartCoupon(founduser.UserCart, coupon)
			} else if coupon.Type == "BxBy" && coupon.Details.ProductCoupon.ProductID == product.ProductID {
				founduser.UserCart.Products[index] = ApplyBXYXCoupon(founduser.UserCart.Products[index], coupon)
				total = total + founduser.UserCart.Products[index].FinalPrice
			}

		}
	}

	_, err := database.Client.UserCollection.UpdateByID(ctx, founduser.ID, founduser)
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to updated cart with coupon discounts")
		return
	}
	c.JSON(200, founduser.UserCart)
}

func RemoveCouponFromCart(c *gin.Context) {
	user := models.User{}
	coupon := models.Coupens{}
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	founduser := GetUser(ctx, *user.Email)
	couponid := ""
	res, err := database.Client.CouponCollection.Find(ctx, bson.M{"_id": couponid})
	if err == nil {

	}
	err = res.All(ctx, &coupon)

	if coupon.Type == "Cart" {
		RemoveCartCoupon(founduser.UserCart, coupon)

	} else {
		for index, product := range founduser.UserCart.Products {

			if product.ProductID == coupon.Details.ProductCoupon.ProductID {
				if coupon.Type == "product" {
					founduser.UserCart.Products[index] = RemoveProductCoupon(product, coupon)
				} else if product.ProductID == coupon.Details.BxGyCoupon.ProductID {
					founduser.UserCart.Products[index] = RemoveBXYXCoupon(product, coupon)
				}
			}
		}
	}

	_, err = database.Client.UserCollection.UpdateByID(ctx, founduser.ID, founduser)
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to remove coupons from cart")
		return
	}
	c.JSON(200, "")
}

func ApplyCouponFromCart(c *gin.Context) {
	coupon := models.Coupens{}
	userEmail, _ := c.Get("email")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	founduser := GetUser(ctx, userEmail.(string))
	couponid, _ := c.Params.Get("ID")
	res, err := database.Client.CouponCollection.Find(ctx, bson.M{"_id": couponid})
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to fetch coupon invalid coupon id")
		return
	}
	err = res.All(ctx, &coupon)

	if coupon.Type == "Cart" {
		ApplyCartCoupon(founduser.UserCart, coupon)

	} else {
		for index, product := range founduser.UserCart.Products {

			if product.ProductID == coupon.Details.ProductCoupon.ProductID {
				if coupon.Type == "product" {
					founduser.UserCart.Products[index] = ApplyProductCoupon(product, coupon)
				} else if product.ProductID == coupon.Details.BxGyCoupon.ProductID {
					founduser.UserCart.Products[index] = ApplyBXYXCoupon(product, coupon)
				}
			}
		}
	}
	_, err = database.Client.UserCollection.UpdateByID(ctx, founduser.ID, founduser)
	if err != nil {
		c.JSON(http.StatusBadGateway, "failed to apply coupon to cart")
		return
	}
	c.JSON(200, "coupon updated in cart")
}
func ApplyProductCoupon(product models.ProductUser, coupon models.Coupens) models.ProductUser {
	discount := 0
	if coupon.Details.ProductCoupon.Quantity != 0 && product.Quantity < coupon.Details.ProductCoupon.Quantity {
		return product
	} else if coupon.Details.ProductCoupon.Discount != 0 {
		discount = (*(product.TotalPrice) / 100) * coupon.Details.ProductCoupon.Discount
		product.FinalPrice = discount
		return product
	} else {
		discount := (*product.TotalPrice) - coupon.Details.Cartcoupon.DiscountAmount
		product.FinalPrice = discount
	}
	product.CouponApplied = coupon.ID

	return product
}

func ApplyCartCoupon(cart models.Cart, coupons models.Coupens) models.Cart {

	if cart.TotalPrice >= coupons.Details.Cartcoupon.CartTotal {
		discount := (cart.TotalPrice / 100 * coupons.Details.Cartcoupon.Discount)
		cart.CouponDiscount = discount
		cart.TotalPrice = cart.TotalPrice - discount
		cart.Coupon = coupons.ID
	}

	return cart

}

func ApplyBXYXCoupon(product models.ProductUser, coupon models.Coupens) models.ProductUser {
	product.ActualQuantity = product.Quantity
	rep := coupon.Details.BxGyCoupon.Reptitions
	quantity := product.Quantity % coupon.Details.BxGyCoupon.MinQuantity
	if quantity > rep {
		quantity = rep
	}
	product.Quantity = quantity * coupon.Details.BxGyCoupon.Quantity
	product.CouponApplied = coupon.ID
	return product

}

func RemoveProductCoupon(product models.ProductUser, coupon models.Coupens) models.ProductUser {
	product.FinalPrice = 0
	product.CouponApplied = primitive.ObjectID{}
	return product

}

func RemoveCartCoupon(cart models.Cart, coupon models.Coupens) models.Cart {

	cart.TotalPrice = cart.TotalPrice + cart.CouponDiscount
	cart.Coupon = primitive.ObjectID{}
	cart.CouponDiscount = 0
	return cart

}

func RemoveBXYXCoupon(product models.ProductUser, coupon models.Coupens) models.ProductUser {
	product.Quantity = product.ActualQuantity
	product.CouponApplied = primitive.ObjectID{}
	return product
}
