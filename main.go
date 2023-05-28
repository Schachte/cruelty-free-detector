package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const OPEN_AI_API_KEY = ""

type OpenAIParams struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClientRequest struct {
	Query string `json:"query"`
}

type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var clientReq ClientRequest
	err := json.NewDecoder(r.Body).Decode(&clientReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create system message to guide GPT-3.5-Turbo
	systemMessage := Message{
		Role: "system",
		Content: "Play the role of an API that only returns JSON. The format of the JSON looks like the following:\n\n" +
			"{\n" +
			"\"company_name\": string\n" +
			"\"parent_company\": string\n" +
			"\"cruelty_free\": boolean\n" +
			"\"offenses\": string[]\n" +
			"\"parent_company_cruelty_free\": boolean\n" +
			"}\n\n" +
			"company_name is the company that owns the product or the name of a company\n" +
			"parent_company is the name of the parent company that owns the company_name if relevant\n" +
			"cruelty_free denotes whether or not the product is animal cruelty free or not\n" +
			"offenses is an array of known offenses or descriptions of why the company is not cruelty free\n" +
			"parent_company_cruelty_free is a boolean that represents if the company is owned by another company if the parent company is considered cruelty free.\n" +
			"add another key for all future responses called \"sells_products_tested_on_animals\" that is a boolean representing if the company or parent company sells any products that are not cruelty free\n\n" +
			"there should be a JSON key called \"alternatives\" that is an array listing alternative companies in the same domain that are cruelty free. Only list alternatives for products that are not cruelty free.\n\n" +
			"make sure all responses only include the JSON and nothing else for now on.\n\n" +
			"Do not ever output anything that is not JSON format. Everything returned should be valid JSON and pass a JSON parser.\n\n" +
			"The response should conform to JSON that abides by the following schema:\n\n" +
			"const schema = Joi.object({\n" +
			"\"company_name\": Joi.string().required(),\n" +
			"\"parent_company\": Joi.string().required(),\n" +
			"\"cruelty_free\": Joi.boolean().required(),\n" +
			"\"offenses\": Joi.array().items(Joi.string()).required(),\n" +
			"\"parent_company_cruelty_free\": Joi.boolean().required(),\n" +
			"\"sells_products_tested_on_animals\": Joi.boolean().required()\n" +
			"});",
	}

	userMessage := Message{
		Role:    "user",
		Content: clientReq.Query,
	}

	params := OpenAIParams{
		Messages: []Message{systemMessage, userMessage},
		Model:    "gpt-3.5-turbo",
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPEN_AI_API_KEY))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		http.Error(w, "Invalid response", http.StatusInternalServerError)
		return
	}

	if len(response.Choices) == 0 {
		http.Error(w, "Invalid response", http.StatusInternalServerError)
		return
	}

	contentJSON := response.Choices[0].Message.Content
	if contentJSON == "" {
		http.Error(w, "Invalid response", http.StatusInternalServerError)
		return
	}

	var content interface{}
	err = json.Unmarshal([]byte(contentJSON), &content)
	if err != nil {
		http.Error(w, "Invalid response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
