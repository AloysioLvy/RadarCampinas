'use client'

import { useState } from "react"
import Link from "next/link"
import { ArrowLeft } from "lucide-react"
import { useChat } from "ai/react"
import Navbar from "@/components/navbar"
import Footer from "@/components/footer"
import ChatInput from "@/components/ChatInput"
import ChatMessages from "@/components/ChatMessages"
import InstructionsCard from "@/components/InstructionsCard"
import { CardFooter } from "@/components/ui/card"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

export default function ChatbotPage() {
  const [location, setLocation] = useState("")
  const [messages, setMessages] = useState<any[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return

    const userMessage = { role: "user", content: input }
    setMessages((prev) => [...prev, userMessage])
    setIsLoading(true)

    try {
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ denuncia: input }),
      })
      const data = await res.json()

      const botMessage = { role: "assistant", content: data.resultado }
      setMessages((prev) => [...prev, botMessage])
    } catch (err) {
      console.error("Erro:", err)
    } finally {
      setIsLoading(false)
      setInput("")
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Navbar />

      <main className="flex-1 container mx-auto p-4">
        <div className="max-w-3xl mx-auto">
          <div className="mb-4">
            <Link href="/" className="text-[#4d4dff] hover:text-[#3a3ad9] flex items-center">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Voltar para o mapa
            </Link>
          </div>

          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Reportar Ocorrência</CardTitle>
              <CardDescription>
                Converse com nosso assistente para reportar ocorrências em Campinas. Suas informações ajudam a manter o
                mapa atualizado.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mb-4">
                <label htmlFor="location" className="block text-sm font-medium text-gray-700 mb-1">
                  Localização da Ocorrência
                </label>
                <div className="flex">
                  {/* Campo de localização se quiser usar depois */}
                </div>
              </div>

              <div className="border rounded-md p-4 h-[400px] overflow-y-auto mb-4 bg-gray-50">
                <ChatMessages messages={messages} />
              </div>
            </CardContent>
            <CardFooter>
              <ChatInput
                input={input}
                onChange={handleInputChange}
                onSubmit={handleSubmit}
                isLoading={isLoading}
              />
            </CardFooter>
          </Card>

          <InstructionsCard />
        </div>
      </main>

      <Footer />
    </div>
  )
}
