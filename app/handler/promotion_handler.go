package handler

import (
	"encoding/json"
	"io/ioutil"

	"github.com/choobot/choo-pos-backend/app/model"
)

type PromotionHandler interface {
	CalculateDiscount(order *model.Order, productsMap map[string]model.Product) (*model.Order, error)
	AddPromotionDetailToProduct(products []model.Product) ([]model.Product, error)
}

type FixPromotionHandler struct {
}

func (this *FixPromotionHandler) CalculateDiscount(order *model.Order, productsMap map[string]model.Product) (*model.Order, error) {

	file, err := ioutil.ReadFile("data/promotions.json")
	if err != nil {
		return nil, err
	}
	var promotions model.Promotions
	err = json.Unmarshal(file, &promotions)
	if err != nil {
		return nil, err
	}
	totalDiscountPerPromotion := map[string]float64{}
	orderItemsPerPromotion := map[string][]model.OrderItem{}
	for _, promotion := range promotions.Promotions {
		totalDiscount := 0.0
		orderItemsPerPromotion[promotion.Id] = make([]model.OrderItem, len(order.Items))
		copy(orderItemsPerPromotion[promotion.Id], order.Items)
		if promotion.Type == "tier_unique_num" {
			num := 0
			discountedProductsMap := map[string]model.Product{}
			for _, productId := range promotion.ProductIds {
				if _, ok := productsMap[productId]; ok {
					num++
					discountedProductsMap[productId] = productsMap[productId]
				}
			}
			if num > len(promotion.Tiers) {
				num = len(promotion.Tiers)
			}
			if num > 0 {
				discountPercent := promotion.Tiers[num-1]
				// Update item price in Order
				for i, item := range orderItemsPerPromotion[promotion.Id] {
					if _, ok := discountedProductsMap[item.Product.Id]; ok {
						discountPrice := (1 - discountPercent) * discountedProductsMap[item.Product.Id].Price
						orderItemsPerPromotion[promotion.Id][i].Price = discountPrice
						totalDiscount += discountPrice
						delete(discountedProductsMap, item.Product.Id)
					}
				}
			}
		} else if promotion.Type == "not_in" {
			discountedProductsMap := map[string]model.Product{}
			for _, productId := range promotion.ProductIds {
				if _, ok := productsMap[productId]; ok {
					discountedProductsMap[productId] = productsMap[productId]
				}
			}
			discountPercent := promotion.Tiers[0]
			// Update item price in Order
			for i, item := range orderItemsPerPromotion[promotion.Id] {
				if _, ok := discountedProductsMap[item.Product.Id]; !ok {
					discountPrice := (1 - discountPercent) * item.Product.Price
					orderItemsPerPromotion[promotion.Id][i].Price = discountPrice
					totalDiscount += discountPrice
				}
			}
		}
		totalDiscountPerPromotion[promotion.Id] = totalDiscount
	}

	// Find max promotion which has max discount
	mostDiscount := 0.0
	mostDiscountPromotionId := ""
	for promotionId, discountPerPromotion := range totalDiscountPerPromotion {
		if discountPerPromotion > mostDiscount {
			mostDiscount = discountPerPromotion
			mostDiscountPromotionId = promotionId
		}
	}

	if mostDiscount != 0 {
		order.Items = orderItemsPerPromotion[mostDiscountPromotionId]
	}

	return order, nil
}

func (this *FixPromotionHandler) contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (this *FixPromotionHandler) AddPromotionDetailToProduct(products []model.Product) ([]model.Product, error) {
	file, err := ioutil.ReadFile("data/promotions.json")
	if err != nil {
		return nil, err
	}
	var promotions model.Promotions
	err = json.Unmarshal(file, &promotions)
	if err != nil {
		return nil, err
	}
	for _, promotion := range promotions.Promotions {
		for i, product := range products {
			if this.contains(promotion.ProductIds, product.Id) {
				products[i].Promotion = promotion.Name
			}
		}
	}

	return products, nil
}
