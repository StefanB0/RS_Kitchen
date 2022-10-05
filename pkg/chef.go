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
}



func (c *Cook) Start(_runspeed time.Duration, _finishdishChannel chan KitchenDish, _managerContact chan *Cook) {
	c.runspeed = _runspeed
	c.finishDishChannel = _finishdishChannel
	c.managerContact = _managerContact

	c.inputDishChannel = make(chan KitchenDish, c.Proficiency)
	c.dCounter = make(chan struct{}, c.Proficiency)

	log.Println(c.Name, "started working!")
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
	return newDish
}

func (c *Cook) cookDish(kdish KitchenDish) {
	time.Sleep(c.runspeed * time.Duration(kdish.dish.PreparationTime))
	c.ReturnDish(kdish)

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

