package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "sync"
)

type Item struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

var (
    items   = make(map[int]Item)
    nextID  = 1
    itemsMu sync.Mutex
)

func main() {
    http.HandleFunc("/items", itemsHandler)
    http.HandleFunc("/items/", itemHandler)
    log.Println("Server is running on port 9000")
    log.Fatal(http.ListenAndServe(":9000", nil))
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        listItems(w, r)
    case http.MethodPost:
        createItem(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func itemHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.URL.Path[len("/items/"):])
    if err != nil {
        http.Error(w, "Invalid item ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
    case http.MethodGet:
        getItem(w, r, id)
    case http.MethodPut:
        updateItem(w, r, id)
    case http.MethodDelete:
        deleteItem(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func listItems(w http.ResponseWriter, r *http.Request) {
    itemsMu.Lock()
    defer itemsMu.Unlock()
    itemsList := make([]Item, 0, len(items))
    for _, item := range items {
        itemsList = append(itemsList, item)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(itemsList)
}

func createItem(w http.ResponseWriter, r *http.Request) {
    var item Item
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    itemsMu.Lock()
    item.ID = nextID
    nextID++
    items[item.ID] = item
    itemsMu.Unlock()
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(item)
}

func getItem(w http.ResponseWriter, r *http.Request, id int) {
    itemsMu.Lock()
    item, ok := items[id]
    itemsMu.Unlock()
    if !ok {
        http.Error(w, "Item not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request, id int) {
    var updatedItem Item
    if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    itemsMu.Lock()
    _, ok := items[id]
    if !ok {
        itemsMu.Unlock()
        http.Error(w, "Item not found", http.StatusNotFound)
        return
    }
    updatedItem.ID = id
    items[id] = updatedItem
    itemsMu.Unlock()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedItem)
}

func deleteItem(w http.ResponseWriter, r *http.Request, id int) {
    itemsMu.Lock()
    _, ok := items[id]
    if ok {
        delete(items, id)
    }
    itemsMu.Unlock()
    if !ok {
        http.Error(w, "Item not found", http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

