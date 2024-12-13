package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// genero, primer nombre,
// primer apellido, email, ciudad,pais y uuid
// 15,000 solicitudes
// Menos de 2.25 segundos
// distintos
func getUser(c *gin.Context) {
	resp, err := http.Get("https://randomuser.me/api/?results=5000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)

}

func main() {
	r := gin.Default()
	r.GET("/user", getUser)

	err := r.Run("127.0.0.1:8000")
	if err != nil {
		fmt.Printf("Error")
		return
	}

	fmt.Printf("Server Started")

}