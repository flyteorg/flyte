// Web Worker manager for TGZ extraction
// This runs off the main thread to avoid blocking the UI

import type { DirectoryNode } from './tgzUtils'

let worker: Worker | null = null
let workerReady = false
const pendingRequests = new Map<
  string,
  {
    resolve: (value: DirectoryNode) => void
    reject: (error: Error) => void
  }
>()

let requestIdCounter = 0

function createWorker(): Worker {
  // Use a module worker that is bundled with the app (no plaintext inline script)
  return new Worker(new URL('./tgzExtractor.worker.ts', import.meta.url), {
    type: 'module',
  })
}

async function getWorker(): Promise<Worker> {
  if (worker && workerReady) {
    return worker
  }

  return new Promise((resolve, reject) => {
    try {
      worker = createWorker()

      worker.onmessage = (event) => {
        const { id, type, payload } = event.data

        if (type === 'EXTRACT_TGZ_SUCCESS') {
          const request = pendingRequests.get(id)
          if (request) {
            pendingRequests.delete(id)
            // Convert Array back to Uint8Array for content
            const deserializeNode = (node: {
              name: string
              isDirectory: boolean
              size: number
              children: Array<unknown>
              path: string
              content: number[] | null
            }): DirectoryNode => {
              return {
                name: node.name,
                isDirectory: node.isDirectory,
                size: BigInt(node.size),
                children: node.children.map((child) =>
                  deserializeNode(child as typeof node),
                ),
                path: node.path,
                content:
                  node.content && Array.isArray(node.content)
                    ? new Uint8Array(node.content)
                    : new Uint8Array(),
              }
            }
            const root = deserializeNode(payload.root)
            request.resolve(root)
          }
        } else if (type === 'EXTRACT_TGZ_ERROR') {
          const request = pendingRequests.get(id)
          if (request) {
            pendingRequests.delete(id)
            const error = new Error(payload.error.message)
            error.name = payload.error.name
            request.reject(error)
          }
        }
      }

      worker.onerror = (error) => {
        console.error('TGZ Worker error:', error)
        for (const [id, request] of pendingRequests.entries()) {
          pendingRequests.delete(id)
          request.reject(
            new Error(`Worker error: ${error.message || 'Unknown error'}`),
          )
        }
        worker = null
        workerReady = false
        reject(error)
      }

      workerReady = true
      resolve(worker)
    } catch (error) {
      reject(error)
    }
  })
}

export async function extractTgzInWorker(
  presignedUrl: string,
  includeFileContents: boolean = true,
): Promise<DirectoryNode> {
  const worker = await getWorker()
  const id = `req_${++requestIdCounter}_${Date.now()}`

  return new Promise((resolve, reject) => {
    pendingRequests.set(id, { resolve, reject })

    // Send presigned URL to worker - it will download and extract
    worker.postMessage({
      id,
      type: 'EXTRACT_TGZ',
      payload: {
        presignedUrl,
        includeFileContents,
      },
    })

    // Timeout after 5 minutes
    setTimeout(
      () => {
        if (pendingRequests.has(id)) {
          pendingRequests.delete(id)
          reject(new Error('Extraction timeout'))
        }
      },
      5 * 60 * 1000,
    )
  })
}
