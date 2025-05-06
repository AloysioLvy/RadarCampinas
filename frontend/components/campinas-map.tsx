"use client"

import { useEffect, useRef, useState } from "react"

export default function CampinasMap() {
  const mapRef = useRef<HTMLDivElement>(null)
  const [mapLoaded, setMapLoaded] = useState(false)
  

  useEffect(() => {
    if (typeof window === "undefined" || !mapRef.current || mapLoaded) return

    // importando dinamicamente a leaflet, biblioteca de mapas em javascript
    const loadMap = async () => {
      const L = await import("leaflet")
      await import("leaflet/dist/leaflet.css")

      //exclui alguns icones padroes indesejaveis do leaflet
      delete L.Icon.Default.prototype._getIconUrl
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png",
        iconUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png",
        shadowUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png",
      })

      // cordenadas de campinas
      const campinasCoords: [number, number] = [-22.9064, -47.0616]

      // inicializando
      const map = L.map(mapRef.current).setView(campinasCoords, 13)

  
      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      }).addTo(map)

      // adicionando zonas de perigo (dados de exemplo))
      const dangerZones = [
        { coords: [-22.9064, -47.0616], radius: 500, name: "Centro" },
        { coords: [-22.9364, -47.0916], radius: 300, name: "Jardim Chapadão" },
        { coords: [-22.8764, -47.0316], radius: 400, name: "Barão Geraldo" },
      ]

      dangerZones.forEach((zone) => {
        L.circle(zone.coords as [number, number], {
          color: "	#DC143C",
          fillColor: "	#DC143C",
          fillOpacity: 0.35,
          radius: zone.radius,
        })
          .addTo(map)
          .bindPopup(`<b>Zona de Risco:</b> ${zone.name}`)
      })

      setMapLoaded(true)
    }

    loadMap()
  }, [mapLoaded])
  return (
    <div className="h-full w-full">
      <div ref={mapRef} className="h-full w-full" />
    </div>
  )
}



