package pkg

import (
	"sync"
	"time"
)

type Cook struct {
	Id          int
	Rank        int
	Proficiency int
	Name        string
	Catchphrase string
	fCounter    int
}

type PendingList struct {
	DishList []KitchenDish
	sync.Mutex
}

func (c *Cook) work(pendingL *PendingList, finishedCh chan KitchenDish, runSpeed time.Duration) {
	for {
		if c.fCounter < c.Proficiency {
			c.fCounter++
			kdish := c.ChooseFood(pendingL)
			if kdish.index != 0 {
				go c.CookFood(kdish, runSpeed, finishedCh)
			} else {
				c.fCounter--
			}
		}
	}
}

func (c *Cook) ChooseFood(pendingL *PendingList) KitchenDish {
	pendingL.Lock()
	defer pendingL.Unlock()

	for i := range pendingL.DishList {
		if c.CheckComplexity(pendingL.DishList[i]) {
			pendingL.DishList = removeDish(pendingL.DishList, i)
			return pendingL.DishList[i]
		}
	}
	return KitchenDish{}
}

func (c *Cook) CheckComplexity(d KitchenDish) bool {
	switch {
	case d.Complexity > c.Rank:
		return false
	case d.Complexity < c.Rank-1:
		return false
	}

	return true
}

func (c *Cook) CookFood(kdish KitchenDish, runSpeed time.Duration, finishedCh chan KitchenDish) {
	time.Sleep(runSpeed * time.Duration(kdish.PreparationTime))
	c.fCounter--
	c.ReturnDish(kdish, finishedCh)
}

func (c *Cook) ReturnDish(kdish KitchenDish, finishedCh chan KitchenDish) {
	finishedCh <- kdish
}
