# RadarCampinas - Machine Learning Service

Este projeto é um serviço em **FastAPI** que treina um modelo de classificação usando dados de bairros e salva previsões no banco de dados **MySQL** (anteriormente PostgreSQL).  

O modelo utiliza **RandomForestClassifier** do **scikit-learn**.

---

## Estrutura do projeto

machine_learning_service/
│
├── app.py # Aplicação FastAPI principal
├── requirements.txt # Dependências do projeto
├── utils/
│ ├── db.py # Funções de conexão e operações no banco
│ └── preprocess.py # Função para preparar features e target
└── venv/ # Ambiente virtual

yaml
Copiar código

---

## Dependências

Inclua no seu `requirements.txt`:

fastapi
uvicorn
scikit-learn
pandas
mysql-connector-python

go
Copiar código

Instale usando:

```bash
pip install -r requirements.txt
Configuração do Banco de Dados
Certifique-se de estar na rede da universidade.

Crie um banco MySQL com as seguintes credenciais (ajuste se necessário):

python
Copiar código
host="regulus.cotuca.unicamp.br"
port=3306
user="BD24452"
password="BD24452"
database="BD24452"
Certifique-se de que a tabela para previsões existe:

sql
Copiar código
CREATE TABLE predictCrime (
    neighborhood VARCHAR(255),
    risk_level INT
);
Altere o nome da tabela na função get_base_knowledge() para a tabela que contém sua base de conhecimento.

Configuração do Python
Crie um ambiente virtual:

bash
Copiar código
python3 -m venv venv
source venv/bin/activate  # Linux/macOS
venv\Scripts\activate     # Windows
Instale as dependências:

bash
Copiar código
pip install -r requirements.txt
Executando o Serviço
Inicie o FastAPI com Uvicorn:

bash
Copiar código
uvicorn app:app --reload --port 8000
O serviço estará disponível em:

cpp
Copiar código
http://127.0.0.1:8000
Endpoints
Treinar o modelo e salvar previsões no banco

POST /training

Retorna JSON com mensagem e número de linhas processadas:

json
Copiar código
{
  "message": "Treinamento concluído",
  "rows": 100
}
Logs
Logs detalhados sobre o processo de treinamento e salvamento de previsões são exibidos no terminal.

Incluem: início do treinamento, número de linhas carregadas, sucesso ou falha de cada etapa.

Observações
O serviço espera que você tenha acesso à rede da universidade para conectar ao banco de dados MySQL.

Certifique-se de que a porta correta (3306) está aberta.

Se houver problemas de conexão, verifique se consegue acessar o banco pelo DBeaver usando as mesmas credenciais.

O endpoint /training deve ser chamado via POST, não GET.

Exemplo de uso via curl:
bash
Copiar código
curl -X POST http://127.0.0.1:8000/training