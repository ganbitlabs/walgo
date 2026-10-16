import { describe, it, expect, vi, beforeEach } from 'vitest'
import React from 'react'
import { createRoot } from 'react-dom/client'
import { act } from 'react-dom/test-utils'

const backend = vi.hoisted(() => ({
  ReadFile: vi.fn(),
  WriteFile: vi.fn(),
  DeleteFile: vi.fn(),
  CreateFile: vi.fn(),
  CreateDirectory: vi.fn(),
  ListFiles: vi.fn(),
  GetProject: vi.fn(),
  RenameFile: vi.fn(),
  MoveFile: vi.fn(),
  CopyFile: vi.fn(),
}))
vi.mock('../../wailsjs/go/main/App', () => backend)

import { useEditProject } from './useEditProject'

;(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true

type Hook = ReturnType<typeof useEditProject>

function renderEditor() {
  const ref: { current: Hook } = { current: null as unknown as Hook }
  const Probe = () => {
    ref.current = useEditProject()
    return null
  }
  act(() => {
    createRoot(document.createElement('div')).render(React.createElement(Probe))
  })
  return ref
}

const file = (path: string) => ({ name: path.split('/').pop()!, path, isDir: false })

function deferred<T>() {
  let resolve!: (v: T) => void
  const promise = new Promise<T>((r) => (resolve = r))
  return { promise, resolve }
}

beforeEach(() => {
  Object.values(backend).forEach((fn) => fn.mockReset())
  backend.WriteFile.mockResolvedValue({ success: true })
})

describe('useEditProject', () => {
  it('blocks saving when the read failed, so the file is never overwritten with an empty buffer', async () => {
    backend.ReadFile.mockResolvedValue({ success: false, content: '', error: 'access denied' })
    const hook = renderEditor()

    await act(() => hook.current.selectFile(file('/site/a.md')))

    expect(hook.current.fileReady).toBe(false)
    expect(hook.current.fileLoadError).toBe('access denied')
    let result: { success: boolean } | undefined
    await act(async () => {
      result = await hook.current.saveFile()
    })
    expect(result?.success).toBe(false)
    expect(backend.WriteFile).not.toHaveBeenCalled()
  })

  it('drops a stale read so A content is never saved into B', async () => {
    const readA = deferred<{ success: boolean; content: string }>()
    backend.ReadFile.mockImplementation((path: string) =>
      path === '/site/a.md' ? readA.promise : Promise.resolve({ success: true, content: 'B content' }),
    )
    const hook = renderEditor()

    let pendingA!: Promise<void>
    act(() => {
      pendingA = hook.current.selectFile(file('/site/a.md'))
    })
    await act(() => hook.current.selectFile(file('/site/b.md')))
    await act(async () => {
      readA.resolve({ success: true, content: 'A content' })
      await pendingA
    })

    expect(hook.current.selectedFile?.path).toBe('/site/b.md')
    expect(hook.current.fileContent).toBe('B content')
    await act(async () => {
      await hook.current.saveFile()
    })
    expect(backend.WriteFile).toHaveBeenCalledWith('/site/b.md', 'B content')
  })

  it('reports a failed write instead of claiming success', async () => {
    backend.ReadFile.mockResolvedValue({ success: true, content: 'hello' })
    backend.WriteFile.mockResolvedValue({ success: false, error: 'disk full' })
    const hook = renderEditor()

    await act(() => hook.current.selectFile(file('/site/a.md')))
    let result: { success: boolean; error?: string } | undefined
    await act(async () => {
      result = await hook.current.saveFile()
    })

    expect(result).toEqual({ success: false, error: 'disk full' })
  })

  it('reports a failed delete and keeps the open file', async () => {
    backend.ReadFile.mockResolvedValue({ success: true, content: 'hello' })
    backend.DeleteFile.mockResolvedValue({ success: false, error: 'access denied' })
    const hook = renderEditor()

    await act(() => hook.current.selectFile(file('/site/a.md')))
    let result: { success: boolean; error?: string } | undefined
    await act(async () => {
      result = await hook.current.deleteFile('/site/a.md')
    })

    expect(result).toEqual({ success: false, error: 'access denied' })
    expect(hook.current.selectedFile?.path).toBe('/site/a.md')
    expect(hook.current.fileContent).toBe('hello')
  })
})
