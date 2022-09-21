package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"restaurant/kitchen/pkg"

	"github.com/gorilla/mux"
)

const (
	runSpeed       = time.Millisecond
	dinningHallUrl = "http://hall:8882/distribution"
	LISTENPORT     = ":8881"
)

var (
	orderCounter         = make(chan pkg.KitchenOrder, 40)
	orderCounterFinished = make(chan pkg.OrderResponse, 40)

	chefLog          *log.Logger
	eventLog         *log.Logger
	communicationLog *log.Logger
	errorLog         *log.Logger
)

type MyServer struct {
	http.Server
	shutdownReq chan bool
	reqCount    uint32
}

func workingCook(parameters pkg.Cook) {
	for {
		newOrder := <-orderCounter
		var cookingDetails []struct {
			Cook_ID int
			Food_ID int
		}

		for _, dishNr := range newOrder.OrderBody.Items {
			dish := pkg.DishMenu[dishNr-1]
			time.Sleep(time.Duration(dish.PreparationTime) * runSpeed)
			receipt := struct {
				Cook_ID int
				Food_ID int
			}{parameters.Id, dish.Id}
			cookingDetails = append(cookingDetails, receipt)
		}

		_finishedTime := time.Now()
		_cookingTime := _finishedTime.Sub(newOrder.ReceivedTime)

		responseBody := pkg.OrderResponse{
			OrderID:        newOrder.OrderBody.OrderID,
			TableID:        newOrder.OrderBody.TableID,
			WaiterID:       newOrder.OrderBody.WaiterID,
			Items:          newOrder.OrderBody.Items,
			Priority:       newOrder.OrderBody.Priority,
			MaxWait:        newOrder.OrderBody.MaxWait,
			PickUpTime:     newOrder.OrderBody.PickUpTime,
			CookingTime:    int(_cookingTime / runSpeed),
			CookingDetails: cookingDetails,
		}

		chefLog.Println(parameters.Name, "Finished order:", newOrder.OrderBody.OrderID, ":", newOrder.OrderBody.Items, "Time:", _cookingTime)
		orderCounterFinished <- responseBody
	}
}

func initializeCooks(cooks []pkg.Cook) {
	for _, cook := range cooks {
		log.Println(cook.Name, "started working!")
		go workingCook(cook)
	}
}

func receiveRequest(w http.ResponseWriter, r *http.Request) {
	_receivedTime := time.Now()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Invalid request")
	}

	var parsedRequest pkg.Order
	json.Unmarshal(reqBody, &parsedRequest)

	newOrder := &pkg.KitchenOrder{OrderBody: parsedRequest, ReceivedTime: _receivedTime}
	communicationLog.Println("Order received:", newOrder.OrderBody.OrderID)

	orderCounter <- *newOrder
	finishedOrder := <-orderCounterFinished

	sendOrderDinningHall(finishedOrder)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(finishedOrder)
}

func sendOrderDinningHall(_response pkg.OrderResponse) {
	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(_response)

	req, _ := http.NewRequest("POST", dinningHallUrl, payloadBuffer)
	client := &http.Client{}
	client.Do(req)
  communicationLog.Println("Order send:", _response.OrderID)
}

func initLogs() {
	// chefFile, err1 := os.OpenFile("logs/chef_logs.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	// eventFile, err2 := os.OpenFile("logs/event_logs.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	// communicationFile, err3 := os.OpenFile("logs/communication_logs.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	// errorFile, err4 := os.OpenFile("logs/error_logs.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)

	// if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
	// 	log.Fatal(err1, err2, err3, err4)
	// }

	chefLog = log.New(log.Writer(), "Chef: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	eventLog = log.New(log.Writer(), "Event: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	communicationLog = log.New(log.Writer(), "Communication: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	errorLog = log.New(log.Writer(), "Error: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func NewServer() *MyServer {
	//create server
	s := &MyServer{
		Server: http.Server{
			Addr:         LISTENPORT,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		shutdownReq: make(chan bool),
	}

	router := mux.NewRouter()

	//register handlers
	router.HandleFunc("/shutdown", s.ShutdownHandler)
	router.HandleFunc("/order", receiveRequest).Methods("POST")

	//set http server handler
	s.Handler = router

	return s
}

func (s *MyServer) WaitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
	case sig := <-s.shutdownReq:
		log.Printf("Shutdown request (/shutdown %v)", sig)
	}

	log.Printf("Stoping http server ...")

	//Create shutdown context with 10 second timeout
  ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v", err)
	}
}

func (s *MyServer) ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shutdown server"))

	//Do nothing if shutdown request already issued
	//if s.reqCount == 0 then set to 1, return true otherwise false
	if !atomic.CompareAndSwapUint32(&s.reqCount, 0, 1) {
		log.Printf("Shutdown through API call in progress...")
		return
	}

	go func() {
		s.shutdownReq <- true
	}()
}

func startServer() {
	server := NewServer()

	done := make(chan bool)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Listen and serve: %v", err)
		}
		done <- true
	}()

	//wait shutdown
	server.WaitShutdown()

	<-done
	log.Printf("DONE!")
}

func main() {
	initLogs()
	initializeCooks(pkg.Staff)

	startServer()
}
