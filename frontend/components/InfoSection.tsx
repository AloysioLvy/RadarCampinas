// components/InfoSection.tsx
import { MapPin, MessageSquare } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"

export default function InfoSection() {
  return (
    <div className="bg-white rounded-lg shadow-md p-4 h-full">
      <h2 className="text-xl font-semibold mb-4">Informações</h2>
      <p className="text-gray-700 mb-4">
        O Radar Campinas é uma plataforma que mostra as áreas de risco na cidade de Campinas, São Paulo. Os
        dados são baseados em relatórios de ocorrências e atualizados regularmente.
      </p>
      <div className="space-y-4">
        <div className="flex items-start">
          <MapPin className="text-[#4d4dff] mr-2 mt-1 flex-shrink-0" />
          <p className="text-sm text-gray-700">
            Visualize as zonas de risco no mapa para planejar rotas mais seguras pela cidade.
          </p>
        </div>
        <div className="flex items-start">
          <MessageSquare className="text-[#4d4dff] mr-2 mt-1 flex-shrink-0" />
          <p className="text-sm text-gray-700">
            Contribua reportando ocorrências através do nosso chatbot para manter o mapa atualizado.
          </p>
        </div>
      </div>
      <div className="mt-6">
        <Link href="/chatbot">
          <Button className="w-full bg-[#4d4dff] hover:bg-[#3a3ad9]">Reportar Ocorrência</Button>
        </Link>
      </div>
    </div>
  )
}
