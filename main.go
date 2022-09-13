package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"restaurant/kitchen/resources"
	"time"

	"github.com/gorilla/mux"
)

var runSpeed = time.Second / 10
var orderCounter = make(chan KitchenOrder, 40)
var orderCounterFinished = make(chan Response, 40)

type KitchenOrder struct {
	orderBody         Order
	receivedTime      time.Time
	fullyPreparedTime time.Time
	cookingTime       time.Duration
	// responseWriter    http.ResponseWriter
}

type Order struct {
	OrderID    int   `json:"order_id"`
	TableID    int   `json:"table_id"`
	WaiterID   int   `json:"waiter_id"`
	Items      []int `json:"items"`
	Priority   int   `json:"priority"`
	MaxWait    int   `json:"max_wait"`
	PickUpTime int   `json:"pick_up_time"`
}

type Response struct {
	OrderID        int   `json:"order_id"`
	TableID        int   `json:"table_id"`
	WaiterID       int   `json:"waiter_id"`
	Items          []int `json:"items"`
	Priority       int   `json:"priority"`
	MaxWait        int   `json:"max_wait"`
	PickUpTime     int   `json:"pick_up_time"`
	CookingTime    int   `json:"cooking_time"`
	CookingDetails []struct {
		Cook_ID int
		Food_ID int
	} `json:"cooking_details"`
}

func workingCook(parameters resources.Cook) {
	newOrder := <-orderCounter
	var cookingDetails []struct {
		Cook_ID int
		Food_ID int
	}

	for _, dishNr := range newOrder.orderBody.Items {
		dish := resources.DishMenu[dishNr-1]
		time.Sleep(time.Duration(dish.PreparationTime) * runSpeed)
		receipt := struct {
			Cook_ID int
			Food_ID int
		}{parameters.Id, dish.Id}
		cookingDetails = append(cookingDetails, receipt)
	}

	_finishedTime := time.Now()
	_cookingTime := _finishedTime.Sub(newOrder.receivedTime)

	responseBody := Response{
		OrderID:        newOrder.orderBody.OrderID,
		TableID:        newOrder.orderBody.TableID,
		WaiterID:       newOrder.orderBody.WaiterID,
		Items:          newOrder.orderBody.Items,
		Priority:       newOrder.orderBody.Priority,
		MaxWait:        newOrder.orderBody.MaxWait,
		PickUpTime:     newOrder.orderBody.PickUpTime,
		CookingTime:    int(_cookingTime/runSpeed),
		CookingDetails: cookingDetails,
	}

	fmt.Fprint(os.Stdout, parameters.Name, " finished cooking ", len(newOrder.orderBody.Items), " dishes, it took ", int64(_cookingTime/runSpeed), " seconds and now they want out of the basement \n")
	orderCounterFinished <- responseBody
}

func initializeCooks(cooks []resources.Cook) {
	for _, cook := range cooks {
		go workingCook(cook)
	}
}

func receiveRequest(w http.ResponseWriter, r *http.Request) {
	_receivedTime := time.Now()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Invalid request")
	}

	var parsedRequest Order
	json.Unmarshal(reqBody, &parsedRequest)

	newOrder := &KitchenOrder{orderBody: parsedRequest, receivedTime: _receivedTime}
	orderCounter <- *newOrder
	finishedOrder := <-orderCounterFinished

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(finishedOrder)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HomePageAccessed")
}

func main() {
	initializeCooks(resources.Staff)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/orders", receiveRequest).Methods("POST")

	log.Fatal(http.ListenAndServe(":8086", router))
}
