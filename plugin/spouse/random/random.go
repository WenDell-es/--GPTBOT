package random

import (
	"gptbot/plugin/spouse/model"
	"math"
	"math/rand"
)

// 定义一个减少比例和增加比例，可以根据需要调整
const (
	decreaseRatio = 0.5
	increaseRatio = 0.1
)

func GetRandomCard(cards []model.Card, weight map[string]float64) model.Card {
	n := draw(cards, weight)
	update(cards, weight, n)
	return cards[n]
}

func draw(cards []model.Card, weight map[string]float64) int {
	// 计算所有卡片权重的总和
	total := 0.0
	for i := 0; i < len(cards); i++ {
		if _, ok := weight[cards[i].Name]; !ok {
			weight[cards[i].Name] = 1
		}
		total += weight[cards[i].Name]
	}
	// 防止数值膨胀
	if total*(1+increaseRatio) > math.MaxFloat64/2 {
		for i := 0; i < len(cards); i++ {
			weight[cards[i].Name] /= 100
		}
		total /= 100
	}
	for {
		// 生成一个0到total之间的随机数
		r := rand.Float64() * total
		// 遍历所有卡片，找到第一个使得累计权重大于等于r的卡片
		for i := 0; i < len(cards); i++ {
			// 累计权重
			r -= weight[cards[i].Name]
			// 如果累计权重小于0，说明找到了目标卡片
			if r <= 0 {
				// 返回卡片编号
				return i
			}
		}
	}
}

func update(cards []model.Card, weight map[string]float64, i int) {

	// 把被抽中的卡片权重乘以(1 - 减少比例)
	weight[cards[i].Name] *= 1 - decreaseRatio
	// 把其他没有被抽中的卡片权重乘以(1 + 增加比例)
	for j := 0; j < len(cards); j++ {
		if j != i {
			weight[cards[j].Name] *= 1 + increaseRatio
		}
	}
}
