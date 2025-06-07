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
  const correctAdderss = address.trim();
  try { 
    const url = "https://nominatim.openstreetmap.org/search";
    const response = await axios.get(url, {
      params: {
        q: correctAdderss,
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
    const displayName = response.data[0].display_name;
    const neighborhoodName = displayName.split(",");
    const cityName = neighborhoodName[2];
   
    return { 
      latitude: lat, longitude: lon,  neighborhoodName:neighborhoodName[1].trim(),
    cityName: cityName.trim()};
      
  } catch (error) {
    console.error("Geocoding error:", error);
    return null;
  }
}

export async function POST(req: Request) {
  try {
    const { messages } = await req.json();
    console.log("User: ",messages)
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



    let finalData = null;
    
    try {
      finalData = JSON.parse(result);
    } catch {
    
    }
    
    if (finalData && finalData.localizacao != null) {
      const crime_weight = calculateWeightCrime(finalData.tipo_de_crime);
      const locationInfo = await geocodeAddress(finalData.localizacao);


      const payload = {
        crimeType: finalData.tipo_de_crime,
        crimeWeight: crime_weight,
        latitude: locationInfo?.latitude || null,
        longitude: locationInfo?.longitude || null,
        location: locationInfo?.neighborhoodName||null,
        crimeData: finalData.data_crime
      }
    
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
      const locationMatch = result.match(/Localização:\s*(.*)/i);
    if (locationMatch && locationMatch[1].trim()){
          const locationLine = locationMatch[0]; 
          const locationName = String(locationMatch [1]).trim();
          const locationInfo = await geocodeAddress(locationName);  
          const enrichedLocation = `Região: ${locationInfo?.neighborhoodName}, ${locationInfo?.cityName}`;
          const enrichedResult = result.replace(locationLine, enrichedLocation);
          return NextResponse.json({result: enrichedResult,
    locationInfo, });
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
