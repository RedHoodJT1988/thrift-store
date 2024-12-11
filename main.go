package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"html/template"
	"os"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"github.com/joho/godotenv"
)

var db *gorm.DB

type Product struct {
	gorm.Model
	Name string `json:"name"`
	Description string `json:"description"`
	Price float64 `json:"price"`
	Quantity int `json:"quantity"`
}

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func initDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
        os.Getenv("PG_HOST"),
        os.Getenv("PG_USER"),
        os.Getenv("PG_PASSWORD"),
        os.Getenv("PG_DATABASE"),
        os.Getenv("PG_PORT"),
    )
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}

func getAdminPage(w http.ResponseWriter, r *http.Request) {
	var products []Product
	db.Find(&products)
	renderTemplate(w, "admin.html", products)
}

func getNewProductForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new-product.html", nil)
}

func getEditProductForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var p Product
	db.First(&p, id)
	renderTemplate(w, "edit-product.html", p)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Create(&p)

	w.WriteHeader(http.StatusCreated)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Model(&Product{}).Where("id = ?", id).Updates(p)

	w.WriteHeader(http.StatusNoContent)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db.Delete(&Product{}, id)

	w.WriteHeader(http.StatusNoContent)
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t, err := template.ParseFiles(filepath.Join("templates", name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	initDB()

	db.AutoMigrate(&Product{})

	r := mux.NewRouter()
	r.HandleFunc("/admin", getAdminPage).Methods("GET")
	r.HandleFunc("/admin/new-product", getNewProductForm).Methods("GET")
	r.HandleFunc("/admin/products/{id}/edit", getEditProductForm).Methods("GET")
	r.HandleFunc("/admin/products/{id}", updateProduct).Methods("PUT")
	r.HandleFunc("/admin/products/{id}", deleteProduct).Methods("DELETE")
	r.HandleFunc("/admin/products", createProduct).Methods("POST")

	fmt.Println("Server started on prot 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}