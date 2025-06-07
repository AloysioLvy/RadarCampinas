"use client"

import type React from "react"

import { useState } from "react"
import Link from "next/link"
import Navbar from "@/components/navbar"
import Footer from "@/components/footer"
import ChatInput from "@/components/ChatInput"
import ChatMessages from "@/components/ChatMessages"
import InstructionsCard from "@/components/InstructionsCard"
import { ArrowLeft } from "lucide-react"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"


export default function ChatbotPage() {
  const [location, setLocation] = useState("")
  const [messages, setMessages] = useState<any[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [showAlert, setShowAlert] = useState(false)
  const [botMessage, setBotMessage] = useState("")
  


  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return

  

    const userMessage = { role: "user", content: input }
    const newMessages = [...messages, userMessage]


    setMessages(newMessages)
    setIsLoading(true)

    try {
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ messages: newMessages }),
      })
      const data = await res.json()

      const botMessage = { role: "assistant", content: data.result }
      const botMessageContent = botMessage.content

      if (botMessageContent.includes("Está correto")) {
        setBotMessage(botMessageContent)
        setShowAlert(true)
      }

      setMessages((prev) => [...prev, botMessage])
    } catch (err) {
      console.error("Erro:", err)
    } finally {
      setIsLoading(false)
      setInput("")
    }
  }

  const handleConfirm = async () => {
    // Add user confirmation message
    const confirmMessage = { role: "user", content: "Sim, confirmo que todas as informações estão corretas." }
    const confirmMessageContent = confirmMessage.content
    setMessages((prev) => [...prev, confirmMessage])
    setShowAlert(false)
    setIsLoading(true)

    const userMessage = { role: "user", content: input }
    const newMessages = [...messages, userMessage]
    const newMessagesContent = [...messages, confirmMessage]


    
    

    try {
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ messages: newMessagesContent }),
      })
      const data = await res.json()

      const botMessage = { role: "assistant", content: data.result }
      setMessages((prev) => [...prev, botMessage])
    } catch (err) {
      console.error("Erro:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleReject = async () => {
    // Add user rejection message
    const rejectMessage = { role: "user", content: "Não, preciso corrigir algumas informações." }
    setMessages((prev) => [...prev, rejectMessage])
    setShowAlert(false)
    setIsLoading(true)

    try {
      // Send rejection to API
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ denuncia: "Não, preciso corrigir algumas informações." }),
      })
      const data = await res.json()

      const botMessage = { role: "assistant", content: data.result }
      setMessages((prev) => [...prev, botMessage])
    } catch (err) {
      console.error("Erro:", err)
    } finally {
      setIsLoading(false)
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
                <div className="flex">{/* Location field if you want to use it later */}</div>
              </div>

              <div className="border rounded-md p-4 h-[400px] overflow-y-auto mb-4 bg-gray-50">
                <ChatMessages
                  messages={messages}
                  showAlert={showAlert}
                  onConfirm={handleConfirm}
                  onReject={handleReject}
                  onCloseAlert={() => setShowAlert(false)}
                />
              </div>
            </CardContent>
            <CardFooter>
              <ChatInput input={input} onChange={handleInputChange} onSubmit={handleSubmit} isLoading={isLoading} />
            </CardFooter>
          </Card>

          <InstructionsCard />
        </div>
      </main>

      <Footer />
    </div>
  )
}
