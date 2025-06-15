import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/styles/globals.css";
import React from 'react'
import Header from '@/layouts/Header'
import Footer from '@/layouts/Footer'
import ScrollVisible from '@/components/ScrollVisible'
import GlobalProgressProvider from '@/components/client/GlobalProgressProvider'
import ChangeTitle from '@/components/client/ChangeTitle'

const inter = Inter({ subsets: ["latin"] });

// error.tsx 通用props
export interface ErrorProps {
  error: Error,
  reset: () => void
}

export const metadata: Metadata = {
  title: "Next",
  description: "一个分享前后端技术，Java、Golang、Vue和React技巧以及一些随笔的博客",
};

// 顶部导航列表
const navList = [
  { name: '博客', url: '/', icon: <i className="inline-block i-tabler:home" /> },
  { name: '动态', url: '/topic', icon: <i className="inline-block i-tabler:news" /> },
  { name: '分类', url: '/categorys', icon: <i className="inline-block i-tabler:category" /> },
  { name: '标签', url: '/tags', icon: <i className="inline-block i-tabler:tags" /> },
  { name: '相册', url: '#', icon: <i className="inline-block i-tabler:brand-google-photos" /> },
  { name: '关于', url: '/about', icon: <i className="inline-block i-tabler:user-bolt" /> },
  { name: '友情链接', url: '/links', icon: <i className="inline-block i-tabler:external-link" />  },
]

const RootLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <html lang="zh-cn">
      <body className={inter.className}>
        <ScrollVisible />
        <ChangeTitle />
        <GlobalProgressProvider>
          <>
            <Header navList={navList} />
            <main>
              {children}
            </main>
            <Footer />
          </>
        </GlobalProgressProvider>
      </body>
    </html>
  );
}

export default RootLayout;