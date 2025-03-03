'use client'

import '@/styles/components/comment.scss'
import 'highlight.js/styles/atom-one-dark.min.css'
import 'github-markdown-css/github-markdown-dark.css'
import '@/styles/components/markdown.scss'
import React, { FormEvent, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import Button from '@/components/Button'
import type { Comment, Page } from '@/lib/types'
import { clientAuthTokenKey, pageComment, saveComment, totalComment, userInfo } from '@/lib/client-api'
import CommentList from '@/components/comment/CommentList'
import { LevelContext } from '@/components/comment/context/LevelContext'
import useMessage from '@/components/message'
import { CommentState, StateContext, StateContextProps } from '@/components/comment/context/StateContext'
import Image from 'next/image'
import { getCommentRender } from '@/tools/markdown'
import MarkdownIt from 'markdown-it'

export interface CommentProps {
  type: number;
  articleId?: number;
  topicId?: number;
}

const Comment: React.FC<CommentProps> = ({ type, articleId, topicId }) => {
  const [ state, setState ] = useState<CommentState>({
    type: type,
    articleId: articleId,
    topicId: topicId
  })

  const getUserInfo = useCallback( async () => {
    const token = localStorage.getItem(clientAuthTokenKey)
    if (!token) {
      return;
    }
    const result = await userInfo();
    if (result.code === 200) {
      setState(prevState => {
        return {
          ...prevState,
          user: result.data
        }
      })
    } else if (result.code === 401) {
      // 401说明登录过期，从本地缓存中删除token
      localStorage.removeItem(clientAuthTokenKey)
    }
  }, [setState])

  const logout = useCallback(async () => {
    localStorage.removeItem(clientAuthTokenKey)
    setState(prevState => {
      return {
        ...prevState,
        user: undefined
      }
    })
  }, [])

  useEffect(() => {
    getUserInfo();
  }, [getUserInfo])

  return (
    <StateContext.Provider value={{ state, setState }}>
      <div className="comment flex flex-col gap-row-4">
        <div className="comment-header flex flex-col gap-row-4">
          {state.user && (
            <div className="flex justify-between items-center">
              <div className="comment-header-user flex justify-between gap-col-4">
                <Image src={state.user.avatar}
                       alt="avatar"
                       objectFit="cover"
                       height={48}
                       width={48}
                       className="user-avatar"
                />
                <div className="flex flex-col justify-between py-1">
                  <div className="flex items-center gap-col-2">
                    <a className="username" href={state.user.link} title={state.user.link} target="_blank">
                      {state.user.nickname}
                    </a>
                    <span className="user-level px-1">
                      {`Lv${state.user.level}`}
                    </span>
                    {state.user.labels && (
                      <ul className="list-none flex gap-col-2">
                        {state.user.labels.map(label => (
                          <li className="px-1" key={label}>{label}</li>
                        ))}
                      </ul>)}
                  </div>
                  <summary className="summary desc-text line-clamp-1 text-xs">{state.user.summary || '还没有简介'}</summary>
                </div>
              </div>
              <span className="desc-text text-xs underline cursor-pointer" onClick={logout}>退出登录</span>
            </div>
          )}
          <CommentEditor fid={0} rid={0} />
        </div>
        <CommentBody type={type} articleId={articleId} topicId={topicId} />
      </div>
    </StateContext.Provider>
  )
}

const CommentTotal: React.FC<CommentProps> = React.memo(({ type, articleId, topicId }) => {
  const [total, setTotal] = useState<number>(0)
  useEffect(() => {
    totalComment({
      page: 1,
      size: 10,
      commentType: type,
      articleId: articleId,
      topicId: topicId
    }).then(res => {
      if (res.code === 200) {
        setTotal(res.data);
      }
    })
  }, [type, articleId, topicId])
  return (
    <h2 className="text-sm info-text"><span className="main-text text-6">{total}</span> 条评论</h2>
  )
})
CommentTotal.displayName = 'comment-total';

export const CommentEditor: React.FC<{
  fid: number;
  rid: number;
  parentNickname?: string;
  onClose?: () => void;
}> = ({ fid, rid, parentNickname, onClose }) => {
  const [ inputValue, setInputValue ] = useState<string>('');
  const [ buttonLoading, setButtonLoading ] = useState<boolean>(false);
  const [ previewContent, setPreviewContent ] = useState<string>('');
  const [ isPreview, setIsPreview ] = useState<boolean>(false);

  const { state } = useContext<StateContextProps>(StateContext);
  const { showLoading, showDanger, showSuccess } = useMessage();
  const commentRender = useMemo<MarkdownIt>(() => getCommentRender(), [])
  
  const onSubmit = async (e: FormEvent) => {
    // 阻止默认提交
    e.preventDefault();
    setButtonLoading(true);
    const loadingToast = showLoading('评论提交中');
    try {
      const result = await saveComment({
        commentType: state.type,
        articleId: state.articleId,
        topicId: state.topicId,
        content: inputValue,
        fid: fid,
        rid: rid,
      })
      if (result.code === 200) {
        showSuccess('评论成功');
        setInputValue('');
        isPreview && setIsPreview(false);
      } else {
        showDanger(result.message);
      }
    } finally {
      loadingToast.close();
      setButtonLoading(false);
    }
  }

  useEffect(() => {
    if (!isPreview) {
      return;
    }
    setPreviewContent(commentRender.render(inputValue))
  }, [commentRender, inputValue, isPreview])
  
  return (
    <form className="comment-editor flex flex-col gap-row-4" onSubmit={onSubmit}>
      <textarea className="text-sm"
                value={inputValue}
                placeholder={state.user ? '留下点什么吧...' : '登录后才可以评论...'}
                rows={5}
                disabled={!state.user}
                required
                onChange={e => {
                  setInputValue(e.target.value)
                }}
      ></textarea>
      {isPreview && <div className="markdown-body" dangerouslySetInnerHTML={{ __html: previewContent }} />}
      <div className="flex justify-between">
        <div className="editor-option flex gap-col-2 items-center">
          { parentNickname && (
            <div className="parent-nickname-container flex items-center">
              <div className="parent-nickname text-xs py-1.5 px-2">
                {`回复 @${parentNickname}`}
              </div>
              <button className="px-1.5" onClick={onClose ? () => onClose() : undefined} type="button">
                <i className="inline-block i-tabler:x" />
              </button>
            </div>

          )}
          <button className="px-1.5" type="button" onClick={() => {
            setIsPreview(prevState => !prevState);
          }}>
            <i className="inline-block i-tabler:markdown" />
          </button>
        </div>
        {state.user ? (
          <Button text="发送" icon="i-tabler:send"
                  type="submit"
                  loading={buttonLoading}
          />
        ) : (
          <Button text="Github登录" icon="i-tabler:brand-github"
                  type="button"
                  onClick={() => {
                    const base64Path = btoa(window.location.pathname)
                    window.location.href = `https://github.com/login/oauth/authorize?client_id=${process.env.NEXT_PUBLIC_GITHUB_OAUTH_CLIENT_ID}&state=${base64Path}&scope=user:email`;
                  }}
          />
        )}
      </div>
    </form>
  )
}

const CommentBody: React.FC<CommentProps> = React.memo(({ type, articleId, topicId }) => {
  const [page, setPage] = useState<Page<Comment>>()
  const queryCommentPage = useCallback(async (page: number) => {
    const result = await pageComment({ page: page, size: 10, commentType: type, articleId, topicId })
    if (result.code === 200) {
      setPage(result.data);
    }
  }, [type, articleId, topicId])

  useEffect(() => {
    queryCommentPage(1);
    return () => {
      setPage(undefined);
    }
  }, [queryCommentPage])
  return (
    <>
      <CommentTotal type={type} articleId={articleId} topicId={topicId} />
      {page && (
        <LevelContext.Provider value={1}>
          <CommentList page={page} type={type} articleId={articleId} topicId={topicId} />
        </LevelContext.Provider>
      )}
    </>
  )
})
CommentBody.displayName = 'comment-body';

export default Comment;