package pkg

import "time"

type CookingAparatus struct {
	name     string
	waitList chan KitchenDish
	runspeed time.Duration
}

func NewCookingAparatus(_name string, _runspeed time.Duration) *CookingAparatus {
	return &CookingAparatus{
		name: _name, 
		waitList: make(chan KitchenDish, DISHBUFFER),
		runspeed: _runspeed,
	}
}

func (ca *CookingAparatus) Work() {
	for {
		kdish := <-ca.waitList
		ca.cookDish(kdish)
	}
}

func (ca *CookingAparatus) ViewName() string {
	return ca.name
}

func (ca *CookingAparatus) ViewChannel() chan KitchenDish{
	return ca.waitList
}

func (ca *CookingAparatus) PutOrder(kDish KitchenDish) {
	ca.waitList <- kDish
}

func (ca *CookingAparatus) cookDish(kDish KitchenDish) {
	time.Sleep(ca.runspeed * time.Duration(kDish.dish.PreparationTime))
	kDish.cook.ReturnDish(kDish)
}


// TODO cooking aparatus basic
// TODO cooking aparatus context switch with on-hold list