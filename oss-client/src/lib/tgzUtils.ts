import untar from 'js-untar'
import pako from 'pako'
import { getWindow } from './windowUtils'
import { extractTgzInWorker } from './tgzWorker'

// DirectoryNode type definition
export type DirectoryNode = {
  name: string
  isDirectory: boolean
  size: bigint
  children: DirectoryNode[]
  path: string
  content: Uint8Array
}

// Map file extensions to Monaco language ids
export const EXTENSION_TO_LANGUAGE: Record<string, string> = {
  py: 'python',
  js: 'javascript',
  ts: 'typescript',
  jsx: 'javascript',
  tsx: 'typescript',
  json: 'json',
  yaml: 'yaml',
  yml: 'yaml',
  md: 'markdown',
  sh: 'shell',
  bash: 'shell',
  zsh: 'shell',
  go: 'go',
  rs: 'rust',
  java: 'java',
  cpp: 'cpp',
  c: 'cpp',
  h: 'cpp',
  hpp: 'cpp',
  html: 'html',
  css: 'css',
  scss: 'scss',
  sql: 'sql',
  xml: 'xml',
  dockerfile: 'dockerfile',
}

type TarFile = {
  name: string
  type: string | number
  buffer: ArrayBuffer
}

// Custom error for file size exceeded
export class FileSizeExceededError extends Error {
  constructor(
    message: string,
    public readonly presignedUrl: string,
    public readonly fileSize: number,
    public readonly maxSize: number,
  ) {
    super(message)
    this.name = 'FileSizeExceededError'
  }
}

// Custom error for 403/CORS errors that includes presignedUrl
export class AccessDeniedError extends Error {
  constructor(
    message: string,
    public readonly presignedUrl: string,
    public readonly originalError: unknown,
  ) {
    super(message)
    this.name = 'AccessDeniedError'
  }
}

// Size threshold: 10MB (10 * 1024 * 1024 bytes)
const MAX_FILE_SIZE = 10 * 1024 * 1024

/**
 * Checks the file size via HEAD request
 */
async function checkFileSize(presignedUrl: string): Promise<number> {
  const response = await fetch(presignedUrl, {
    method: 'HEAD',
    credentials: 'include',
  })
  if (!response.ok) {
    throw new Error(
      `Failed to check file size: ${response.statusText} (${response.status})`,
    )
  }

  const contentLength = response.headers.get('content-length')
  if (!contentLength) {
    // If Content-Length is not available, we can't check the size
    // Return a large number to allow the download to proceed
    return 0
  }

  return parseInt(contentLength, 10)
}

/**
 * Downloads a tgz file from a presigned URL and extracts it client-side
 * @param presignedUrl The presigned URL to download the tgz file from
 * @param includeFileContents Whether to include file contents in the response
 * @param checkSize Whether to check file size before downloading (default: false)
 * @returns A promise that resolves to the root DirectoryNode
 */
export async function extractTgzFromUrl(
  presignedUrl: string,
  includeFileContents: boolean = true,
  checkSize: boolean = false,
  useWorker: boolean = false, // Off by default; main thread is typically fast enough
): Promise<DirectoryNode> {
  try {
    // Check file size before downloading (if enabled)
    if (checkSize) {
      const fileSize = await checkFileSize(presignedUrl)
      if (fileSize > 0 && fileSize > MAX_FILE_SIZE) {
        throw new FileSizeExceededError(
          `File size (${formatFileSize(fileSize)}) exceeds the maximum allowed size (${formatFileSize(MAX_FILE_SIZE)}). Please download directly using the link below.`,
          presignedUrl,
          fileSize,
          MAX_FILE_SIZE,
        )
      }
    }

    // Use web worker if available and requested
    // The main thread extraction is fast enough for most use cases
    if (useWorker && typeof Worker !== 'undefined') {
      try {
        // Download and extract in worker to keep main thread free
        return await extractTgzInWorker(presignedUrl, includeFileContents)
      } catch {
        // Fall back to main thread if worker setup or execution fails
      }
    }

    // Fallback to main thread extraction
    // Download the tgz file
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

    // Fallback to main thread extraction
    const gzippedData = new Uint8Array(arrayBuffer)

    // Decompress gzip
    let decompressedData: Uint8Array
    try {
      decompressedData = pako.ungzip(gzippedData)
    } catch (error) {
      throw new Error(
        `Failed to decompress gzip: ${error instanceof Error ? error.message : String(error)}`,
      )
    }

    if (decompressedData.length === 0) {
      throw new Error('Decompressed data is empty')
    }

    // Extract tar
    let files: TarFile[]
    try {
      // Convert Uint8Array to ArrayBuffer for untar
      // Create a new ArrayBuffer copy to ensure proper type
      const arrayBuffer = decompressedData.slice().buffer
      files = await untar(arrayBuffer)
    } catch (error) {
      throw new Error(
        `Failed to extract tar: ${error instanceof Error ? error.message : String(error)}`,
      )
    }

    if (!files || files.length === 0) {
      throw new Error('No files found in tar archive')
    }

    // Build directory tree
    const nodesByPath = new Map<string, DirectoryNode>()
    const root: DirectoryNode = {
      name: '',
      isDirectory: true,
      size: BigInt(0),
      children: [],
      path: '',
      content: new Uint8Array(),
    }
    nodesByPath.set('', root)

    // Process all files and create directory structure
    for (const file of files) {
      // Skip if file name is empty
      if (!file.name || file.name.trim() === '') {
        continue
      }

      const pathParts = file.name.split('/').filter((p: string) => p !== '')
      if (pathParts.length === 0) {
        continue
      }

      let currentPath = ''
      let parentNode = root

      // Check if this is a directory (tar type '5' is directory)
      const isDirectory = file.type === '5' || file.type === 5

      // Create directory nodes for each path segment
      for (let i = 0; i < pathParts.length; i++) {
        const part = pathParts[i]
        const isLast = i === pathParts.length - 1
        const newPath = currentPath ? `${currentPath}/${part}` : part

        if (!nodesByPath.has(newPath)) {
          const nodeIsDirectory = !isLast || isDirectory
          const node: DirectoryNode = {
            name: part,
            isDirectory: nodeIsDirectory,
            size:
              isLast && !isDirectory && file.buffer
                ? BigInt(file.buffer.byteLength)
                : BigInt(0),
            children: [],
            path: newPath,
            content: new Uint8Array(),
          }

          nodesByPath.set(newPath, node)
          parentNode.children.push(node)

          if (isLast && !isDirectory && file.buffer && includeFileContents) {
            // Include file content if requested
            node.content = new Uint8Array(file.buffer)
          }
        }

        parentNode = nodesByPath.get(newPath)!
        currentPath = newPath
      }
    }

    return root
  } catch (error) {
    console.error('Error extracting tgz file:', error)
    throw error instanceof Error ? error : new Error(String(error))
  }
}

export const getFileContent = (node: DirectoryNode): string | null => {
  const content = (node as { content?: Uint8Array | string }).content
  if (!content || (content instanceof Uint8Array && content.length === 0)) {
    return null
  }
  try {
    // Content is Uint8Array from protobuf
    let bytes: Uint8Array
    if (content instanceof Uint8Array) {
      bytes = content
    } else if (typeof content === 'string') {
      // If it's a base64 string (from JSON), decode it
      const binaryString = atob(content)
      bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
    } else {
      return null
    }
    // Try to decode as UTF-8 text
    const decoder = new TextDecoder('utf-8', { fatal: false })
    return decoder.decode(bytes)
  } catch {
    return null
  }
}

// Find the first file in the tree recursively
export const findFirstFile = (
  currentNode: DirectoryNode,
): DirectoryNode | null => {
  // If current node is a file, return it
  if (!currentNode.isDirectory) {
    return currentNode
  }

  // If it's a directory, search through children
  if (currentNode.children && currentNode.children.length > 0) {
    // Sort children: directories first, then files
    const sortedChildren = [...currentNode.children].sort((a, b) => {
      if (a.isDirectory === b.isDirectory) {
        return a.name.localeCompare(b.name)
      }
      return a.isDirectory ? -1 : 1
    })

    // Search through children
    for (const child of sortedChildren) {
      const found = findFirstFile(child)
      if (found) {
        return found
      }
    }
  }

  return null
}

/**
 * Finds a file by its path in the directory tree
 * @param node The root node to search from
 * @param filePath The path to search for (e.g., "examples/basics/types/optional_int_collection.py")
 * @returns The DirectoryNode if found, null otherwise
 */
export const findFileByPath = (
  node: DirectoryNode,
  filePath: string,
): DirectoryNode | null => {
  // Normalize the search path (remove leading/trailing slashes)
  const normalizedSearchPath = filePath.replace(/^\/+|\/+$/g, '')

  // Normalize paths for comparison (remove leading/trailing slashes)
  const normalizePath = (path: string) => path.replace(/^\/+|\/+$/g, '')

  // If this node is a file, check if it matches
  if (!node.isDirectory) {
    const nodePathNormalized = normalizePath(node.path)

    // Exact match
    if (nodePathNormalized === normalizedSearchPath) {
      return node
    }

    // Check if path ends with the search path (handles root directory with empty name)
    if (
      nodePathNormalized.endsWith(`/${normalizedSearchPath}`) ||
      node.path.endsWith(normalizedSearchPath)
    ) {
      return node
    }

    // Check if the filename matches (in case path structure differs)
    const searchFileName = normalizedSearchPath.split('/').pop()
    if (node.name === searchFileName) {
      return node
    }
  }

  // Search in children
  if (node.isDirectory && node.children) {
    for (const child of node.children) {
      const found = findFileByPath(child, normalizedSearchPath)
      if (found) {
        return found
      }
    }
  }

  return null
}

// Determine language from file extension
export const getLanguageFromPath = (path: string): string => {
  const ext = path.split('.').pop()?.toLowerCase()
  return (
    EXTENSION_TO_LANGUAGE[ext || ''] ||
    // TODO: is this best fallback?
    'plaintext'
  )
}

/**
 * Recursively collects all files from a directory tree
 */
function collectFiles(
  node: DirectoryNode,
  basePath: string = '',
): Array<{ path: string; content: Uint8Array }> {
  const files: Array<{ path: string; content: Uint8Array }> = []

  if (!node.isDirectory) {
    // It's a file, add it if it has content
    if (node.content && node.content.length > 0) {
      files.push({
        path: basePath ? `${basePath}/${node.name}` : node.name,
        content: node.content,
      })
    }
    return files
  }

  // It's a directory, recursively collect files from children
  const currentPath = basePath ? `${basePath}/${node.name}` : node.name
  for (const child of node.children || []) {
    files.push(...collectFiles(child, currentPath))
  }

  return files
}

/**
 * Recursively creates directory structure and files using File System Access API
 */
async function createDirectoryStructure(
  node: DirectoryNode,
  parentHandle: FileSystemDirectoryHandle,
): Promise<void> {
  if (!node.isDirectory) {
    // It's a file, create it
    if (node.content && node.content.length > 0) {
      const sanitizedName = sanitizeFileName(node.name)
      if (!sanitizedName) {
        // Skip files with empty names after sanitization
        return
      }
      const fileHandle = await parentHandle.getFileHandle(sanitizedName, {
        create: true,
      })
      const writable = await fileHandle.createWritable()
      // Convert Uint8Array to Blob for File System Access API compatibility
      // Create a new Uint8Array to ensure we have a regular ArrayBuffer
      const contentArray = new Uint8Array(node.content)
      const blob = new Blob([contentArray])
      await writable.write(blob)
      await writable.close()
    }
    return
  }

  // It's a directory, create it and process children
  if (node.name) {
    // Only create directory if it has a name (skip root)
    const sanitizedName = sanitizeFileName(node.name)
    if (!sanitizedName) {
      // Skip directories with empty names after sanitization, process children in parent
      for (const child of node.children || []) {
        await createDirectoryStructure(child, parentHandle)
      }
      return
    }
    const dirHandle = await parentHandle.getDirectoryHandle(sanitizedName, {
      create: true,
    })

    for (const child of node.children || []) {
      await createDirectoryStructure(child, dirHandle)
    }
  } else {
    // Root directory, process children directly
    for (const child of node.children || []) {
      await createDirectoryStructure(child, parentHandle)
    }
  }
}

/**
 * Formats file size in bytes to human-readable format
 */
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

/**
 * Sanitizes a folder/filename to remove invalid characters for the file system
 */
function sanitizeFileName(name: string): string {
  // Remove or replace invalid characters for macOS/Windows/Linux
  // Colons (:) are invalid on macOS, and other special characters can cause issues
  return (
    name
      .replace(/[:<>"|?*\x00-\x1f]/g, '_') // Replace invalid chars with underscore
      .replace(/\.\./g, '_') // Replace .. with _ to prevent path traversal
      .replace(/^\.+/, '') // Remove leading dots
      .replace(/\/$/, '') // Remove trailing slashes
      .trim() || 'folder'
  ) // Fallback to 'folder' if empty
}

/**
 * Downloads the entire directory tree as an unzipped folder structure
 * Uses File System Access API to let users choose where to save the folder
 * Requires a browser that supports the File System Access API
 */
export async function downloadFolderAsZip(
  rootNode: DirectoryNode,
  folderName: string = 'code',
): Promise<void> {
  // Sanitize the folder name to avoid filesystem errors
  const safeFolderName = sanitizeFileName(folderName)
  // Collect all files from the tree
  const files = collectFiles(rootNode)

  if (files.length === 0) {
    throw new Error('No files to download')
  }

  // Check if File System Access API is available
  const win = getWindow()
  if (!win || !('showDirectoryPicker' in win)) {
    throw new Error(
      'File System Access API is not supported in this browser. Please use a modern browser like Chrome or Edge.',
    )
  }

  const winWithPicker = win as unknown as {
    showDirectoryPicker?: (options?: {
      mode?: 'read' | 'readwrite'
    }) => Promise<FileSystemDirectoryHandle>
  }

  if (!winWithPicker.showDirectoryPicker) {
    throw new Error(
      'File System Access API is not available. Please use a modern browser like Chrome or Edge.',
    )
  }

  try {
    // Open directory picker to let user choose where to save
    const selectedDirHandle = await winWithPicker.showDirectoryPicker({
      mode: 'readwrite',
    })

    // Create a new subfolder within the selected directory to avoid system file conflicts
    const folderHandle = await selectedDirHandle.getDirectoryHandle(
      safeFolderName,
      { create: true },
    )

    // Create the folder structure inside the new subfolder
    await createDirectoryStructure(rootNode, folderHandle)
  } catch (error) {
    // User cancelled - don't throw an error
    if ((error as Error).name === 'AbortError') {
      return
    }
    // Re-throw other errors
    throw error
  }
}

export const getTgzUrlFromArgs = (
  args: string[] | undefined,
): string | undefined => {
  if (!args) return undefined
  const index = args.indexOf('--tgz')
  return index !== -1 && index + 1 < args.length ? args[index + 1] : undefined
}

/**
 * Check if --tgz flag exists in an array
 */
export const hasTgzInArray = (array: string[] | undefined): boolean => {
  if (!array || !Array.isArray(array)) return false
  const index = array.indexOf('--tgz')
  return index !== -1 && index + 1 < array.length
}

/**
 * Check if a container has a tgz code bundle
 * Checks both args and command arrays (apps use command, tasks use args)
 */
export const hasTgzInContainer = (
  container: { args?: string[]; command?: string[] } | undefined,
): boolean => {
  if (!container) return false
  return hasTgzInArray(container.args) || hasTgzInArray(container.command)
}
