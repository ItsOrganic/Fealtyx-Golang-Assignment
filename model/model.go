package model

type StudentDetails struct {
	Id    int    `json:"id" binding:"required" `
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type LLAMA3 struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
type LLAMA3Response struct {
	Response string `json:"response"`
}
