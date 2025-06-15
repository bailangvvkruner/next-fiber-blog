'use client'

import '@/styles/components/tags-word-cloud.scss'
import React, { useEffect } from 'react'
import DynamicCard from '@/components/DynamicCard'
import TagCloud from 'TagCloud'
import { useRouter } from '@bprogress/next'

/**
 * 标签词云组件
 * @constructor
 */
export const WordCloud: React.FC<{
  children: React.ReactNode
}> = ({ children }): React.ReactNode => {

  const router = useRouter();

  useEffect( () => {
    const wordContainer = document.querySelector<HTMLDivElement>('.hidden-word-cloud-list');
    if (!wordContainer) {
      return ;
    }
    const wordList = wordContainer.querySelectorAll('.temp-word-cloud-li-item');
    const texts: string[] = []
    wordList.forEach(item => {
      const firstElement = item.firstElementChild as HTMLElement
      if (!firstElement) {
        return;
      }
      const spanElement = document.createElement('span');
      spanElement.innerText = firstElement.innerText;
      spanElement.classList.add('word-cloud-item')
      spanElement.style.color = firstElement.style.color;
      spanElement.setAttribute("path", firstElement.getAttribute("href") || '');
      texts.push(spanElement.outerHTML);
    })
    const wordCloud = TagCloud('.tags-word-cloud', texts, {
      useContainerInlineStyles: false,
      useHTML: true,
    });
    const cloudContainer = document.querySelector('.tags-word-cloud');
    if (cloudContainer) {
      const words = cloudContainer.getElementsByClassName('word-cloud-item');
      if (words && words.length > 0) {
        for (let i = 0; i < words.length; i++) {
          const element = words[i] as HTMLSpanElement;
          element.addEventListener('click', function() {
            const path = this.getAttribute("path");
            path && router.push(path, { scroll: true });
          })
        }
      }
    }
    return () => {
      wordCloud.destroy();
    }
  }, [children])
  return (
    <DynamicCard title="TAGS" icon="i-tabler:tags">
      <div className="tags-word-cloud"></div>
      <ul className="hidden-word-cloud-list hidden">
        { children }
      </ul>
    </DynamicCard>
  )
}

export const WordCloudItem: React.FC<{
  children: React.ReactNode
}> = ({ children }) => {
  return (
    <li className="temp-word-cloud-li-item">{ children }</li>
  )
}