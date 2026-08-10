import { useCallback, useEffect, useState } from 'react'

type Tema = 'light' | 'dark'

const STORAGE_KEY = 'rendix-theme'

function temaInicial(): Tema {
  const guardado = localStorage.getItem(STORAGE_KEY)
  if (guardado === 'light' || guardado === 'dark') return guardado
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function useTheme() {
  const [tema, setTema] = useState<Tema>(temaInicial)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', tema === 'dark')
    localStorage.setItem(STORAGE_KEY, tema)
  }, [tema])

  const toggleTema = useCallback(() => {
    setTema((t) => (t === 'dark' ? 'light' : 'dark'))
  }, [])

  return { tema, toggleTema }
}
