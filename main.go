package main

import (
	"monk/controllers"
	"monk/middleware"
	"monk/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.Userroutes(router)
	router.Use(middleware.Authentication)
	router.GET("/user", controllers.GetUserDetails)
	router.POST("/user/address", controllers.AddAddress)
	router.POST("/user/address/work", controllers.EditWorkAddress)
	router.POST("/user/address/home", controllers.EditHomeAddress)
	router.DELETE("/user/address/work", controllers.DeleteWorkAddress)
	router.DELETE("/user/address/home", controllers.DeleteHomeAddress)
	router.POST("/Product", controllers.CreateProduct)
	router.GET("/Product", controllers.GetAllProduct)
	router.GET("/Product/:ID", controllers.GetProduct)
	router.DELETE("/Product/:ID", controllers.DeleteProduct)
	router.POST("/Product/:ID", controllers.UpdateProduct)
	router.POST("/Cart", controllers.AddToCart)
	router.DELETE("/Cart", controllers.RemoveFromCart)
	router.POST("/Cart/Checkout", controllers.CheckOutCart)
	router.GET("/Cart", controllers.GetItemsFromCart)
	router.GET("/Order", controllers.GetOrders)
	router.GET("/Order/User", controllers.GetOrdersForUser)
	router.POST("/coupons", controllers.CreateCoupons)
	router.GET("/coupons", controllers.GetAllCoupons)
	router.GET("/coupons/:ID", controllers.GetCouponByID)
	router.DELETE("/Coupon/:ID", controllers.DeleteCouponByID)
	router.PUT("/Coupon/:ID", controllers.UpdateCouponByID)
	router.POST("/Cart/applicable-coupons", controllers.FetchAndApplyAllCouponsByCart)
	router.POST("/Cart/Coupon", controllers.ApplyCouponsToCart)
	router.POST("/apply-coupon/:ID", controllers.ApplyCouponFromCart)
	router.DELETE("/Cart/Coupon", controllers.RemoveCouponFromCart)
	router.Run(":8081")

}
