package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      *string            `validate:"required,min=2,max=30"`
	LastName       *string            `validate:"required,min=2,max=30"`
	Password       *string            `validate:"required,min=6"`
	Email          *string            `validate:"email,required"`
	Phone          *string            `validate:"required"`
	Token          *string
	RefershToken   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UserID         string
	UserCart       Cart
	AddressDetails []Address
	OrderDetails   []Order
}
type Cart struct {
	Products       []ProductUser
	TotalPrice     int
	Discount       int
	FinalPrice     int
	Coupon         primitive.ObjectID
	CouponDiscount int
}
type Product struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	ProductName *string
	Price       int
	Rating      int
	Image       *string
}

type ProductUser struct {
	ProductID      primitive.ObjectID `json:"_id" bson:"_id"`
	ProductName    *string
	TotalPrice     *int
	CouponApplied  primitive.ObjectID
	ActualQuantity int
	FinalPrice     int
	Quantity       int
}

type Address struct {
	House   *string
	Street  *string
	City    *string
	Pincode *string
	Phone   *string
}

type Order struct {
	OrderID       primitive.ObjectID `json:"_id" bson:"_id"`
	OrderCart     []ProductUser
	OrderedAT     time.Time
	Price         int
	PaymentMethod Payment
	PlacedBy      string
}

type Payment struct {
	Digital bool
	COD     bool
}

type Coupens struct {
	ID          primitive.ObjectID
	Type        string    `validate:"required"`
	Details     Details   `validate:"required"`
	Expire      time.Time `validate:"required"`
	Description string    `validate:"required"`
}

type Details struct {
	Cartcoupon    Cartcoupon
	ProductCoupon ProductCoupon
	BxGyCoupon    BxGyCoupon
}

type Cartcoupon struct {
	CartTotal      int
	Type           string
	Discount       int
	DiscountAmount int
}

type ProductCoupon struct {
	ProductID primitive.ObjectID
	Discount  int
	Type      string
	Quantity  int
	Amount    int
}

type BxGyCoupon struct {
	ProductID   primitive.ObjectID
	MinQuantity int
	Quantity    int
	Type        string
	Reptitions  int
}
