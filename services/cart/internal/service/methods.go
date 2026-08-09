package service

import (
	"context"
	"strings"
	"time"
)

type Cart struct {
	CartID    string
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
	ProductID      string
	MerchantID     string
	Quantity       int32
	ProductName    string
	UnitPriceCents int64
}

func (s *Service) GetCart(ctx context.Context, cartID string) (Cart, error) {
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}

	cart, ok, err := s.loadCart(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		cart = newCart(cartID)
	}
	return cart, nil
}

func (s *Service) AddCartItem(ctx context.Context, cartID, productID, merchantID, productName string, quantity int32, unitPriceCents int64) (Cart, error) {
	if quantity <= 0 {
		return Cart{}, ErrInvalidQuantity
	}
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validateUUID(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}
	if err := validateUUID(merchantID, ErrInvalidMerchantID); err != nil {
		return Cart{}, err
	}
	if err := validateSnapshot(productName, unitPriceCents); err != nil {
		return Cart{}, err
	}

	cart, ok, err := s.loadCart(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		cart = newCart(cartID)
	}
	idx := findCartItem(cart.Items, productID)
	if idx >= 0 {
		cart.Items[idx].MerchantID = merchantID
		cart.Items[idx].Quantity += quantity
		cart.Items[idx].ProductName = productName
		cart.Items[idx].UnitPriceCents = unitPriceCents
	} else {
		cart.Items = append(cart.Items, CartItem{
			ProductID:      productID,
			MerchantID:     merchantID,
			Quantity:       quantity,
			ProductName:    productName,
			UnitPriceCents: unitPriceCents,
		})
	}
	cart.UpdatedAt = time.Now().UTC()
	if err := s.saveCart(ctx, cart); err != nil {
		return Cart{}, err
	}
	return cart, nil
}

func (s *Service) SetCartItemQuantity(ctx context.Context, cartID, productID, merchantID, productName string, quantity int32, unitPriceCents int64) (Cart, error) {
	if quantity <= 0 {
		return s.RemoveCartItem(ctx, cartID, productID)
	}
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validateUUID(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}
	if err := validateUUID(merchantID, ErrInvalidMerchantID); err != nil {
		return Cart{}, err
	}
	if err := validateSnapshot(productName, unitPriceCents); err != nil {
		return Cart{}, err
	}

	cart, ok, err := s.loadCart(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		cart = newCart(cartID)
	}
	idx := findCartItem(cart.Items, productID)
	item := CartItem{
		ProductID:      productID,
		MerchantID:     merchantID,
		Quantity:       quantity,
		ProductName:    productName,
		UnitPriceCents: unitPriceCents,
	}
	if idx >= 0 {
		cart.Items[idx] = item
	} else {
		cart.Items = append(cart.Items, item)
	}
	cart.UpdatedAt = time.Now().UTC()
	if err := s.saveCart(ctx, cart); err != nil {
		return Cart{}, err
	}
	return cart, nil
}

func (s *Service) RemoveCartItem(ctx context.Context, cartID, productID string) (Cart, error) {
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validateUUID(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}

	cart, ok, err := s.loadCart(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		return Cart{}, ErrItemNotFound
	}
	idx := findCartItem(cart.Items, productID)
	if idx < 0 {
		return Cart{}, ErrItemNotFound
	}
	cart.Items = append(cart.Items[:idx], cart.Items[idx+1:]...)
	cart.UpdatedAt = time.Now().UTC()
	if err := s.saveCart(ctx, cart); err != nil {
		return Cart{}, err
	}
	return cart, nil
}

func (s *Service) RemoveCartItems(ctx context.Context, cartID string, productIDs []string) (Cart, error) {
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if len(productIDs) == 0 {
		return Cart{}, ErrEmptyProductIDs
	}
	for _, productID := range productIDs {
		if err := validateUUID(productID, ErrInvalidProductID); err != nil {
			return Cart{}, err
		}
	}

	cart, ok, err := s.loadCart(ctx, cartID)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		return newCart(cartID), nil
	}

	remove := make(map[string]struct{}, len(productIDs))
	for _, productID := range productIDs {
		remove[productID] = struct{}{}
	}
	kept := make([]CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		if _, drop := remove[item.ProductID]; drop {
			continue
		}
		kept = append(kept, item)
	}
	cart.Items = kept
	cart.UpdatedAt = time.Now().UTC()
	if err := s.saveCart(ctx, cart); err != nil {
		return Cart{}, err
	}
	return cart, nil
}

func (s *Service) ClearCart(ctx context.Context, cartID string) error {
	if err := validateUUID(cartID, ErrInvalidCartID); err != nil {
		return err
	}
	return s.deleteCart(ctx, cartID)
}

func validateSnapshot(productName string, unitPriceCents int64) error {
	if strings.TrimSpace(productName) == "" || unitPriceCents <= 0 {
		return ErrInvalidSnapshot
	}
	return nil
}
