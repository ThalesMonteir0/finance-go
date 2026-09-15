package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"gofinance/dto/filters"
	"gofinance/dto/ia"
	"gofinance/dto/transaction"
	"io"
	"mime/multipart"
	"os"
)

type pdfService interface {
	Read(path string) (io.Reader, error)
	ParseItauStatement(reader io.Reader, filters filters.Filter) ([]transaction.Transaction, float64, error)
}

type transactionRepository interface {
	Save(ctx context.Context, transaction []transaction.Transaction) error
}

type IAClient interface {
	ChatCompletion(ctx context.Context, req ia.ChatCompletionRequest) (*ia.ChatCompletionResponse, error)
}

type ReadStatementUseCase struct {
	pdfSvc                pdfService
	transactionRepository transactionRepository
	iaClient              IAClient
}

func NewReadStatementUseCase(pdfSvc pdfService, iaClient IAClient) *ReadStatementUseCase {
	return &ReadStatementUseCase{
		pdfSvc: pdfSvc,
		//transactionRepository: transactionRepo,
		iaClient: iaClient,
	}
}

func (c *ReadStatementUseCase) Execute(ctx context.Context, file multipart.File, filters filters.Filter) (*transaction.TransactionsResponse, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fileTemp, err := os.CreateTemp("./file", "*.pdf")
	if err != nil {
		return nil, err
	}
	defer fileTemp.Close()
	defer os.Remove(fileTemp.Name())

	_, err = io.Copy(fileTemp, file)
	if err != nil {
		return nil, err
	}

	reader, err := c.pdfSvc.Read(fileTemp.Name())
	if err != nil {
		return nil, err
	}

	transactions, totalAmountOut, err := c.pdfSvc.ParseItauStatement(reader, filters)
	if err != nil {
		return nil, err
	}

	var outTransactions []transaction.Transaction

	for _, transactionX := range transactions {
		if transactionX.Type == "OUT" {
			outTransactions = append(outTransactions, transactionX)
		}
	}

	transactionsBytes, err := json.Marshal(outTransactions)
	if err != nil {
		return nil, err
	}

	body := ia.ChatCompletionRequest{
		Model: "deepseek-v4-flash",
		Messages: []ia.Message{
			{
				Role: "user",
				Content: "Você é um classificador de transações financeiras.Sua tarefa é identificar, para cada transação fornecida, qual é o estabelecimento/empresa real associado à descrição do pagamento e, em seguida, classificar esse estabelecimento em um tipo de gasto.IMPORTANTE:- Sempre que possível, PESQUISE NA INTERNET o estabelecimento correspondente à descrição da transação.- Não se limite a inferir o tipo apenas pelo texto da descrição.- A descrição pode estar abreviada, incompleta ou conter apenas o nome comercial, razão social ou identificador do recebedor.- Quando encontrar o estabelecimento real, use o nome pelo qual ele é conhecido comercialmente e identifique seu tipo.- Exemplo: 'RD SAUDE' pode corresponder à Raia Drogasil/Drogasil → tipo FARMACIA.- Exemplo: 'AMAZON.COM' → MARKETPLACE.- Exemplo: 'SHOPEE' → MARKETPLACE.- Exemplo: 'TIKTOK SHOP → MARKETPLACE.- Exemplo: 'JOSE CA' quando for uma transferência para uma pessoa física → TRANSFERENCIA PESSOAL.- Não classifique pela forma de pagamento (PIX, cartão etc.), mas pelo estabelecimento que recebeu o dinheiro.### Prioridade dos tiposDê preferência aos seguintes tipos quando forem aplicáveis:MARKETPLACE SUPERMERCADO DELIVERY CAFETERIA CINEMA FARMACIA RESTAURANTE POSTO DE COMBUSTIVEL TRANSPORTE VESTUARIO ELETRONICOS ACADEMIA ENTRETENIMENTO SERVICOSEDUCACAO SAUDE TRANSFERENCIA PESSOAL OUTROS A lista não é fechada. Se o estabelecimento claramente pertencer a uma categoria que não esteja na lista, crie uma categoria adequada.Se não for possível identificar com segurança o estabelecimento ou seu tipo mesmo após pesquisa, use: 'estabelecimento':  'INDETERMINADO ' ### Regras importantes1. Analise TODAS as transações fornecidas.2. Para cada transação, identifique primeiro o estabelecimento real.3. Depois determine o tipo de estabelecimento.4. Use o valor absoluto da transação para calcular a porcentagem. Ignore o sinal negativo dos gastos.5. Some todos os gastos para obter o total.6. Agrupe os valores pelo TIPO de estabelecimento.7. Calcule a porcentagem que cada tipo representa do total gasto:      porcentagem = (total gasto naquele tipo / total geral gasto) × 1008. Arredonde a porcentagem para 2 casas decimais.9. A soma das porcentagens deve ser aproximadamente 100%.10. Não crie uma categoria diferente para cada estabelecimento. O campo 'estabelecimento' do resultado deve representar o TIPO DE ESTABELECIMENTO, e não o nome da empresa.11. Se houver vários estabelecimentos do mesmo tipo, some todos antes de calcular a porcentagem.### ExemploEntrada:PIX QRS AMAZON.COM.10/09 -10.00PIX TRANSF JOSE CA09/09 -104.13PIX QRS RD SAUDE10/09 -242.34PIX QRS BYTEDANCE B10/09 -44.06Interpretação esperada:- AMAZON.COM → Amazon → MARKETPLACE → R$ 10,00- JOSE CA → pessoa física → TRANSFERENCIA PESSOAL → R$ 104,13- RD SAUDE → Raia Drogasil/Drogasil → FARMACIA → R$ 242,34- BYTEDANCE B → TikTok/TikTok Shop → MARKETPLACE → R$ 44,06Total = R$ 400,53MARKETPLACE = R$ 54,06 → 13,50%TRANSFERENCIA PESSOAL = R$ 104,13 → 26,00%FARMACIA = R$ 242,34 → 60,50%### Formato obrigatório da resposta Retorne SOMENTE UAM STRING DE um JSON válido, sem explicações, sem markdown ou texto adicional.O JSON deve seguir exatamente este formato:[  {    'porcentagem': 13.50,     'estabelecimento':  'MARKETPLACE'  },  { 'porcentagem': 26.00,    'estabelecimento': 'TRANSFERENCIA PESSOAL'  },  {    'porcentagem': 60.50,    'estabelecimento': 'FARMACIA'  }]Não retorne os estabelecimentos individuais no JSON final. Retorne somente os tipos agrupados e suas respectivas porcentagens. INDENTIFIQUE:" +
					string(transactionsBytes) + "DEVOLVA UMA STRING, APENAS UMA STRING, NAO UTULIZE MARKDOWN!!",
			},
		},
	}

	result, err := c.iaClient.ChatCompletion(ctx, body)
	if err != nil {
		return nil, err
	}

	fmt.Println(result.Choices[0].Message.Content)

	var a []ia.EstablishmentPercentage
	if err = json.Unmarshal([]byte(result.Choices[0].Message.Content), &a); err != nil {
		return nil, err
	}

	response := transaction.TransactionsResponse{
		Transactions:               transactions,
		PercentagePerEstablishment: a,
		TotalAmountOut:             totalAmountOut,
	}

	return &response, nil
}
