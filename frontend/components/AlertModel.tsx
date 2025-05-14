"use client"

import { X, AlertTriangle } from 'lucide-react'
import { cn } from "@/lib/utils"

export function AlertModel({ onClose, className }: { onClose: () => void, className?: string }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop with blur effect */}
      <div className="absolute inset-0 bg-black/30 backdrop-blur-sm" onClick={onClose} />

      {/* Alert container */}
      <div className={cn("relative bg-gray-200 rounded-lg shadow-lg max-w-md w-full mx-4 p-6", className)}>
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute top-2 right-2 rounded-full bg-[#4d4dff] text-white p-1 hover:bg-[#3a3ad9] transition-colors"
          aria-label="Fechar"
        >
          <X size={20} />
        </button>

        {/* Alert content */}
        <div className="flex flex-col items-center text-center">
          <div className="bg-white p-3 rounded-md mb-4">
            <AlertTriangle size={40} className="text-[#4d4dff]" />
          </div>

          <p className="text-gray-800 text-sm">
            Por favor, revise os dados que o assistente identificou. Sua confirmação ajuda a manter a precisão do sistema.
          </p>
        </div>
      </div>
    </div>
  )
}
