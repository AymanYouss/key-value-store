// package main

// import (
// 	"os"
// )

// func main() {

// 	db := NewInMem()

// 	repl := &Repl{
// 		db:  db,
// 		in:  os.Stdin,
// 		out: os.Stdout,
// 	}
// 	repl.Start()

// }

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// )

// func handleGet(w http.ResponseWriter, r *http.Request) {
// 	key := r.URL.Query().Get("key")
// 	if key == "" {
// 		http.Error(w, "Missing key parameter", http.StatusBadRequest)
// 		return
// 	}
// 	fmt.Println("value is", string(key))
// 	value, err := db.Get([]byte(key))
// 	fmt.Println("value is", string(value))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.Write([]byte(value))
// }

// func handleSet(w http.ResponseWriter, r *http.Request) {
// 	var data map[string]string
// 	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	key, ok := data["key"]
// 	if !ok {
// 		http.Error(w, "Missing key in JSON", http.StatusBadRequest)
// 		return
// 	}

// 	value, ok := data["value"]
// 	if !ok {
// 		http.Error(w, "Missing value in JSON", http.StatusBadRequest)
// 		return
// 	}

// 	db.Set([]byte(key), []byte(value))
// 	w.WriteHeader(http.StatusNoContent)
// }

// func handleDelete(w http.ResponseWriter, r *http.Request) {
// 	key := r.URL.Query().Get("key")
// 	if key == "" {
// 		http.Error(w, "Missing key parameter", http.StatusBadRequest)
// 		return
// 	}

// 	value, err := db.Del(key)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.Write([]byte(value))
// }

// var db *memDB

// func main() {
// 	db = NewInMem()

// 	http.HandleFunc("/get", handleGet)
// 	http.HandleFunc("/set", handleSet)
// 	http.HandleFunc("/del", handleDelete)

//		// Start HTTP server on port 8080
//		if err := http.ListenAndServe(":8080", nil); err != nil {
//			fmt.Printf("Error starting HTTP server: %s\n", err)
//		}
//	}
package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var tpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Key-Value Store</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }

        h1 {
            color: #333;
        }

        form {
            margin-bottom: 20px;
        }

        label {
            display: block;
            margin-bottom: 5px;
        }

        input {
            padding: 8px;
            margin-bottom: 10px;
        }

        button {
            padding: 8px 16px;
            background-color: #4CAF50;
            color: white;
            border: none;
            cursor: pointer;
        }

        button:hover {
            background-color: #45a049;
        }

        p {
            color: #333;
        }
    </style>
</head>
<body>
    <h1>Key-Value Store</h1>
    <form method="post" action="/set">
        <label for="key">Key:</label>
        <input type="text" name="key" required>
        <label for="value">Value:</label>
        <input type="text" name="value" required>
        <button type="submit">Set</button>
    </form>

    <form method="get" action="/get">
        <label for="key">Key:</label>
        <input type="text" name="key" required>
        <button type="submit">Get</button>
    </form>

    <form method="post" action="/del">
        <label for="key">Key:</label>
        <input type="text" name="key" required>
        <button type="submit">Delete</button>
    </form>

    <div>
        {{if .Result}}
            <p>Result: {{.Result}}</p>
        {{end}}
    </div>
</body>
</html>

`))

type PageVariables struct {
	Result string
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")

	db.Set([]byte(key), []byte(value))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")

	res, err := db.Get([]byte(key))
	result := string(res)
	if err != nil {
		result = "Key not found"
	}

	renderTemplate(w, "index", PageVariables{Result: result})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")

	result, err := db.Del(key)
	if err != nil {
		result = "Key not found"
	}

	renderTemplate(w, "index", PageVariables{Result: result})
}

func renderTemplate(w http.ResponseWriter, tmpl string, data PageVariables) {
	err := tpl.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var db *memDB

func main() {
	db = NewInMem()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index", PageVariables{})
	})
	r.HandleFunc("/set", handleSet).Methods("POST")
	r.HandleFunc("/get", handleGet).Methods("GET")
	r.HandleFunc("/del", handleDelete).Methods("POST")

	http.Handle("/", r)

	// Start HTTP server on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting HTTP server: %s\n", err)
	}
}
