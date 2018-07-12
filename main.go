package main

import (
	"encoding/json"
	"fmt"
	m "go-gorillamux-gorm-pg/models"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB
var e error

func main() {
	db, e = gorm.Open("postgres", "user=postgres password=pratama dbname=postgres sslmode=disable")
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println("Connection Established")
	}
	defer db.Close()
	db.SingularTable(true)
	db.AutoMigrate(&m.Customer{}, &m.Contact{})
	db.Model(&m.Contact{}).AddForeignKey("cust_id", "customer(customer_id)", "CASCADE", "CASCADE")
	db.Model(&m.Customer{}).AddIndex("index_customer_id_name", "customer_id", "customer_name")

	router := mux.NewRouter()
	router.HandleFunc("/clear", clearCache).Methods("GET")
	router.HandleFunc("/{customers:customers\\/?}", getCustomers).Methods("GET")
	router.HandleFunc("/customers/{id}", getCustomerById).Methods("GET")
	router.HandleFunc("/customers/{name}/list", getCustomersByName).Methods("GET")
	router.HandleFunc("/customers", insertCustomer).Methods("POST")
	router.HandleFunc("/customers/{id}", updateCustomer).Methods("PUT")
	router.HandleFunc("/customers/{id}", deleteCustomer).Methods("DELETE")
	headersOk := handlers.AllowedHeaders([]string{"Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	http.ListenAndServe(":8070", handlers.CORS(originsOk, headersOk, methodsOk)(router))
}

// Clear cache
func clearCache(w http.ResponseWriter, r *http.Request) {
	key := "customer list"
	e := `"` + key + `"`
	w.Header().Set("Etag", e)
	//	w.Header().Set("Etag", "customer list")
	w.Header().Set("Cache-Control", "max-age=0, private, no-store, no-cache, must-revalidate")
	//	w.Header().Del("Etag")
	//	w.Header().Set("Refresh", "url=http://localhost:8070/customers")
	//	w.Header().Del("Cache-Control")
}

// Get customers
func getCustomers(w http.ResponseWriter, r *http.Request) {
	var customers []m.Customer
	if e := db.Preload("Contacts").Find(&customers).Error; e != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "06")
		w.Header().Set("Response-Desc", "Data Not Found")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"data not found"}`))
	} else {
		key := "customer list"
		e := `"` + key + `"`
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "00")
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Etag", e)
		w.Header().Set("Response-Desc", "Success")

		// Set Caching
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, e) {
				//				w.WriteHeader(http.StatusNotModified)
				w.WriteHeader(304)
				return
			} else {
				w.WriteHeader(200)
			}
		}
		json.NewEncoder(w).Encode(customers)
	}
}

// Get customers by name
func getCustomersByName(w http.ResponseWriter, r *http.Request) {
	var customers []m.Customer
	param := mux.Vars(r)
	if e := db.Where("customer_name = ?", param["name"]).Preload("Contacts").Find(&customers).Error; e != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "06")
		w.Header().Set("Response-Desc", "Data Not Found")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"data not found"}`))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "00")
		w.Header().Set("Response-Desc", "Success")
		json.NewEncoder(w).Encode(&customers)
	}
}

// Get customer by id
func getCustomerById(w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	param := mux.Vars(r)
	if e := db.Where("customer_id = ?", param["id"]).Preload("Contacts").First(&customer).Error; e != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "06")
		w.Header().Set("Response-Desc", "Data Not Found")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"data not found"}`))
	} else {
		key := "customer list"
		e := `"` + key + `"`
		w.Header().Set("Etag", e)
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "00")
		w.Header().Set("Response-Desc", "Success")

		//		// Set Caching
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, e) {
				//				w.WriteHeader(http.StatusNotModified)
				w.WriteHeader(304)
				return
			} else {
				w.WriteHeader(200)
			}
		}

		json.NewEncoder(w).Encode(&customer)
	}
}

// Insert cusotmer
func insertCustomer(w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	var _ = json.NewDecoder(r.Body).Decode(&customer)
	db.Create(&customer)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Response-Code", "00")
	w.Header().Set("Response-Desc", "Success")
	json.NewEncoder(w).Encode(&customer)
}

// Update customer
func updateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	param := mux.Vars(r)
	if e := db.Where("customer_id = ?", param["id"]).Preload("Contacts").First(&customer).Error; e != nil {
		w.Header().Set("Content-Type", "application-json")
		w.Header().Set("Response-Code", "06")
		w.Header().Set("Response-Desc", "Data Not Found")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"data not found"}`))
	} else {
		_ = json.NewDecoder(r.Body).Decode(&customer)
		db.Save(&customer)
		w.Header().Set("Content-Type", "application-json")
		w.Header().Set("Response-Code", "00")
		w.Header().Set("Response-Desc", "Success")
		json.NewEncoder(w).Encode(&customer)
	}
}

// Delete customer
func deleteCustomer(w http.ResponseWriter, r *http.Request) {
	var customer m.Customer
	param := mux.Vars(r)
	if e := db.Where("customer_id = ?", param["id"]).Preload("Contacts").First(&customer).Error; e != nil {
		w.Header().Set("Content-Type", "application-json")
		w.Header().Set("Response-Code", "06")
		w.Header().Set("Response-Desc", "Data Not Found")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"data not found"}`))
	} else {
		db.Where("customer_id=?", param["id"]).Preload("Contacts").Delete(&customer)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Response-Code", "00")
		w.Header().Set("Response-Desc", "Success")
	}
}
