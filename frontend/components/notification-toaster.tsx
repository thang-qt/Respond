"use client"

import { Toaster } from "sonner"

export function NotificationToaster() {
  return (
    <Toaster
      position="top-right"
      toastOptions={{
        className: "font-sans",
        style: {
          background: "var(--bg-surface)",
          color: "var(--text-primary)",
          border: "1px solid var(--border-default)",
          boxShadow: "0px 10px 30px rgba(15,12,10,0.12)",
          fontSize: "13px",
        },
      }}
    />
  )
}
