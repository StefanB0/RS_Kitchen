package main

import (
	"log"
	"restaurant/kitchen/pkg"
	"time"
)

const (
	dinningHallUrl = "http://hall:8882/distribution"
	LISTENPORT     = ":8881"
	DISHBUFFER     = 100
	COOK_NR        = 4
	TABLE_NR       = 10
	runSpeed       = time.Millisecond * 20
	fractions      = 10
)

var (
	manager      *pkg.Manager
	AparatusList = []*pkg.CookingAparatus{pkg.NewCookingAparatus("oven", runSpeed), pkg.NewCookingAparatus("oven", runSpeed), pkg.NewCookingAparatus("stove", runSpeed)}

	orderChannel   = make(chan pkg.Order, TABLE_NR)
	finishDish     = make(chan pkg.KitchenDish, DISHBUFFER)
	contactChannel = make(chan *pkg.Cook, DISHBUFFER)

	dishMenu []pkg.Dish
	staff    []pkg.Cook
)

func initializeAparatus() {
	for i := 0; i < len(AparatusList); i++ {
		go AparatusList[i].Work()
	}
}

func initializeCooks(cooks []pkg.Cook) {
	for i := 0; i < COOK_NR; i++ {
		cooks[i].Start(runSpeed, manager.ViewFinishedDishChannel(), manager.ViewContactChannel(), AparatusList)
	}
}

func main() {
	log.Println("Kitchen take off!")
	dishMenu = pkg.ReadMenu("pkg/menu.json")
	staff = pkg.ReadCooks("pkg/staff.json")

	manager = pkg.NewManager(dishMenu, orderChannel, dinningHallUrl, runSpeed)
	manager.Start()
	initializeAparatus()
	initializeCooks(staff)

	pkg.StartServer(manager, LISTENPORT)
}
