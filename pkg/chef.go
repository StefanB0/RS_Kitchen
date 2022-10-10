package pkg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Cook struct {
	Id          int
	Rank        int
	Proficiency int
	Name        string
	Catchphrase string

	runspeed time.Duration

	dCounter          chan struct{}
	inputDishChannel  chan KitchenDish
	finishDishChannel chan KitchenDish
	managerContact    chan *Cook
	aparatusList      []*CookingAparatus
}

func (c *Cook) Start(_runspeed time.Duration, _finishdishChannel chan KitchenDish, _managerContact chan *Cook, _aparatusList []*CookingAparatus) {
	c.runspeed = _runspeed
	c.finishDishChannel = _finishdishChannel
	c.managerContact = _managerContact
	c.aparatusList = _aparatusList

	c.inputDishChannel = make(chan KitchenDish, c.Proficiency)
	c.dCounter = make(chan struct{}, c.Proficiency)

	go c.work()
}

func (c *Cook) work() {
	for {
		if len(c.dCounter) < c.Proficiency {
			c.dCounter <- struct{}{}
			go func() {
				kDish := c.getDish()
				c.cookDish(kDish)
			}()
		}
	}
}

func (c *Cook) getDish() KitchenDish {
	cp := c
	c.managerContact <- cp

	newDish := <-c.inputDishChannel
	newDish.cook = c
	return newDish
}

func (c *Cook) cookDish(kdish KitchenDish) {
	if kdish.dish.CookingApparatus == "null" {
		for i := 0; i < kdish.dish.PreparationTime; i++ {
			time.Sleep(c.runspeed)
		}
		c.ReturnDish(kdish)
	} else {
		cookingAparatus := getOptimalAparatus(c.aparatusList, kdish.dish.CookingApparatus)
		cookingAparatus.addToHold(kdish)
	}

	<-c.dCounter
}

func (c *Cook) ReturnDish(kdish KitchenDish) {
	c.finishDishChannel <- kdish
}

func ReadCooks(path string) []Cook {

	jsonfile, err := os.Open(path)
	defer jsonfile.Close()

	if err != nil {
		log.Println(err)
	}

	bytevalue, _ := ioutil.ReadAll(jsonfile)
	newStaff := []Cook{}
	json.Unmarshal(bytevalue, &newStaff)

	return newStaff

}

