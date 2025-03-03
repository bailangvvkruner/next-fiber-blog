import React from 'react'
import { articleArchives } from '@/lib/api'
import DynamicCard from '@/components/DynamicCard'

const Archives: React.FC = async () => {
  const result = await articleArchives();
  return (
    <DynamicCard title="ARCHIVES" icon="i-tabler:archive">
      <ul className="list-none text-sm flex flex-col gap-row-2.5">
        { result.data.map(item => (
          <li key={item.month} className="flex justify-between">
            <time dateTime={item.month}>{ item.month }</time>
            <span>{ item.total }</span>
          </li>
        )) }
      </ul>
    </DynamicCard>
  )
}

export default Archives;