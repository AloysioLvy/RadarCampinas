"use client";

import { useEffect, useRef, useState } from "react";

interface Prediction {
  neighborhood: string;
  lat: number;
  long: number;
  risk_level: 0 | 1 | 2;
}

interface RiskDistribution {
  0: number;
  1: number;
  2: number;
}

interface Summary {
  message: string;
  target_year: number;
  target_month: number;
  rows_trained: number;
  rows_predicted: number;
  neighborhoods: number;
  risk_distribution: RiskDistribution;
}

interface MonthlyResponse {
  year: number;
  month: number;
  summary?: Summary;
  predictions: Prediction[];
  message?: string; // 🔥 para quando não há previsões
}

export default function CampinasMap() {
  const mapRef = useRef<HTMLDivElement>(null);
  const [mapLoaded, setMapLoaded] = useState(false);
  const mapInstance = useRef<any>(null);
  const leafletRef = useRef<any>(null);
  const layersGroup = useRef<any>(null);

  const [selectedDate, setSelectedDate] = useState<string>("2025-12");
  const [loading, setLoading] = useState(false);

  // ==========================
  // 📌 Carrega previsões
  // ==========================
  const loadPredictions = async (year: string, month: string) => {
    const cacheKey = `predictions-${year}-${month}`;

    // 1️⃣ TENTA CARREGAR DO CACHE
    const cached = localStorage.getItem(cacheKey);

    if (cached) {
      const parsed = JSON.parse(cached);
      drawRiskZones(parsed);
      return;
    }

    // 2️⃣ SE NÃO EXISTIR CACHE → FAZ GET
    const getUrl = `${process.env.NEXT_PUBLIC_MACHINE_LEARNING_ROUTE_URL}/predictions?year=${year}&month=${month}`;

    setLoading(true);

    try {
      const res = await fetch(getUrl, { method: "GET" });
      
      if (!res.ok) {
        throw new Error(`GET falhou com status ${res.status}`);
      }

      const data: MonthlyResponse = await res.json();

      // 🔥 Verifica se há previsões
      if (data.predictions && data.predictions.length > 0) {
        localStorage.setItem(cacheKey, JSON.stringify(data.predictions));
        drawRiskZones(data.predictions);
        setLoading(false);
        return;
      }

      // Se não há previsões, cai no fallback
      throw new Error("Nenhuma previsão encontrada");

    } catch (error) {
      console.warn("⚠️ GET falhou ou sem dados. Tentando treinar via POST...", error);

      // 3️⃣ FALLBACK: FAZ POST PARA TREINAR
      const trainUrl = `${process.env.NEXT_PUBLIC_MACHINE_LEARNING_ROUTE_URL}/training-monthly?year=${year}&month=${month}`;

      try {
        const trainRes = await fetch(trainUrl, { method: "POST" });

        if (!trainRes.ok) {
          throw new Error(`POST falhou com status ${trainRes.status}`);
        }

        const trainData = await trainRes.json();

        // 4️⃣ RETRY: Tenta GET novamente após treinar
        const retryRes = await fetch(getUrl, { method: "GET" });

        if (!retryRes.ok) {
          throw new Error(`Retry GET falhou com status ${retryRes.status}`);
        }

        const retryData: MonthlyResponse = await retryRes.json();

        if (retryData.predictions && retryData.predictions.length > 0) {
          localStorage.setItem(cacheKey, JSON.stringify(retryData.predictions));
          drawRiskZones(retryData.predictions);
        } else {
          console.error("❌ Nenhuma previsão retornada mesmo após treinar");
        }

      } catch (innerError) {
        console.error("❌ Falha ao treinar e recarregar previsões:", innerError);
      }
    }

    setLoading(false);
  };

  // ==========================
  // 🎨 Desenhar zonas
  // ==========================
  const drawRiskZones = (predictions: Prediction[]) => {
    if (!mapInstance.current || !leafletRef.current) {
      console.warn("⚠️ Mapa ainda não carregado, adiando drawRiskZones...");
      return;
    }

    const L = leafletRef.current;

    // 🔥 LIMPA CAMADAS ANTIGAS
    if (layersGroup.current) {
      layersGroup.current.clearLayers();
    } else {
      layersGroup.current = L.layerGroup().addTo(mapInstance.current);
    }

    // 🎨 DESENHA NOVAS CAMADAS
    predictions.forEach((p) => {
      const color =
        p.risk_level === 2 ? "red" : p.risk_level === 1 ? "yellow" : null;

      if (!color) return; // risco 0 não desenha nada

      L.circle([p.lat, p.long], {
        color,
        fillColor: color,
        fillOpacity: 0.4,
        radius: 500,
      })
        .bindPopup(`<b>${p.neighborhood}</b><br>Risco: ${p.risk_level}`)
        .addTo(layersGroup.current);
    });

  };

  // ==========================
  // 🔄 Data change = request
  // ==========================
  useEffect(() => {
    if (!selectedDate || !mapLoaded) return;

    const [year, month] = selectedDate.split("-");
    loadPredictions(year, month);
  }, [selectedDate, mapLoaded]);

  // ==========================
  // 🗺️ Inicialização do mapa
  // ==========================
  useEffect(() => {
    if (typeof window === "undefined" || !mapRef.current || mapLoaded) return;

    const loadMap = async () => {
      const L = await import("leaflet");
      await import("leaflet/dist/leaflet.css");

      leafletRef.current = L;

      delete L.Icon.Default.prototype._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl:
          "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png",
        iconUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png",
        shadowUrl:
          "https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png",
      });

      const campinasCoords: [number, number] = [-22.9064, -47.0616];

      mapInstance.current = L.map(mapRef.current).setView(campinasCoords, 12);

      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution:
          '&copy; <a href="https://www.openstreetmap.org">OpenStreetMap</a>',
      }).addTo(mapInstance.current);

      setMapLoaded(true);
    };

    loadMap();
  }, [mapLoaded]);

  return (
    <div className="h-full w-full relative">
      {/* Date Picker */}
      <div className="flex items-center mb-4 p-2 bg-white shadow rounded w-fit absolute top-2 left-2 z-[1000]">
        <label className="mr-2 font-medium">Selecionar Mês:</label>
        <input
          type="month"
          className="border p-2 rounded"
          value={selectedDate}
          onChange={(e) => setSelectedDate(e.target.value)}
        />
      </div>

      {/* MAPA */}
      <div ref={mapRef} className="h-full w-full" />

      {/* ============ LOADING ============ */}
      {loading && (
        <div className="absolute inset-0 bg-black bg-opacity-40 flex items-center justify-center z-[2000]">
          <div className="animate-spin w-14 h-14 border-4 border-white border-t-transparent rounded-full"></div>
        </div>
      )}
    </div>
  );
}