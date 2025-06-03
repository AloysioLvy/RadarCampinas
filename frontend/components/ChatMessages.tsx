"use client"

import { AlertTriangle } from "lucide-react"
import { AlertModel } from "./AlertModel"

interface ChatMessagesProps {
  messages: any[]
  showAlert: boolean
  onConfirm: () => void
  onReject: () => void
  onCloseAlert: () => void
}

export default function ChatMessages({ messages, showAlert, onConfirm, onReject, onCloseAlert }: ChatMessagesProps) {
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
      {messages.map((message, index) => (
        <div key={index}>
          <div className={`flex ${message.role === "user" ? "justify-end" : "justify-start"}`}>
            <div
              className={`max-w-[80%] rounded-lg px-4 py-2 ${
                message.role === "user" ? "bg-[#4d4dff] text-white" : "bg-gray-200 text-gray-800"
              }`}
            >
              <div className="whitespace-pre-wrap">{message.content}</div>
            </div>
          </div>

          {/* Show confirmation buttons after the last assistant message that contains summary */}
          {message.role === "assistant" && index === messages.length - 1 && showAlert && (
            <div className="flex justify-start mt-2">
              <div className="max-w-[80%]">
                <AlertModel onClose={onCloseAlert} onConfirm={onConfirm} onReject={onReject} />
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
