// components/ChatMessages.tsx
'use client'

import { AlertTriangle } from "lucide-react"
import { Message } from "ai"

interface ChatMessagesProps {
  messages: Message[]
}

export default function ChatMessages({ messages }: ChatMessagesProps) {
  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center text-gray-500">
        <AlertTriangle className="h-12 w-12 mb-4 text-[#4d4dff]" />
        <p className="mb-2">Descreva a ocorrência que você presenciou ou tem conhecimento</p>
        <p className="text-sm">Nosso assistente irá coletar as informações necessárias</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {messages.map((message) => (
        <div key={message.id} className={`flex ${message.role === "user" ? "justify-end" : "justify-start"}`}>
          <div
            className={`max-w-[80%] rounded-lg px-4 py-2 ${
              message.role === "user" ? "bg-[#4d4dff] text-white" : "bg-gray-200 text-gray-800"
            }`}
          >
            {message.content.split('\n').map((line, index) => (
              <p key={index}>{line}</p>
            ))}
            
          </div>
        </div>
      ))}
    </div>
  )
}