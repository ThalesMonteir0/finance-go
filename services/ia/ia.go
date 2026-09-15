package ia

import (
	"context"
	"encoding/json"
	"fmt"
	dtoia "gofinance/dto/ia"
	"gofinance/dto/transaction"
	"strings"
)

const basePrompt = `identifique o tipo de estabelecimento desse pagamento através da descrição que estou lhe passando. Os tipos podem ser: MARKETPLACE, SUPERMERCADO, DELIVERY, CAFETERIA, CINEMA, TRANSFERENCIA PESSOAL (para quando for pix para pessoa fisica) e entre outros, mas sempre dê prioridade para esses tipos de estabelecimento. Caso você não conheça o tipo bote 'INDETERMINADO'. Para plataformas como tiktok shop, shopee e etc sempre é MARKETPLACE. Não precisa seguir 100%% a lista que passei, apenas dei exemplos de como proceder, caso encontre o tipo do estabelecimento e não está na lista coloque-o. Após identificar, quero que você some tudo e que você me devolva em json seguindo o seguinte objeto: [{"porcentagem": float64, "estabelecimento": string}], no campo porcentagem coloque a porcentagem por tipo de estabelecimento em que o meu dinheiro foi gasto. Por exemplo: 20%% do todo foi gasto em FARMACIA(estabelecimento). Responda apenas com o JSON (um array do objeto acima), sem markdown e sem texto adicional. - identifique: %s`

type iaClient interface {
	GenerateContent(ctx context.Context, req dtoia.GeminiGenerateContentRequest) (*dtoia.GeminiGenerateContentResponse, error)
}

type Service struct {
	client iaClient
}

func New(client iaClient) *Service {
	return &Service{client: client}
}

func (s *Service) ClassifyEstablishments(ctx context.Context, transactions []transaction.Transaction) ([]dtoia.EstablishmentPercentage, error) {
	prompt := buildPrompt(transactions)

	resp, err := s.client.GenerateContent(ctx, dtoia.GeminiGenerateContentRequest{
		Contents: []dtoia.GeminiContent{
			{Parts: []dtoia.GeminiPart{{Text: prompt}}},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("ia: nenhuma resposta retornada pelo modelo")
	}

	content := extractJSON(resp.Candidates[0].Content.Parts[0].Text)

	var result []dtoia.EstablishmentPercentage
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("ia: falha ao interpretar resposta do modelo: %w", err)
	}

	return result, nil
}

func buildPrompt(transactions []transaction.Transaction) string {
	lines := make([]string, 0, len(transactions))
	for _, t := range transactions {
		lines = append(lines, fmt.Sprintf("%s %.2f", t.Description, t.Amount))
	}

	return fmt.Sprintf(basePrompt, strings.Join(lines, ", "))
}

// extractJSON remove eventuais cercas de markdown (```json ... ```) que o modelo às vezes
// usa para envolver a resposta, deixando apenas o JSON puro para o Unmarshal.
func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}
