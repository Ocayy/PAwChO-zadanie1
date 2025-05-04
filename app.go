package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	author = "Jakub Nowosad"
	port   = "8080"
)

type City struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

var cities = map[string]City{
	"warszawa": {"Warszawa", 52.2297, 21.0122},
	"krakow":   {"Kraków", 50.0647, 19.9450},
	"londyn":   {"Londyn", 51.5074, -0.1278},
	"paryz":    {"Paryż", 48.8566, 2.3522},
	"berlin":   {"Berlin", 52.5200, 13.4050},
}

func main() {
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	fmt.Println("=== APLIKACJA URUCHOMIONA ===")
	fmt.Printf("Data uruchomienia: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Autor: %s\n", author)
	fmt.Printf("Port: %s\n", port)
	fmt.Println("============================")

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/weather", handleWeather)
	http.HandleFunc("/health", handleHealth)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	var options strings.Builder
	for key, city := range cities {
		options.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, key, city.Name))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Aplikacja Pogodowa</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 0 auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 20px; border-radius: 10px; }
        select, button { width: 100%%; padding: 10px; margin: 10px 0; }
        button { background: #007bff; color: white; border: none; cursor: pointer; }
        #result { margin-top: 20px; display: none; }
        .weather-info { background: white; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Sprawdź Pogodę</h1>
        <select id="city">
            <option value="">Wybierz miasto...</option>
            %s
        </select>
        <button onclick="getWeather()">Sprawdź pogodę</button>
        <div id="result">
            <h2 id="cityName"></h2>
            <div class="weather-info">
                <p>Temperatura: <span id="temp"></span>°C</p>
                <p>Wilgotność: <span id="humidity"></span>%%</p>
                <p>Prędkość wiatru: <span id="wind"></span> km/h</p>
            </div>
        </div>
    </div>
    <script>
        function getWeather() {
            const city = document.getElementById('city').value;
            if (!city) return;
            
            fetch('/weather?city=' + city)
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').style.display = 'block';
                    document.getElementById('cityName').textContent = data.city;
                    document.getElementById('temp').textContent = data.temperature;
                    document.getElementById('humidity').textContent = data.humidity;
                    document.getElementById('wind').textContent = data.wind_speed;
                });
        }
    </script>
</body>
</html>`, options.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	cityKey := r.URL.Query().Get("city")
	city, ok := cities[cityKey]
	if !ok {
		http.Error(w, "City not found", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("http://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&hourly=relative_humidity_2m", city.Lat, city.Lon)
	
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var weatherData map[string]interface{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	currentWeather := weatherData["current_weather"].(map[string]interface{})
	hourlyData := weatherData["hourly"].(map[string]interface{})
	humidityArray := hourlyData["relative_humidity_2m"].([]interface{})

	result := map[string]interface{}{
		"city":        city.Name,
		"temperature": currentWeather["temperature"],
		"humidity":    humidityArray[0],
		"wind_speed":  currentWeather["windspeed"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}