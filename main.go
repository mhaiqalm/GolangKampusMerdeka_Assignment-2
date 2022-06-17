package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/mysql"
)

type Response struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

// Order represents the model for an order
type Order struct {
	OrderID      int       `json:"orderId" gorm:"primaryKey" `
	CustomerName string    `json:"customerName"`
	OrderedAt    time.Time `json:"orderedAt"`
	Items        []Item    `json:"items" gorm:"foreignkey:OrderID"`
}

// Item represents the model for an order
type Item struct {
	ItemID      int    `json:"itemId" gorm:"primaryKey" `
	ItemCode    string `json:"itemCode"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	OrderID     int    `json:"-"`
}

var db *gorm.DB
var orders []Order
var prevOrderID = 0
var err error

func dbInit() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/orders_by?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Order{}, &Item{})
	fmt.Println("Success Open DB")
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db.Preload("Items").Find(&orders)
	json.NewEncoder(w).Encode(orders)
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {

	var order Order
	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		fmt.Println("ERROR")
		log.Fatal(err)
	}

	db.Create(&order)
	json.NewEncoder(w).Encode(order)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	inputOrderID, error := strconv.Atoi(params["orderId"])
	if error != nil {
		log.Fatal(error.Error())
		return
	}
	for _, order := range orders {
		if order.OrderID == inputOrderID {
			json.NewEncoder(w).Encode(order)
			return
		}
	}
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	var updatedOrder Order
	json.NewDecoder(r.Body).Decode(&updatedOrder)
	params := mux.Vars(r)
	inputOrderID, error := strconv.Atoi(params["orderId"])
	if error != nil {
		log.Fatal(error.Error())
		return
	}
	updatedOrder.OrderID = inputOrderID
	db.Where("order_id = ?", params["orderId"]).Delete(&Item{})
	db.Save(&updatedOrder)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)

}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	db.Where("order_id = ?", params["orderId"]).Delete(&Item{})
	db.Where("order_id = ?", params["orderId"]).Delete(&Order{})

	response := Response{
		Code:   "200",
		Status: "Success",
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	dbInit()
	router := mux.NewRouter()
	// Create
	router.HandleFunc("/orders", CreateOrder).Methods("POST")
	// // Read
	router.HandleFunc("/orders/{orderId}", getOrder).Methods("GET")
	// Read-all
	router.HandleFunc("/orders", getOrders).Methods("GET")
	// Update
	router.HandleFunc("/orders/{orderId}", updateOrder).Methods("PUT")
	// Delete
	router.HandleFunc("/orders/{orderId}", deleteOrder).Methods("DELETE")
	fmt.Println("STARTING SERVER AT localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}