"use client"

import { Button } from "@/components/ui/button"

interface AlertModelProps {
  onClose: () => void
  onConfirm: () => void
  onReject: () => void
}

export function AlertModel({ onClose, onConfirm, onReject }: AlertModelProps) {
  return (
    <div className="flex justify-center gap-3 mt-4 mb-4">
      <Button
        variant="outline"
        onClick={onConfirm}
        className="px-8 py-2 bg-[#4d4dff] hover:bg-[#3a3ad9] text-white"
      >
        Sim
      </Button>
      <Button onClick={onReject} className="px-8 py-2 bg-gray-200 text-gray-700 hover:bg-gray-300">
        Não
      </Button>
    </div>
  )
}

