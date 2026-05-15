package service

import (
	"context"
	"time"
)

type Cart struct {
	CartID    string
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
	ProductID  string
	MerchantID string
	Quantity   int32
}

func (s *Service) GetCart(ctx context.Context, cartID string) (Cart, error) {
	if err := validate(cartID, ErrInvalidCartID); err != nil {
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

func (s *Service) AddCartItem(ctx context.Context, cartID, productID, merchantID string, quantity int32) (Cart, error) {
	if quantity <= 0 {
		return Cart{}, ErrInvalidQuantity
	}
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}
	if err := validate(merchantID, ErrInvalidMerchantID); err != nil {
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
	} else {
		cart.Items = append(cart.Items, CartItem{ProductID: productID, MerchantID: merchantID, Quantity: quantity})
	}
	cart.UpdatedAt = time.Now().UTC()
	if err := s.saveCart(ctx, cart); err != nil {
		return Cart{}, err
	}
	return cart, nil
}

func (s *Service) SetCartItemQuantity(ctx context.Context, cartID, productID, merchantID string, quantity int32) (Cart, error) {
	if quantity <= 0 {
		return s.RemoveCartItem(ctx, cartID, productID)
	}
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}
	if err := validate(merchantID, ErrInvalidMerchantID); err != nil {
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
	item := CartItem{ProductID: productID, MerchantID: merchantID, Quantity: quantity}
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
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
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

func (s *Service) ClearCart(ctx context.Context, cartID string) error {
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return err
	}
	return s.deleteCart(ctx, cartID)
}
