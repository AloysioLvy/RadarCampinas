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

const HeinousCrimes = [
      "latrocínio",
      "homicídio qualificado",
      "homicídio praticado por grupo de extermínio",
      "feminicídio",
      "genocídio",
      "estupro",
      "estupro de vulnerável",
      "atentado violento ao pudor",
      "favorecimento à prostituição",
      "exploração sexual",
      "tráfico de pessoas",
      "tráfico de drogas",
      "organização criminosa",
      "comércio ilegal de armas",
      "extorsão qualificada",
      "sequestro e cárcere privado",
      "extorsão mediante seqüestro",
      "envenenamento de alimentos",
      "epidemia com resultado morte",
      "falsificação de medicamentos",
      "tráfico internacional de armas",
      'Sequestro e extorsão qualificada'
];
function calculateWeightCrime(typeOfCrime:string):number {
  const crimeNormalized = typeOfCrime.trim().toLowerCase();
  const isHeinous = HeinousCrimes.some(crime => crime.toLowerCase().includes(crimeNormalized));
  return isHeinous ? 9 : 3;
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
    console.log(response.data)
    const { lat, lon } = response.data[0];
    const display_name = response.data[0].display_name;
    const neighborhood = display_name.split(",")
    return { 
      latitude: lat, longitude: lon,  localizacao:neighborhood[1]};
      
  } catch (error) {
    console.error("Geocoding error:", error);
    return null;
  }
}

export async function POST(req: Request) {
  try {
    const { messages } = await req.json();
    
    if (!process.env.OPENAI_API_KEY) {
      console.error("API key not found");
      return NextResponse.json(
        { error: "Missing OpenAI API configuration " },
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
    
    try {
      console.log("final data = " + finalData)
      finalData = JSON.parse(result);
    } catch {
      console.log("entrou no catch")
    }

    if (finalData && finalData.localizacao) {
      const pesoCrime = calculateWeightCrime
      (finalData.tipo_de_crime);
      const coords = await geocodeAddress(finalData.localizacao);
      


      const payload = {
        ...finalData,
        peso_crime: pesoCrime,
        latitude: coords?.latitude || null,
        longitude: coords?.longitude || null,
        localizacao: coords?.localizacao||null,
      };
      console.log("PAYLOAD"+JSON.stringify(payload))
      try {
        if(!process.env.BACKEND_ROUTE_URL){
          console.error("BACKEND ROUTE URL not found");
        }
        const backendResponse = await axios.post(
          `${process.env.BACKEND_ROUTE_URL}`,
          payload
        
        );
        
        console.log("Backend response:", backendResponse.data);
      
        return NextResponse.json({
          sucess: true,
          data_sent: payload,
          backend_response: backendResponse.data,
        });
      
      } catch (backendError: any) {
        console.error("Error sending to backend:", backendError);
        return NextResponse.json(
          { erro: "Failed to send data to backend", detalhes: backendError.message },
          { status: 500 }
        );
      }
      
    }

    return NextResponse.json({ result: result });
  } catch (error: any) {
    console.error("Error calling OpenAI:", error);
    if (error.response) {
      console.error("Status:", error.response.status);
      console.error("Data:", error.response.data);
    }
    return NextResponse.json(
      { error: "Error processing report", detalhes: error.message },
      { status: 500 }
    );
  }
}
