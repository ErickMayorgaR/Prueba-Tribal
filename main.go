package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// genero, primer nombre,primer apellido, email, ciudad,pais y uuid
// 15,000 solicitudes
// Menos de 2.25 segundos
// distintos

// Se agregaron estructuras para manejar de manera eficiente la conversion a los tipados de datos
type User struct {
	Gender  string `json:"gender"`
	First   string `json:"first"`
	Last    string `json:"last"`
	Email   string `json:"email"`
	City    string `json:"city"`
	Country string `json:"country"`
	UUID    string `json:"uuid"`
}

type APIResponse struct {
	Results []struct {
		Gender string `json:"gender"`
		Name   struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
		Email string `json:"email"`
		Login struct {
			UUID string `json:"uuid"`
		} `json:"login"`
	} `json:"results"`
}

func fetchUsers(ch chan<- []User, url string, wg *sync.WaitGroup, client *http.Client) {
	defer wg.Done()
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	//Se realiza la decodificacion de Json unicamente a las estructuras relacionadas a los elementos requeridos en la peticion
	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}
	//Se accede a los datos especificos de genero, primer nombre,primer apellido, email, ciudad,pais y uuid que se necesitan y se guardan en estructura User
	var users []User
	for _, result := range apiResp.Results {
		users = append(users, User{
			Gender:  result.Gender,
			First:   result.Name.First,
			Last:    result.Name.Last,
			Email:   result.Email,
			City:    result.Location.City,
			Country: result.Location.Country,
			UUID:    result.Login.UUID,
		})
	}
	ch <- users
}

func getUser(c *gin.Context) {
	//Se realizan 300 peticiones en go routines de 50 elementos en vez de 3 de 5000
	const (
		totalRequests     = 300
		resultsPerRequest = 50
		apiURLTemplate    = "https://randomuser.me/api/?results=%d"
	)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var wg sync.WaitGroup
	ch := make(chan []User, totalRequests)
	url := fmt.Sprintf(apiURLTemplate, resultsPerRequest)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go fetchUsers(ch, url, &wg, client)
	}

	wg.Wait()
	close(ch)

	// Se itera sobre los rangos de respuesta y se verifica que solamente se devuelvan usuarios no repetidos
	uniqueUsers := make(map[string]User)
	for users := range ch {
		for _, user := range users {
			if _, exists := uniqueUsers[user.UUID]; !exists {
				uniqueUsers[user.UUID] = user
			}
		}
	}
	finalUsers := make([]User, 0, len(uniqueUsers))
	for _, user := range uniqueUsers {
		finalUsers = append(finalUsers, user)
	}

	c.JSON(http.StatusOK, finalUsers)
}

func main() {
	r := gin.Default()

	r.GET("/user", getUser)

	err := r.Run(":8000")
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
