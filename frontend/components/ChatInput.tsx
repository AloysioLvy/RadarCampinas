// components/ChatInput.tsx
import { FormEventHandler } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Send } from "lucide-react"

interface ChatInputProps {
  input: string
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  onSubmit: FormEventHandler
  isLoading: boolean
  disabled?: boolean
  
}

export default function ChatInput({ input, onChange, onSubmit, isLoading, disabled }: ChatInputProps) {
  return (
    <form onSubmit={onSubmit} className="flex w-full">
      <Input
          placeholder={disabled ? "Muito obrigado pela colaboração 🕵️‍♂️" : "Digite sua mensagem..."}

        value={input}
        onChange={onChange}
        disabled={disabled}

        className="flex-1 mr-2"
      />
      <Button type="submit" disabled={isLoading || !input.trim()} className="bg-[#4d4dff] hover:bg-[#3a3ad9]">
        <Send className="h-4 w-4" />
      </Button>
    </form>
  )
}