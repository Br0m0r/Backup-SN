import assert from 'node:assert/strict'
import { mkdtemp, mkdir, rm, writeFile } from 'node:fs/promises'
import { createServer } from 'node:http'
import os from 'node:os'
import path from 'node:path'
import { after, before, test } from 'node:test'

import { createHandler } from './server.mjs'

let root
let server
let baseURL

before(async () => {
  root = await mkdtemp(path.join(os.tmpdir(), 'social-network-frontend-'))
  await mkdir(path.join(root, 'assets'))
  await writeFile(path.join(root, 'index.html'), '<main>app</main>')
  await writeFile(path.join(root, 'assets', 'app.js'), 'console.log("app")')

  server = createServer(createHandler(root))
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve))
  baseURL = `http://127.0.0.1:${server.address().port}`
})

after(async () => {
  await new Promise((resolve) => server.close(resolve))
  await rm(root, { recursive: true, force: true })
})

test('serves immutable assets', async () => {
  const response = await fetch(`${baseURL}/assets/app.js`)
  assert.equal(response.status, 200)
  assert.equal(response.headers.get('cache-control'), 'public, max-age=31536000, immutable')
  assert.equal(await response.text(), 'console.log("app")')
})

test('falls back to the SPA entry point', async () => {
  const response = await fetch(`${baseURL}/profile/7`)
  assert.equal(response.status, 200)
  assert.equal(response.headers.get('cache-control'), 'no-cache')
  assert.equal(await response.text(), '<main>app</main>')
})

test('rejects unsupported methods', async () => {
  const response = await fetch(`${baseURL}/`, { method: 'POST' })
  assert.equal(response.status, 405)
  assert.equal(response.headers.get('allow'), 'GET, HEAD')
})
