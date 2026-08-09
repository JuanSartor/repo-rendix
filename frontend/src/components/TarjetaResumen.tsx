interface Props {
  label: string
  valor: string
  sub: string
  color: 'green' | 'red' | 'neutral'
}

const colorMap = {
  green: 'text-green-400',
  red: 'text-red-400',
  neutral: 'text-white',
}

export default function TarjetaResumen({ label, valor, sub, color }: Props) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-4">
      <p className="text-xs text-gray-400 mb-1 uppercase tracking-wide">{label}</p>
      <p className={`text-lg font-bold font-mono ${colorMap[color]}`}>{valor}</p>
      {sub && <p className="text-xs text-gray-500 mt-1">{sub}</p>}
    </div>
  )
}
