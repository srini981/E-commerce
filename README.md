For each and Every api, have implemted each every case that i thought of , to ensure the full working of the system have enabled other apis(products,users and cart) which were required for the coupons api to function as well 
To ensure proper authention have used a middleware which utilized JWT tokens to enable authentication

coupons API:
CreateCoupons api :
cases:
1. json body is not properly parsered 
2.Validation of the required fields
3.if we cant generate a coupon based on the json body

GetAllCoupons api:
cases:
only if we can fetch the coupons from the db as it a get request

GetCouponsByID api:
cases:
if we cant fetch the id from the path thrown an error
if we cant find the coupon in the database thrown an error

delete coupon api:
cases:
if we cant fetch the id from the path thrown an error
if we cant delete the coupon in the database thrown an error


FetchAndApplyAllCouponsByCart api:
if we cant fetch the email id of the user thrown an error
if we cant find the coupons on the product avaliable in the cart thrown an error 
as the above case only checks for product coupons and BXYX coupons , if there  we thereis an error in finding cart based coupons thrown an error
and after finding the coupons if there is any error in parsing those coupons have thrown an error
after applying the coupons if i cant update the cart details , have thrown the error

ApplyCouponFromCart api:
cases:
if we cant fetch the email id of the user thrown an error
if we cant find the coupons on the product avaliable in the cart thrown an error 
after applying the coupons if i cant update the cart details , have thrown the error

UpdateCoupon api
if we cant fetch the id from the path thrown an error
1. json body is not properly parsered 
after applying the coupons if i cant update the cart details , have thrown the error



Extentions
the current architecture  can be upgraded to micro service architecture for handling huge traffic
Having rbac as separate service can have its own advantages and its offers a great varity of flexibility regarding auth 
and for performance reasons we can implement grpc for communication between microservices
and having a separate db service offers a limit in the connections being sent to db
and Unit tests as my child was hospitalized  i couldt complete the unit test, but it can be done using mock and testing package in golang








