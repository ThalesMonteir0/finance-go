package ia

type EstablishmentPercentage struct {
	Porcentagem     float64 `json:"porcentagem"`
	Estabelecimento string  `json:"estabelecimento"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiGenerateContentRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type GeminiGenerateContentResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}
