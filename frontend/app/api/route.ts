// app/api/chat/route.ts (usando axios)
import { NextResponse } from 'next/server';
import axios from 'axios';

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
            content: 'Extraia latitude, longitude, tipo de crime e data de denúncias em formato JSON.',
          },
          {
            role: 'user',
            content: denuncia,
          }
        ],
        temperature: 0.7,
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