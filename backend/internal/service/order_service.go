package service

import (
	"errors"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
)

// OrderService builds and persists customer orders. It pulls authoritative
// prices/stock from the product repo, validates promo codes via the promo repo,
// and writes the resulting order through the order repo.
type OrderService struct {
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
	promoRepo   *repository.PromoRepo
}

// NewOrderService wires the order service to the repositories it depends on.
func NewOrderService(o *repository.OrderRepo, p *repository.ProductRepo, pr *repository.PromoRepo) *OrderService {
	return &OrderService{orderRepo: o, productRepo: p, promoRepo: pr}
}

// CreateOrderRequest is the client payload for placing an order: the shipping
// address, an optional promo code, and the line items (size + quantity).
type CreateOrderRequest struct {
	AddressID      int64  `json:"address_id"`
	PromoCode      string `json:"promo_code"`
	DeliveryMethod string `json:"delivery_method"` // courier | post | pickup
	PaymentMethod  string `json:"payment_method"`  // card_online | on_delivery
	RecipientName  string `json:"recipient_name"`
	Items          []struct {
		ProductSizeID int64 `json:"product_size_id"`
		Quantity      int   `json:"quantity"`
	} `json:"items"`
}

// deliveryCost maps a delivery method to its fixed price (rubles). Resolved on the
// server so the shipping fee can't be tampered with by the client.
func deliveryCost(method string) float64 {
	switch method {
	case "courier":
		return 500
	case "post":
		return 350
	default: // pickup or unspecified
		return 0
	}
}

// Create validates and persists a new order for the given user. It recomputes
// the total from server-side prices (never trusting client amounts), applies an
// optional promo discount, and stores the order with its line items.
func (s *OrderService) Create(userID int64, req CreateOrderRequest) (*model.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	orderItems := make([]model.OrderItem, 0, len(req.Items))
	var total float64

	for _, ri := range req.Items {
		// Fetch the authoritative price and verify the requested quantity is in
		// stock for this size; an error here means missing/out-of-stock.
		productID, price, err := s.productRepo.GetProductPriceForSize(ri.ProductSizeID, ri.Quantity)
		if err != nil {
			return nil, errors.New("product or size not available")
		}
		// PriceAtOrder snapshots the price at order time so later price changes
		// don't retroactively alter historical orders.
		orderItems = append(orderItems, model.OrderItem{
			ProductID:     productID,
			ProductSizeID: ri.ProductSizeID,
			Quantity:      ri.Quantity,
			PriceAtOrder:  price,
		})
		// Running subtotal across all line items before any discount.
		total += price * float64(ri.Quantity)
	}

	delivery := deliveryCost(req.DeliveryMethod)
	order := &model.Order{
		UserID:         userID,
		AddressID:      req.AddressID,
		Status:         "pending",
		TotalPrice:     total,
		DiscountAmount: 0,
		DeliveryMethod: req.DeliveryMethod,
		DeliveryCost:   delivery,
		RecipientName:  req.RecipientName,
		PaymentMethod:  req.PaymentMethod,
	}

	// Apply a promo code only when one was supplied. The code must exist and
	// pass validation; otherwise the whole order is rejected.
	if req.PromoCode != "" {
		promo, err := s.promoRepo.GetByCode(req.PromoCode)
		if err != nil {
			return nil, errors.New("promo code not found")
		}
		if err := ValidatePromo(promo); err != nil {
			return nil, err
		}
		// Compute the discount, then subtract it from the subtotal.
		discount := ApplyDiscount(promo, total)
		order.DiscountAmount = discount
		order.TotalPrice = total - discount
		order.PromoCodeID = &promo.ID
	}

	// Shipping is added on top of the (possibly discounted) goods subtotal.
	order.TotalPrice += delivery

	// Payment-on-delivery needs no online step, so the order is accepted immediately;
	// card payments stay pending until the payment stub confirms them.
	if req.PaymentMethod == "on_delivery" {
		order.PaymentStatus = "on_delivery"
		order.Status = "confirmed"
	} else {
		order.PaymentStatus = "unpaid"
		order.Status = "pending"
	}

	if err := s.orderRepo.Create(order, orderItems); err != nil {
		return nil, err
	}
	order.Items = orderItems
	return order, nil
}
