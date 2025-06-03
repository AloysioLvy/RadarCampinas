import { NextResponse } from "next/server";
import axios from "axios";


const SYSTEM_PROMPT = `
Você é um chatbot responsável por coletar denúncias de crimes.

Sua função é:
1. Extrair os seguintes dados:
   - tipo_de_crime
   - data_da_denuncia (formato DD/MM/AAAA)
   - localizacao (texto)

2. Se algum dado estiver faltando, pergunte educadamente até coletar todos.

3. Quando todos os dados estiverem completos, envie um resumo para o usuário, confirmando:
- Tipo de crime
- Data
- Localização

Pergunte: "Está correto? (responda sim ou não)"

4. Se o usuário confirmar com "sim", então retorne APENAS um JSON puro no seguinte formato:
{
  "tipo_de_crime": "<valor>",
  "data_crime": "<valor>",
  "localizacao": "<valor>"
}

❌ Não inclua nenhuma outra palavra, explicação ou markdown. Apenas o JSON puro.

Se o usuário responder "não", reinicie o processo de coleta.
`;

const crimesHediondos = [
  "homicídio qualificado",
  "latrocínio",
  "extorsão mediante sequestro",
  "estupro",
  "estupro de vulnerável",
  "epidemia com resultado morte",
  "falsificação de medicamento",
  "falsificação de produto terapêutico",
  "genocídio",
  "posse ou porte ilegal de arma de fogo de uso restrito",
  "tráfico de drogas",
  "tortura",
  "financiamento ao tráfico de drogas",
  "assassinato em grupo de extermínio",
  "violência sexual mediante fraude",
  "exploração sexual de criança ou adolescente",
  "favorecimento à prostituição de criança ou adolescente",
  "invasão de dispositivo informático com finalidade sexual envolvendo criança ou adolescente"
];

function calcularPesoCrime(tipoDeCrime: string): number {
  const crimeNormalizado = tipoDeCrime.trim().toLowerCase();
  const isHediondo = crimesHediondos.some(
    (crime) => crime.toLowerCase() === crimeNormalizado
  );
  return isHediondo ? 9 : 3;
}

async function geocodeAddress(address: string) {
  try {
    const url = "https://nominatim.openstreetmap.org/search";
    const response = await axios.get(url, {
      params: {
        q: address,
        format: "json",
        limit: 1,
      },
      headers: {
        "User-Agent": "MeuAppDenuncias/1.0",
      },
    });
    if (response.data.length === 0) {
      return null;
    }
    const { lat, lon } = response.data[0];
    console.log(response.data[0])
    return { 
      latitude: lat, longitude: lon };
  } catch (error) {
    console.error("Erro na geocodificação:", error);
    return null;
  }
}

export async function POST(req: Request) {
  try {
    const { messages } = await req.json();
    console.log("resposta do usuario = " + messages)

    if (!process.env.OPENAI_API_KEY) {
      console.error("API key não encontrada");
      return NextResponse.json(
        { error: "Configuração da API OpenAI ausente" },
        { status: 500 }
      );
    }

    const openaiResponse = await axios.post(
      "https://api.openai.com/v1/chat/completions",
      {
        model: "gpt-4o",
        messages: [{ role: "system", content: SYSTEM_PROMPT }, ...messages],
        temperature: 0.2,
      },
      {
        headers: {
          Authorization: `Bearer ${process.env.OPENAI_API_KEY}`,
          "Content-Type": "application/json",
        },
      }
    );

    const result = openaiResponse.data.choices[0].message.content;
    console.log("Resposta OpenAI:", result);

    let finalData = null;
    console.log("result = " + result)
    try {
      finalData = JSON.parse(result);
      console.log("final data = " + finalData)
    } catch {
      console.log("entrou no catch")
    }

    if (finalData && finalData.localizacao) {
      const pesoCrime = calcularPesoCrime(finalData.tipo_de_crime);
      console.log("Crime"+finalData.tipo_de_crime)
      const coords = await geocodeAddress(finalData.localizacao);
      console.log(JSON.stringify(finalData))


      const payload = {
        ...finalData,
        peso_crime: pesoCrime,
        latitude: coords?.latitude || null,
        longitude: coords?.longitude || null,
      };
      console.log("PAYLOAD"+JSON.stringify(payload))
      try {
        const backendResponse = await axios.post(
          "http://localhost:1023/api/v1/denuncias/texto",
          payload
        );
        
        console.log("Resposta do backend:", backendResponse.data);
      
        return NextResponse.json({
          sucesso: true,
          dados_enviados: payload,
          resposta_backend: backendResponse.data,
        });
      
      } catch (backendError: any) {
        console.error("Erro ao enviar para o backend:", backendError);
        return NextResponse.json(
          { erro: "Falha ao enviar dados ao backend", detalhes: backendError.message },
          { status: 500 }
        );
      }
      
    }

    // Caso ainda não tenha confirmado (JSON não foi gerado)
    return NextResponse.json({ resultado: result });
  } catch (error: any) {
    console.error("Erro ao chamar OpenAI:", error);
    if (error.response) {
      console.error("Status:", error.response.status);
      console.error("Data:", error.response.data);
    }
    return NextResponse.json(
      { error: "Erro ao processar denúncia", detalhes: error.message },
      { status: 500 }
    );
  }
}
