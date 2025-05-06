// components/InstructionsCard.tsx
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card"

export default function InstructionsCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Como funciona?</CardTitle>
      </CardHeader>
      <CardContent>
        <ol className="list-decimal pl-5 space-y-2">
          <li>Informe a localização da ocorrência no campo acima</li>
          <li>Descreva o que aconteceu, incluindo data e horário aproximados</li>
          <li>Forneça detalhes adicionais solicitados pelo assistente</li>
          <li>Sua ocorrência será analisada e adicionada ao mapa</li>
        </ol>
      </CardContent>
    </Card>
  )
}
