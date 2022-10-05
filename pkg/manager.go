package pkg

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	DISHBUFFER = 100
)

type Manager struct {
	dinningHallUrl string
	runspeed       time.Duration

	orderChannel   chan Order
	finishKDish    chan KitchenDish
	contactChannel chan *Cook

	pendingDish chan struct{}

	kMutex sync.Mutex

	menu        []Dish
	priorityLog []int

	dishList []KitchenDish
}

func NewManager(dishmenu []Dish, _orderChannel chan Order, url string, _runspeed time.Duration) *Manager {
	m := &Manager{
		menu:           dishmenu,
		dinningHallUrl: url,
		runspeed:       _runspeed,
		orderChannel:   _orderChannel,
		finishKDish:    make(chan KitchenDish, DISHBUFFER),
		contactChannel: make(chan *Cook, DISHBUFFER),
		pendingDish:    make(chan struct{}, DISHBUFFER),
	}
	return m
}

func (m *Manager) ViewContactChannel() chan *Cook {
	return m.contactChannel
}

func (m *Manager) ViewFinishedDishChannel() chan KitchenDish {
	return m.finishKDish
}

func (m *Manager) Start() {
	go m.getOrders()
	go m.sendDishCook()
	go m.receiveFinishedDishes()
}

func (m *Manager) getOrders() {
	for {
		order := <-m.orderChannel
		newKOrder := newKitchenOrder(order)
		newKOrder.priority = order.Priority
		newDishes := convertDishes(order.Items, m.menu, &newKOrder)
		newDishes = sortDishComplexity(newDishes)

		m.kMutex.Lock()
		m.dishList = append(m.dishList, newDishes...)
		m.sortDishes()
		m.kMutex.Unlock()

		for i := 0; i < len(newDishes); i++ {
			m.pendingDish <- struct{}{}
		}
	}
}

func (m *Manager) sendDishCook() {
	for {
		satisfied := false
		<-m.pendingDish
		cook := <-m.contactChannel
		m.kMutex.Lock()
		for i := 0; i < len(m.dishList); i++ {
			if m.dishList[i].dish.Complexity <= cook.Rank {
				cook.inputDishChannel <- m.dishList[i]
				m.dishList = append(m.dishList[:i], m.dishList[i+1:]...)
				satisfied = true
				break
			}
		}
		m.kMutex.Unlock()

		if !satisfied {
			m.pendingDish <- struct{}{}
			go func(){
				time.Sleep(m.runspeed)
				m.contactChannel <- cook
			}()
		}
	}
}

func (m *Manager) receiveFinishedDishes() {
	for {
		kdish := <-m.finishKDish
		kdish.parent.CookingDetails = append(kdish.parent.CookingDetails, struct {
			Cook_ID int
			Food_ID int
		}{kdish.cookID, kdish.dish.Id})
		kdish.parent.finDishes++
		if kdish.parent.finDishes == len(kdish.parent.OrderBody.Items) {
			kdish.parent.CookingTime = int(time.Now().Sub(kdish.parent.ReceivedTime) / m.runspeed)
			log.Println("Order", kdish.parent.OrderBody.OrderID, "fully finished.", kdish.parent.OrderBody.Items)
			m.sendResponseDinningHall(compileResponse(*kdish.parent))
		}
	}
}

func (m *Manager) sendResponseDinningHall(_response OrderResponse) {
	log.Println("Order sent back:", _response.OrderID)

	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(_response)

	req, _ := http.NewRequest("POST", m.dinningHallUrl, payloadBuffer)
	client := &http.Client{}
	client.Do(req)
}

func (m *Manager) sortDishes() {
	newlist := []KitchenDish{}
	for i := 1; i <= 5; i++ {
		for _, d := range m.dishList {
			if d.parent.OrderBody.Priority == i {
				newlist = append(newlist, d)
			}
		}
	}
	m.dishList = newlist
}

func (m *Manager) addOrder(parsedOrder Order) {
	log.Println("Order", parsedOrder.OrderID, "received")
	m.orderChannel <- parsedOrder
}
