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
  
  const correctionMessages  = [
  "Tudo bem, vamos corrigir as informações.",
  "Claro, vamos ajustar isso juntos.",
  "Sem problemas, me diga o que precisa ser alterado.",
  "Entendi! Vamos lá, estou ouvindo.",
  "Desculpe se houve algum erro. Pode me dizer o que está incorreto?",
  "Certo! Vamos corrigir os dados.",
  "Ok, obrigado por avisar. Vamos fazer as correções.",
  "Combinado! Me informe o que precisa ser atualizado.",
  "Tranquilo! Vamos ajustar as informações agora.",
  "Vamos lá! O que está incorreto?",
  "Fico à disposição para corrigir. Pode continuar.",
  "Tudo certo. Diga o que precisa ser modificado.",
];
const thankYouMessages = [
  "Muito obrigado pela colaboração 😊",
  "Agradeço muito pela sua ajuda!",
  "Obrigado por confiar em nós 🙏",
  "Sua colaboração é muito importante 💙",
  "Agradecemos pelo envio das informações!",
  "Obrigado! Sua denúncia ajuda a tornar o lugar mais seguro.",
  "Gratidão pela sua contribuição!",
  "Obrigado por fazer a sua parte 💪",
  "Valeu pela ajuda! Seguimos juntos.",
  "Obrigado por dedicar seu tempo para relatar isso 🙌",
  "Sua colaboração foi registrada com sucesso. Muito obrigado!",
  "Obrigado pela confiança. Estamos aqui para ajudar!"
];



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
    
      if (botMessageContent.includes("Tipo de crime:")&& (botMessageContent.includes("tipo de crime")||
        botMessageContent.includes("data da denúncia")|| botMessageContent.includes("localização"))) {
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
    let messageIndex =  Math.floor(Math.random() * thankYouMessages.length);
    const confirmMessage = { role: "user", content: "Sim, confirmo que todas as informações estão corretas." }
    const chatMessage = { role: "assistant", content: thankYouMessages[messageIndex]}

    setMessages((prev) => [...prev, confirmMessage])
    setShowAlert(false)
    setIsLoading(true)

    const confirmMessagesContent = [...messages, confirmMessage]
    try {
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ messages: confirmMessagesContent }),
      })
      const data = await res.json()

      const botMessage = chatMessage;
      setMessages((prev) => [...prev, botMessage])
    } catch (err) {
      console.error("Erro:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleReject = async () => {
    // Add user rejection message
    let messageIndex = Math.floor(Math.random() * correctionMessages.length);
    const rejectMessage = { role: "user", content: "Não, preciso corrigir algumas informações." }
    const chatMessage = { role: "assistant", content: correctionMessages[messageIndex]}

    setMessages((prev) => [...prev, rejectMessage])
    setShowAlert(false)
    setIsLoading(true)
    const rejectMessageContent = [...messages, rejectMessage]

    try {
      // Send rejection to API
      
      const res = await fetch("/api", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ denuncia: rejectMessageContent}),
      })
      const data = await res.json()

      const botMessage = chatMessage
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
