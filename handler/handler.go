package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/ItsOrganic/FealtyX-GoLang_Assignment/model"
	"github.com/ItsOrganic/FealtyX-GoLang_Assignment/regex"
	"github.com/gin-gonic/gin"
)

type StudentHandler interface {
	CreateStudent(c *gin.Context)
	GetStudent(c *gin.Context)
	GetStudents(c *gin.Context)
	UpdateStudent(c *gin.Context)
	DeleteStudent(c *gin.Context)
	Summary(c *gin.Context)
}

var (
	student []model.StudentDetails
	mutex   sync.RWMutex
)

type Handler struct{}

// Creating the student method="POST"
func (h *Handler) CreateStudent(c *gin.Context) {
	var newStudent model.StudentDetails

	if err := c.ShouldBindJSON(&newStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// WRite lock
	mutex.Lock()
	defer mutex.Unlock()

	// Check if the student ID already exists
	for _, value := range student {
		if value.Id == newStudent.Id {
			c.JSON(http.StatusConflict, gin.H{"error": "Student ID already exists"})
			return
		}
	}
	if !regex.VerifyEmail(newStudent.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	student = append(student, newStudent)
	c.JSON(http.StatusOK, gin.H{"success": student})
}

// Getting the detail of single student method="GET"
func (h *Handler) GetStudent(c *gin.Context) {
	id := c.Param("id")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Read lock
	mutex.RLock()
	defer mutex.RUnlock()

	for _, value := range student {
		if value.Id == studentID {
			c.JSON(http.StatusOK, gin.H{"success": value})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

// Getting the details of all student method="GET"
func (h *Handler) GetStudents(c *gin.Context) {
	if len(student) < 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No student found"})
		return
	}

	// Read lock
	mutex.RLock()
	defer mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{"success": student})
}

// Updating the student details method="PUT"
func (h *Handler) UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	var updatedStudent model.StudentDetails
	if err := c.ShouldBindJSON(&updatedStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Write lock
	mutex.Lock()
	defer mutex.Unlock()

	updatedStudent.Id = studentID
	for index, value := range student {
		if value.Id == studentID {
			if updatedStudent.Name != "" {
				student[index].Name = updatedStudent.Name
			}
			if updatedStudent.Age != 0 {
				student[index].Age = updatedStudent.Age
			}
			if updatedStudent.Email != "" {
				if !regex.VerifyEmail(updatedStudent.Email) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
					return
				}
				student[index].Email = updatedStudent.Email
			}
			c.JSON(http.StatusOK, gin.H{"success": "Student updated"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

// Deleting the student method="DELETE"
func (h *Handler) DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Write lock
	mutex.Lock()
	defer mutex.Unlock()

	for index, value := range student {
		if value.Id == studentID {
			student = append(student[:index], student[index+1:]...)
			c.JSON(http.StatusOK, gin.H{"success": "Student deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

func (h *Handler) Summary(c *gin.Context) {
	id := c.Param("id")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Read lock
	mutex.RLock()

	// Find the student by ID
	var currentStudent model.StudentDetails
	found := false
	for _, value := range student {
		if value.Id == studentID {
			currentStudent = value
			found = true
			break
		}
	}

	mutex.RUnlock()

	// Return 404 if student not found
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Create a prompt for the student summary
	prompt := fmt.Sprintf("Summerize student with name=%s, age=%d, email=%s in not more than 3 small sentences about the student.(Don't use these things Here is a summary of, dont use any formatter)",
		currentStudent.Name, currentStudent.Age, currentStudent.Email)

	// Construct the Ollama request payload
	ollamaReq := model.LLAMA3{
		Model:  "llama3",
		Prompt: prompt,
		Stream: false,
	}

	// Marshal the request payload to JSON
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request payload"})
		return
	}

	// Send the POST request to Ollama API
	url := "http://localhost:11434/api/generate"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
		return
	}
	defer resp.Body.Close()

	// Check for non-OK status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ollama API returned error: %s", string(bodyBytes))})
		return
	}

	// Decode the response body directly into OllamaResponse struct
	var ollamaResp model.LLAMA3Response
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Println("Error decoding response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Ollama API response"})
		return
	}

	// Check and log the entire parsed response for debugging
	log.Printf("Parsed Ollama response: %+v\n", ollamaResp)

	// Return the parsed response (or the specific field) in JSON format to Postman
	c.JSON(http.StatusOK, gin.H{"summary": ollamaResp.Response})
}
