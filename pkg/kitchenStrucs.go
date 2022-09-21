package pkg

import "time"

type KitchenOrder struct {
	OrderBody         Order
	finDishes         int
	ReceivedTime      time.Time
	FullyPreparedTime time.Time
	CookingTime       time.Duration
}

type KitchenDish struct {
	index  int
	parent Order
	Dish
}
