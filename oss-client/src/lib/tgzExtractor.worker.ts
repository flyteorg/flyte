/// <reference lib="webworker" />
// Dedicated worker for extracting TGZ archives. Bundled with the app to avoid
// shipping a plaintext inline worker or fetching dependencies from CDNs.
import untar from 'js-untar'
import pako from 'pako'

type ExtractTgzMessage = {
  id: string
  type: 'EXTRACT_TGZ'
  payload: {
    presignedUrl: string
    includeFileContents: boolean
  }
}

type SerializedNode = {
  name: string
  isDirectory: boolean
  size: number
  children: SerializedNode[]
  path: string
  content: number[] | null
}

const buildTree = (
  files: Array<{ name: string; type: string | number; buffer?: ArrayBuffer }>,
  includeFileContents: boolean,
) => {
  const nodesByPath = new Map<string, SerializedNode>()
  const root: SerializedNode = {
    name: '',
    isDirectory: true,
    size: 0,
    children: [],
    path: '',
    content: null,
  }
  nodesByPath.set('', root)

  for (const file of files) {
    if (!file.name || file.name.trim() === '') continue

    const pathParts = file.name.split('/').filter((p) => p !== '')
    if (pathParts.length === 0) continue

    let currentPath = ''
    let parentNode = root
    const isDirectory = file.type === '5' || file.type === 5

    for (let i = 0; i < pathParts.length; i++) {
      const part = pathParts[i]
      const isLast = i === pathParts.length - 1
      const newPath = currentPath ? `${currentPath}/${part}` : part

      if (!nodesByPath.has(newPath)) {
        const nodeIsDirectory = !isLast || isDirectory
        const node: SerializedNode = {
          name: part,
          isDirectory: nodeIsDirectory,
          size:
            isLast && !isDirectory && file.buffer
              ? Number(file.buffer.byteLength)
              : 0,
          children: [],
          path: newPath,
          content:
            isLast && !isDirectory && file.buffer && includeFileContents
              ? Array.from(new Uint8Array(file.buffer))
              : null,
        }

        nodesByPath.set(newPath, node)
        parentNode.children.push(node)
      }

      parentNode = nodesByPath.get(newPath) as SerializedNode
      currentPath = newPath
    }
  }

  const serializeNode = (node: SerializedNode): SerializedNode => ({
    name: node.name,
    isDirectory: node.isDirectory,
    size: Number(node.size),
    children: node.children.map(serializeNode),
    path: node.path,
    content: node.content,
  })

  return serializeNode(root)
}

self.addEventListener(
  'message',
  async (event: MessageEvent<ExtractTgzMessage>) => {
    const { id, type, payload } = event.data || {}
    if (type !== 'EXTRACT_TGZ') return

    try {
      const { presignedUrl, includeFileContents } = payload

      const response = await fetch(presignedUrl, { credentials: 'include' })
      if (!response.ok) {
        throw new Error(
          `Failed to download tgz file: ${response.statusText} (${response.status})`,
        )
      }

      const arrayBuffer = await response.arrayBuffer()
      if (arrayBuffer.byteLength === 0) {
        throw new Error('Downloaded file is empty')
      }

      const gzippedData = new Uint8Array(arrayBuffer)
      const decompressedData = pako.ungzip(gzippedData)

      if (!decompressedData || decompressedData.length === 0) {
        throw new Error('Decompressed data is empty')
      }

      const files = await untar(decompressedData.slice().buffer)
      if (!files || files.length === 0) {
        throw new Error('No files found in tar archive')
      }

      const serializedRoot = buildTree(
        files as Array<{
          name: string
          type: string | number
          buffer?: ArrayBuffer
        }>,
        includeFileContents,
      )

      self.postMessage({
        id,
        type: 'EXTRACT_TGZ_SUCCESS',
        payload: { root: serializedRoot },
      })
    } catch (error) {
      self.postMessage({
        id,
        type: 'EXTRACT_TGZ_ERROR',
        payload: {
          error: {
            message: error instanceof Error ? error.message : String(error),
            name: error instanceof Error ? error.name : 'Error',
          },
        },
      })
    }
  },
)

export {}
