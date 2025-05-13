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
  Aqui está o resumo dos dados coletados:
- Tipo de crime: assalto
- Data da denúncia: 11/11/2022
- Localização: Rua Antonio Volpe, nº 386
- Latitude: -23.550520
- Longitude: -46.633308
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
    const denunicaCrime = {
      nome: "",
      latitude: "",
      longitude: "",
      tipo_de_crime:"" ,
      data_da_denuncia: "",
      peso_crime: 3
    };

    const result = response.data.choices[0].message.content;
    console.log('Resposta da OpenAI:', result);
    

    if (result.includes("resumo dos dados coletados:")){
      const tipoCrimeMatch = result.match(/Tipo de crime:\s*(.+)/i);
      const dataMatch = result.match(/Data da denúncia:\s*(\d{2}\/\d{2}\/\d{4})/i);
      const localMatch = result.match(/Localização:\s*(.+)/i);
      const latMatch = result.match(/Latitude:\s*(-?\d+\.\d+)/);
      const longMatch = result.match(/Longitude:\s*(-?\d+\.\d+)/);
      denunicaCrime.nome = localMatch[1]
      denunicaCrime.latitude = latMatch[1]
      denunicaCrime.longitude = longMatch[1]
      denunicaCrime.data_da_denuncia = dataMatch[1]
      denunicaCrime.peso_crime = 3
      denunicaCrime.tipo_de_crime = tipoCrimeMatch[1]
      console.log("-------------->A DENUNCIA DEU CERTO <---------")
      console.log("{"+denunicaCrime.latitude+"}")
      console.log("{"+denunicaCrime.longitude+"}")
      console.log("{"+denunicaCrime.tipo_de_crime+"}")
      console.log("{"+denunicaCrime.peso_crime+"}")
      console.log("{"+denunicaCrime.data_da_denuncia+"}")
    }
    const denunciaJSON = JSON.stringify(denunicaCrime)
    console.log("DENUNCIA JSON:\n"+denunciaJSON)
    
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