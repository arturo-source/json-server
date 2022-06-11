package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	port   = ":8090"
	dbName = "db.json"
	db     = make(map[string][]interface{})
)

type ResponseError struct {
	ErrorMessage string `json:"error_message"`
}

func NewResponseError(message string) []byte {
	dat, _ := json.Marshal(ResponseError{ErrorMessage: message})
	return dat
}

func ReadDB() error {
	if _, err := os.Stat(dbName); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	dat, err := os.ReadFile(dbName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dat, &db)

	return err
}

func WriteDB() error {
	dat, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = os.WriteFile(dbName, dat, 0644)

	return err
}

func GetPositionNumber(w http.ResponseWriter, params []string) int {
	if len(params) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewResponseError("Position number is missing"))
		return -1
	}

	positionNumber, err := strconv.Atoi(params[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewResponseError("Position could not be converted to int"))
		return -1
	}

	return positionNumber
}

func ReadBody(w http.ResponseWriter, r *http.Request) interface{} {
	var dat interface{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&dat)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewResponseError("Error decoding data"))
		return nil
	}

	return dat
}

func GetAll(w http.ResponseWriter, tableName string) {
	data, err := json.Marshal(db[tableName])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error getting data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetFromPosition(w http.ResponseWriter, tableName string, params []string) {
	positionNumber := GetPositionNumber(w, params)
	if positionNumber == -1 {
		return
	}

	if positionNumber >= len(db[tableName]) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(NewResponseError("Position is bigger than the number of elements in the table"))
		return
	}

	dat, err := json.Marshal(db[tableName][positionNumber])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error getting data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func PostData(w http.ResponseWriter, dat interface{}, tableName string) {
	db[tableName] = append(db[tableName], dat)

	err := WriteDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error writing data in db"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data appended"))
}

func PutData(w http.ResponseWriter, dat interface{}, tableName string, params []string) {
	positionNumber := GetPositionNumber(w, params)
	if positionNumber == -1 {
		return
	}

	if positionNumber >= len(db[tableName]) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(NewResponseError("Position is bigger than the number of elements in the table"))
		return
	}

	db[tableName][positionNumber] = dat

	err := WriteDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error writing data in db"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data modified"))
}

func DeleteTable(w http.ResponseWriter, tableName string) {
	if _, exists := db[tableName]; !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write(NewResponseError("Table does not exist"))
		return
	}

	delete(db, tableName)

	err := WriteDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error writing data in db"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Table deleted"))
}

func DeleteFromPosition(w http.ResponseWriter, r *http.Request, tableName string, params []string) {
	positionNumber := GetPositionNumber(w, params)
	if positionNumber == -1 {
		return
	}

	if positionNumber >= len(db[tableName]) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(NewResponseError("Position is bigger than the number of elements in the table"))
		return
	}

	// This is slow for large tables
	// but this program is intended to launch
	// a fast back-end server, not for a real database
	db[tableName] = append(db[tableName][:positionNumber], db[tableName][positionNumber+1:]...)

	err := WriteDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(NewResponseError("Error writing data in db"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data deleted"))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	nParams := len(params)
	tableName := params[1]

	w.Header().Set("Content-Type", "application/json")

	if tableName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewResponseError("Table name is required"))
		return
	}

	switch r.Method {
	case "GET":
		if _, exists := db[tableName]; !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write(NewResponseError("Table not found"))
			break
		}
		if nParams == 2 {
			GetAll(w, tableName)
			break
		}
		GetFromPosition(w, tableName, params)

		break
	case "POST":
		dat := ReadBody(w, r)
		if dat == nil {
			break
		}
		PostData(w, dat, tableName)
		break
	case "PUT":
		dat := ReadBody(w, r)
		if dat == nil {
			break
		}
		PutData(w, dat, tableName, params)
		break
	case "DELETE":
		if nParams == 2 {
			DeleteTable(w, tableName)
			break
		}
		DeleteFromPosition(w, r, tableName, params)
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(NewResponseError("Unknown method. Use instead GET, POST, PUT or DELETE."))
		break
	}
}

func main() {
	err := ReadDB()
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(port, http.HandlerFunc(Handler))
	if err != nil {
		panic(err)
	}
}
