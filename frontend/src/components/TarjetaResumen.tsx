interface Props {
  label: string
  valor: string
  sub: string
  color: 'green' | 'red' | 'neutral'
}

const colorMap = {
  green: 'text-green-600 dark:text-green-400',
  red: 'text-red-600 dark:text-red-400',
  neutral: 'text-gray-900 dark:text-white',
}

export default function TarjetaResumen({ label, valor, sub, color }: Props) {
  return (
    <div className="bg-white border border-gray-200 dark:bg-gray-900 dark:border-gray-800 rounded-xl p-4">
      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1 uppercase tracking-wide">{label}</p>
      <p className={`text-lg font-bold font-mono ${colorMap[color]}`}>{valor}</p>
      {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{sub}</p>}
    </div>
  )
}
