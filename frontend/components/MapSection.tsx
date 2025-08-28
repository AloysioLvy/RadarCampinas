// components/MapSection.tsx
import CampinasMap from "@/components/campinas-map"

export default function MapSection() {
  return (
    <div className="bg-white rounded-lg shadow-md p-4 h-full">
      <h2 className="text-xl font-semibold mb-4">Mapa de Zonas de Risco - Criminalidade</h2>
      <div className="h-[500px] relative rounded-md overflow-hidden">
        <div className="absolute inset-0 bg-gray-200">
          <CampinasMap />
        </div>
      </div>
      <div className="mt-4 flex items-center">
        <div className="w-4 h-4 bg-[#FF0000] rounded-full mr-2"></div>
        <span className="text-sm text-gray-700">Zona de alto risco</span>
      </div>
    </div>
  )
}
