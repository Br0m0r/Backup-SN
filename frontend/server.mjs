import { createReadStream } from 'node:fs'
import { stat } from 'node:fs/promises'
import { createServer } from 'node:http'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const contentTypes = new Map([
  ['.css', 'text/css; charset=utf-8'],
  ['.gif', 'image/gif'],
  ['.html', 'text/html; charset=utf-8'],
  ['.ico', 'image/x-icon'],
  ['.jpeg', 'image/jpeg'],
  ['.jpg', 'image/jpeg'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.png', 'image/png'],
  ['.svg', 'image/svg+xml'],
  ['.webp', 'image/webp'],
  ['.woff', 'font/woff'],
  ['.woff2', 'font/woff2'],
])

export function createHandler(rootDirectory) {
  const root = path.resolve(rootDirectory)
  const indexPath = path.join(root, 'index.html')

  return async function handler(request, response) {
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      response.writeHead(405, { Allow: 'GET, HEAD' })
      response.end()
      return
    }

    let pathname
    try {
      pathname = decodeURIComponent(new URL(request.url, 'http://localhost').pathname)
    } catch {
      response.writeHead(400)
      response.end('Bad request')
      return
    }

    const candidate = path.resolve(root, `.${pathname}`)
    if (candidate !== root && !candidate.startsWith(`${root}${path.sep}`)) {
      response.writeHead(400)
      response.end('Bad request')
      return
    }

    let filePath = candidate
    try {
      const fileInfo = await stat(filePath)
      if (!fileInfo.isFile()) filePath = indexPath
    } catch {
      filePath = indexPath
    }

    const extension = path.extname(filePath).toLowerCase()
    const headers = {
      'Content-Type': contentTypes.get(extension) || 'application/octet-stream',
      'Cache-Control': pathname.startsWith('/assets/')
        ? 'public, max-age=31536000, immutable'
        : 'no-cache',
    }

    try {
      const fileInfo = await stat(filePath)
      headers['Content-Length'] = fileInfo.size
      response.writeHead(200, headers)
      if (request.method === 'HEAD') {
        response.end()
        return
      }
      createReadStream(filePath).pipe(response)
    } catch {
      response.writeHead(404)
      response.end('Not found')
    }
  }
}

function start() {
  const root = path.join(path.dirname(fileURLToPath(import.meta.url)), 'dist')
  const port = Number.parseInt(process.env.PORT || '3000', 10)
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error('PORT must be an integer between 1 and 65535')
  }

  const server = createServer(createHandler(root))
  server.listen(port, '0.0.0.0', () => {
    console.log(`Frontend server listening on :${port}`)
  })

  for (const signal of ['SIGINT', 'SIGTERM']) {
    process.on(signal, () => server.close(() => process.exit(0)))
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  start()
}
