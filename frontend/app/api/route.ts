// app/api/chat/route.ts (usando axios)
import { NextResponse } from 'next/server';
import axios from 'axios';

const prompt = `
Você é um chatbot humanizado que recebe denúncias de crimes em linguagem natural.

O usuário fornecerá informações em texto livre. Sua função é:

1. Extrair os seguintes dados:
   - tipo_de_crime
   - data_da_denuncia (ou data do crime)
   - localização (em texto)

2. Caso alguma dessas informações esteja faltando, pergunte educadamente ao usuário até que todos os dados estejam completos.

3. Após ter a localização textual, converta-a em coordenadas (latitude e longitude). Use um serviço de geocodificação, se necessário.

4. Antes de encerrar, mostre um resumo dos dados coletados e peça confirmação ao usuário.

 IMPORTANTE:
- Somente após o usuário confirmar (enviar sim), gere e retorne APENAS UM JSON com os seguintes campos:
{
  "latitude": "<valor>",
  "longitude": "<valor>",
  "tipo_de_crime": "<valor>",
  "data_da_denuncia": "<valor>"
}
- NÃO retorne texto explicativo, markdown ou frases antes/depois do JSON. Apenas o objeto JSON puro na resposta final.

 Exemplo de resposta final esperada:
{
  "latitude": "-23.550520",
  "longitude": "-46.633308",
  "tipo_de_crime": "roubo",
  "data_da_denuncia": "2025-05-10"
}
`;

export async function POST(req: Request) {
  try {
    const { denuncia } = await req.json();
    
    if (!process.env.OPENAI_API_KEY) {
      console.error("API key não encontrada");
      return NextResponse.json(
        { error: 'Configuração da API OpenAI ausente' }, 
        { status: 500 }
      );
    }

    console.log("DENUNCIA:", denuncia);
    console.log("API_KEY presente:", process.env.OPENAI_API_KEY);
    
    // Endpoint correto para a API atual
    const response = await axios.post(
      'https://api.openai.com/v1/chat/completions',
      {
        model: 'gpt-4o',
        messages: [
          {
            role: 'system',
            content: prompt
          },
          {
            role: 'user',
            content: denuncia,
          }
        ],
        temperature: 0.2,
      },
      {
        headers: {
          'Authorization': `Bearer ${process.env.OPENAI_API_KEY}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    const result = response.data.choices[0].message.content;
    console.log('Resposta da OpenAI:', result);
    
    return NextResponse.json({ resultado: result });
  } catch (error: any) {
    console.error('Erro ao chamar OpenAI:', error);
    
    // Exibir detalhes do erro para depuração
    if (error.response) {
      console.error('Status:', error.response.status);
      console.error('Data:', error.response.data);
    }
    
    return NextResponse.json(
      { error: 'Erro ao processar denúncia', detalhes: error.message }, 
      { status: 500 }
    );
  }
}