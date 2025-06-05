package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func getProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, price, stock FROM products")
		if err != nil {
			http.Error(w, "Error al cosultar los productos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		products := make([]Product, 0)

		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
				http.Error(w, "Error al escanear producto", http.StatusInternalServerError)
				return
			}
			products = append(products, p)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}
}

func createProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Product

		// Leer el cuerpo de la solicitud (JSON)
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}

		// Insertar en la base de datos
		query := "INSERT INTO products (name, price, stock) VALUES (?,?,?)"
		result, err := db.Exec(query, p.Name, p.Price, p.Stock)
		if err != nil {
			http.Error(w, "Error al insertar el producto", http.StatusInternalServerError)
			return
		}

		// Obterner el ID generado
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Error al obtener ID", http.StatusInternalServerError)
			return
		}

		p.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	}
}

func getProductByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/products/"):] // Extrae el ID de la URL

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID Invalido", http.StatusBadRequest)
			return
		}
		query := "SELECT id, name, price, stock FROM products WHERE id = ?"
		row := db.QueryRow(query, id)

		var p Product

		err = row.Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
		if err == sql.ErrNoRows {
			http.Error(w, "Producto no encontrado", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Error al obtener el producto", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}
}

func updateProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Extraer el ID desde la URL
		prefix := "/products/update/"
		idStr := r.URL.Path[len(prefix):]

		// Eliminar barra final si existe
		if len(idStr) > 0 && idStr[len(idStr)-1] == '/' {
			idStr = idStr[:len(idStr)-1]
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID invalido", http.StatusBadRequest)
			return
		}

		var p Product
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}

		query := "UPDATE products SET name = ?, price = ?, stock = ? WHERE ID = ?"
		_, err = db.Exec(query, p.Name, p.Price, p.Stock, id)
		if err != nil {
			http.Error(w, "Error al actualizar el producto", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Producto actualizado correctamente"))
	}
}

func deleteProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		prefix := "/products/delete/"
		idStr := r.URL.Path[len(prefix):]
		fmt.Println("ID recibido:", idStr)

		if len(idStr) > 0 && idStr[len(idStr)-1] == '/' {
			idStr = idStr[:len(idStr)-1]
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		query := "DELETE FROM products WHERE id = ?"
		result, err := db.Exec(query, id)
		if err != nil {
			http.Error(w, "Error al eliminar el producto", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Producto no encontrado", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Producto eliminado correctamente"))
	}
}

func main() {
	fmt.Println("Microservicio de productos iniciado")

	// Conectar a la base de datos
	db, err := sql.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Probar la conexión
	err = db.Ping()
	if err != nil {
		log.Fatal("No se pudo conectar a la base de datos:", err)
	}
	fmt.Println("Conectado a la base de datos")

	http.HandleFunc("/products", getProducts(db))
	http.HandleFunc("/products/create", createProduct(db))
	http.HandleFunc("/products/", getProductByID(db))
	http.HandleFunc("/products/update/", updateProduct(db)) // Este también matchea con /update
	http.HandleFunc("/products/delete/", deleteProduct(db))
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Levantar el servidor en el puerto 8080
	http.ListenAndServe(":8081", nil)

}
