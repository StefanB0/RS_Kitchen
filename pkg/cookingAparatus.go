package pkg

import (
	"sync"
	"time"
)

type CookingAparatus struct {
	ApType           string
	hold             []KitchenDish
	holdSize         int
	MaxLocalPriority int
	newDishChan      chan KitchenDish
	runspeed         time.Duration

	semaphore chan struct{}
	mu        sync.Mutex
}

func NewCookingAparatus(_type string, _runspeed time.Duration) *CookingAparatus {
	return &CookingAparatus{
		ApType:      _type,
		holdSize:    0,
		hold:        make([]KitchenDish, DISHBUFFER),
		newDishChan: make(chan KitchenDish, DISHBUFFER),
		semaphore:   make(chan struct{}, DISHBUFFER),
		runspeed:    _runspeed,
	}
}

func (ca *CookingAparatus) Work() {
	go ca.addHold()
	for {
		<-ca.semaphore
		ca.mu.Lock()
		dish := ca.hold[ca.holdSize-1]
		ca.hold = ca.hold[:ca.holdSize-1]
		ca.holdSize = len(ca.hold)
		ca.mu.Unlock()
		ca.cookDish(dish)
	}
}

func (ca *CookingAparatus) addHold() {
	for {
		dish := <-ca.newDishChan
		ca.mu.Lock()
		ca.hold = append(ca.hold, dish)
		ca.hold = sortDishPriority(ca.hold)
		ca.holdSize = len(ca.hold)
		ca.MaxLocalPriority = ca.calcLocalPriority()
		ca.semaphore <- struct{}{}
		ca.mu.Unlock()
	}
}

func (ca *CookingAparatus) addToHold(kDish KitchenDish) {
	ca.newDishChan <- kDish

}

func (ca *CookingAparatus) calcLocalPriority() int {
	min := 5
	for i := 0; i < ca.holdSize; i++ {
		if ca.hold[i].priority <= min {
			min = ca.hold[i].priority
		}
	}
	return min
}

func (ca *CookingAparatus) ViewChannel() chan KitchenDish {
	return ca.newDishChan
}

func (ca *CookingAparatus) ViewHoldSize() int {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.holdSize
}

func (ca *CookingAparatus) cookDish(kDish KitchenDish) {
	time.Sleep(ca.runspeed)
	kDish.progress++
	if kDish.progress >= kDish.dish.PreparationTime {
		kDish.cook.ReturnDish(kDish)
	} else {
		ca.addToHold(kDish)
	}
}
