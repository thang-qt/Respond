"use client"

import { createContext, useContext, useEffect, useState } from "react"

type Theme = "light" | "dark" | "system"

interface ThemeContextValue {
  theme: Theme
  mounted: boolean
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: "system",
  mounted: false,
  toggleTheme: () => {},
})

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  // Always initialize to "system" to match server-rendered HTML + boot script.
  // The inline script in layout.tsx already prevents the visual flash.
  const [theme, setTheme] = useState<Theme>("system")
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    // Read the actual theme after mount to avoid hydration mismatch
    const stored = localStorage.getItem("theme")
    if (stored === "light" || stored === "dark" || stored === "system") {
      setTheme(stored)
    } else {
      setTheme("system")
    }
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!mounted) return

    const media = window.matchMedia("(prefers-color-scheme: dark)")
    const applyTheme = () => {
      const resolvedTheme = theme === "system" ? (media.matches ? "dark" : "light") : theme
      document.documentElement.classList.toggle("dark", resolvedTheme === "dark")
    }

    applyTheme()
    localStorage.setItem("theme", theme)

    if (theme !== "system") return

    const handleChange = () => applyTheme()
    media.addEventListener("change", handleChange)
    return () => media.removeEventListener("change", handleChange)
  }, [theme, mounted])

  const toggleTheme = () =>
    setTheme((prev) => {
      if (prev === "light") return "dark"
      if (prev === "dark") return "system"
      return "light"
    })

  return <ThemeContext.Provider value={{ theme, mounted, toggleTheme }}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  return useContext(ThemeContext)
}
