package handler

import (
	"encoding/json"
	"io/ioutil"

	"github.com/choobot/choo-pos-backend/app/model"
)

type PromotionHandler interface {
	CalculateDiscount(order *model.Order, productsMap map[string]model.Product) (*model.Order, error)
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
	for _, promotion := range promotions.Promotions {
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
				// Update item price in Cart
				for i, item := range order.Items {
					if _, ok := discountedProductsMap[item.Product.Id]; ok {
						order.Items[i].Price = (1 - discountPercent) * discountedProductsMap[item.Product.Id].Price
						delete(discountedProductsMap, item.Product.Id)
					}
				}
			}
		}
	}

	return order, nil
}
