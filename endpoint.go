package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "nandor"
	password = "password"
	dbname   = "nandor"
)

type Spot struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Website     sql.NullString  `json:"website"`
	Coordinates string  `json:"coordinates"`
	Description sql.NullString  `json:"description"`
	Rating      float64 `json:"rating"`
	Distance    float64 `json:"distance"`
}

type Response struct {
	Spots []Spot `json:"spots"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/spots", getSpots).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getSpots(w http.ResponseWriter, r *http.Request) {
	latitudeStr := r.URL.Query().Get("latitude")
	longitudeStr := r.URL.Query().Get("longitude")
	radiusStr := r.URL.Query().Get("radius")
	shape := r.URL.Query().Get("type")

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude parameter", http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude parameter", http.StatusBadRequest)
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		http.Error(w, "Invalid radius parameter", http.StatusBadRequest)
		return
	}

	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	spots, err := findSpotsInArea(db, latitude, longitude, radius, shape)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving spots", http.StatusInternalServerError)
		return
	}

	response := Response{
		Spots: spots,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func findSpotsInArea(db *sql.DB, latitude, longitude, radius float64, shape string) ([]Spot, error) {
	var query string
	if shape == "circle" {
		query = `
			SELECT id, name, website, coordinates, description, rating,
			ST_Distance(ST_MakePoint($1, $2)::geography, coordinates::geography) AS distance
			FROM "MY_TABLE"
			WHERE ST_DWithin(ST_MakePoint($1, $2)::geography, coordinates::geography, $3)
			ORDER BY distance, rating DESC
		`
	} else if shape == "square" {
		query = `
			SELECT id, name, website, coordinates, description, rating
			FROM "MY_TABLE"
			WHERE ST_DWithin(ST_MakePoint($1, $2)::geography, coordinates::geography, $3)
			ORDER BY rating DESC
		`
		rows, err := db.Query(query, longitude, latitude, radius)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		spots := []Spot{}
		for rows.Next() {
			var spot Spot
			err := rows.Scan(&spot.ID, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating)
			if err != nil {
				return nil, err
			}
			spots = append(spots, spot)
		}

		return spots, nil
	} else {
		return nil, fmt.Errorf("Invalid shape parameter")
	}

	rows, err := db.Query(query, longitude, latitude, radius)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	spots := []Spot{}
	var prevDistance float64
	for rows.Next() {
		var spot Spot
		err := rows.Scan(&spot.ID, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating, &spot.Distance)
		if err != nil {
			return nil, err
		}

		if prevDistance < 50 && spot.Distance < 50 {
			if spot.Rating > spots[len(spots)-1].Rating {
				spots[len(spots)-1], spot = spot, spots[len(spots)-1]
			}
		}

		spots = append(spots, spot)
		prevDistance = spot.Distance
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return spots, nil
}
