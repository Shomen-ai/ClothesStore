package service

import (
	"errors"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
	promoRepo   *repository.PromoRepo
}

func NewOrderService(o *repository.OrderRepo, p *repository.ProductRepo, pr *repository.PromoRepo) *OrderService {
	return &OrderService{orderRepo: o, productRepo: p, promoRepo: pr}
}

type CreateOrderRequest struct {
	AddressID int64  `json:"address_id"`
	PromoCode string `json:"promo_code"`
	Items     []struct {
		ProductSizeID int64 `json:"product_size_id"`
		Quantity      int   `json:"quantity"`
	} `json:"items"`
}

func (s *OrderService) Create(userID int64, req CreateOrderRequest) (*model.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	orderItems := make([]model.OrderItem, 0, len(req.Items))
	var total float64

	for _, ri := range req.Items {
		productID, price, err := s.productRepo.GetProductPriceForSize(ri.ProductSizeID, ri.Quantity)
		if err != nil {
			return nil, errors.New("product or size not available")
		}
		orderItems = append(orderItems, model.OrderItem{
			ProductID:     productID,
			ProductSizeID: ri.ProductSizeID,
			Quantity:      ri.Quantity,
			PriceAtOrder:  price,
		})
		total += price * float64(ri.Quantity)
	}

	order := &model.Order{
		UserID:         userID,
		AddressID:      req.AddressID,
		Status:         "pending",
		TotalPrice:     total,
		DiscountAmount: 0,
	}

	if req.PromoCode != "" {
		promo, err := s.promoRepo.GetByCode(req.PromoCode)
		if err != nil {
			return nil, errors.New("promo code not found")
		}
		if err := ValidatePromo(promo); err != nil {
			return nil, err
		}
		discount := ApplyDiscount(promo, total)
		order.DiscountAmount = discount
		order.TotalPrice = total - discount
		order.PromoCodeID = &promo.ID
	}

	if err := s.orderRepo.Create(order, orderItems); err != nil {
		return nil, err
	}
	order.Items = orderItems
	return order, nil
}
