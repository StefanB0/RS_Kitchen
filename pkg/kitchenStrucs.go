package pkg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type KitchenOrder struct {
	OrderBody      Order
	priority       int
	finDishes      int
	ReceivedTime   time.Time
	CookingTime    int
	CookingDetails []struct {
		Cook_ID int
		Food_ID int
	}
}

type KitchenDish struct {
	priority int
	cookID   int
	dish     Dish
	progress int
	parent   *KitchenOrder
	cook     *Cook
}

type Dish struct {
	Id               int
	Name             string
	PreparationTime  int
	Complexity       int
	CookingApparatus string
}

func ReadMenu(path string) []Dish {
	jsonfile, err := os.Open(path)
	defer jsonfile.Close()

	if err != nil {
		log.Println(err)
	}

	bytevalue, _ := ioutil.ReadAll(jsonfile)
	newMenu := []Dish{}
	json.Unmarshal(bytevalue, &newMenu)

	return newMenu
}

func newKitchenOrder(o Order) KitchenOrder {
	return KitchenOrder{
		OrderBody:    o,
		finDishes:    0,
		ReceivedTime: time.Now(),
	}
}

func sortDishComplexity(list []KitchenDish) []KitchenDish {
	for i := 0; i < len(list); i++ {
		for j := i; j > 0 && list[j-1].dish.Complexity > list[j].dish.Complexity; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	return list
}

func sortDishPriority(list []KitchenDish) []KitchenDish {
	for i := 0; i < len(list); i++ {
		for j := i; j > 0 && list[j-1].priority > list[j].priority; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	return list
}

func convertDishes(dishes []int, menu []Dish, _parent *KitchenOrder) []KitchenDish {
	kDishes := []KitchenDish{}
	for _, dID := range dishes {
		kDishes = append(kDishes, KitchenDish{parent: _parent, dish: menu[dID], priority: _parent.priority, progress: 0})
	}
	return kDishes
}

func getOptimalAparatus(aparatusList []*CookingAparatus, aType string) *CookingAparatus {
	id := 0
	minLoad := DISHBUFFER + 1
	for i := 0; i < len(aparatusList); i++ {
		if aparatusList[i].ApType == aType && aparatusList[i].ViewHoldSize() < minLoad {
			id = i
		}
	}
	return aparatusList[id]
}
