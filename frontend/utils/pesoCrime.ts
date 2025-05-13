const crimesHediondos = [
    "tortura",
    "tráfico de drogas",
    "terrorismo",
    "homicídio em atividade típica de grupo de extermínio",
    "homicídio qualificado",
    "latrocínio",
    "extorsão qualificada pela morte",
    "extorsão mediante sequestro (inclusive na forma qualificada)",
    "estupro",
    "atentado violento ao pudor",
    "epidemia com resultado morte",
    "genocídio",
    "falsificação de produto terapêutico ou medicinal",
    "corrupção de produto terapêutico ou medicinal",
    "alteração de produto terapêutico ou medicinal"
  ];
  
export function pesoCrime(lista: any, tipoCrime: string) {
    var cont= 0;
    while(true){
        if(crimesHediondos[cont] == tipoCrime){
            return 9
        }
        cont ++
    }
  }
  