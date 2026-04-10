function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

function decodeCommonEntities(value: string): string {
  return value
    .replaceAll('&amp;', '&')
    .replaceAll('&quot;', '"')
    .replaceAll('&#39;', "'")
    .replaceAll('&lt;', '<')
    .replaceAll('&gt;', '>')
}

function toSafeUrl(value: string, allowFile: boolean): string {
  const decoded = decodeCommonEntities(value).trim()
  if (!decoded) {
    return ''
  }

  if (allowFile && decoded.toLowerCase().startsWith('data:image/')) {
    return decoded
  }

  if (allowFile) {
    const windowsDrivePath = /^[a-zA-Z]:\\/.test(decoded)
    const windowsDrivePathSlash = /^[a-zA-Z]:\//.test(decoded)
    if (windowsDrivePath) {
      const normalized = decoded.replaceAll('\\', '/')
      return encodeURI(`file:///${normalized}`)
    }
    if (windowsDrivePathSlash) {
      return encodeURI(`file:///${decoded}`)
    }

    const windowsUNCPath = /^\\\\[^\\]+\\[^\\]+/.test(decoded)
    if (windowsUNCPath) {
      const normalized = decoded.replaceAll('\\', '/')
      const withoutPrefix = normalized.replace(/^\/+/, '')
      return encodeURI(`file://${withoutPrefix}`)
    }

    const windowsUNCPathSlash = /^\/\/[^/]+\/[^/]+/.test(decoded)
    if (windowsUNCPathSlash) {
      const withoutPrefix = decoded.replace(/^\/+/, '')
      return encodeURI(`file://${withoutPrefix}`)
    }
  }

  try {
    const url = new URL(decoded)
    const isWeb = url.protocol === 'http:' || url.protocol === 'https:'
    const isFile = allowFile && url.protocol === 'file:'
    return isWeb || isFile ? url.toString() : ''
  } catch {
    return ''
  }
}

function replaceTag(input: string, tag: string, htmlTag: string): string {
  const pattern = new RegExp(`\\[${tag}\\]([\\s\\S]*?)\\[\\/${tag}\\]`, 'gi')
  let current = input
  for (let iteration = 0; iteration < 6; iteration += 1) {
    const next = current.replace(pattern, `<${htmlTag}>$1</${htmlTag}>`)
    if (next === current) {
      return current
    }
    current = next
  }
  return current
}

function replaceSimpleTag(input: string, bbTag: string, htmlTag: string): string {
  return input.replaceAll(new RegExp(`\\[${bbTag}\\]`, 'gi'), `<${htmlTag}>`).replaceAll(new RegExp(`\\[\\/${bbTag}\\]`, 'gi'), `</${htmlTag}>`)
}

function normalizeTableRows(content: string): string[] {
  let normalized = String(content)
    .replaceAll(/\[[^\]]+\]/gi, ' ')
    .replaceAll(/[\u00a0\u1680\u2000-\u200b\u202f\u205f\u2800]{2,}/g, '\n')
    .replaceAll(/(Provinces)([A-Z])/g, '$1\n$2')
    .replaceAll(/(\d[\d,.]*)(?=[A-Z])/g, '$1\n')
    .replaceAll(/\s+/g, ' ')
    .replaceAll(/\n\s+/g, '\n')
    .trim()

  if (!normalized) {
    return []
  }

  if (/\d[\d,.]*[A-Za-z]/.test(normalized)) {
    normalized = normalized.replaceAll(/(\d[\d,.]*)([A-Za-z])/g, '$1\n$2')
  }

  return normalized.split(/\n+/).map((line) => line.trim()).filter((line) => line.length > 0)
}

function parseImplicitTable(content: string): string {
  const rows = normalizeTableRows(content)
  if (rows.length === 0) {
    return ''
  }

  let headerLeft = 'Name'
  let headerRight = 'Value'
  let startIndex = 0
  if (/provinces/i.test(rows[0])) {
    const header = rows[0]
    headerRight = 'Provinces'
    headerLeft = header.replace(/provinces/i, '').trim() || 'Group'
    startIndex = 1
  }

  const bodyRows: string[] = []
  for (let i = startIndex; i < rows.length; i += 1) {
    const row = rows[i]
    const match = row.match(/^(.*?)(\d[\d,.]*)$/)
    if (!match) {
      continue
    }
    const left = match[1].trim()
    const right = match[2].trim()
    if (!left || !right) {
      continue
    }
    bodyRows.push(`<tr><td>${left}</td><td>${right}</td></tr>`)
  }

  if (bodyRows.length === 0) {
    return `<pre>${rows.join('\n')}</pre>`
  }

  return `<table class="steam-table"><tr><th>${headerLeft}</th><th>${headerRight}</th></tr>${bodyRows.join('')}</table>`
}

function replaceTables(input: string): string {
  const tablePattern = /\[table(?:=[^\]]+)?[^\]]*\]([\s\S]*?)\[\/table\]/gi
  return input.replace(tablePattern, (_match, content) => {
    const body = String(content)
    const hasExplicitRows = /\[(tr|td|th)\]/i.test(body)
    if (!hasExplicitRows) {
      return parseImplicitTable(body)
    }

    const normalizedBody = body
      .replaceAll(/\[tr\]/gi, '<tr>')
      .replaceAll(/\[\/tr\]/gi, '</tr>')
      .replaceAll(/\[th\]/gi, '<th>')
      .replaceAll(/\[\/th\]/gi, '</th>')
      .replaceAll(/\[td\]/gi, '<td>')
      .replaceAll(/\[\/td\]/gi, '</td>')

    return `<table class="steam-table">${normalizedBody}</table>`
  })
}

function replaceLists(input: string): string {
  const listPattern = /\[(list|olist)\]([\s\S]*?)\[\/\1\]/gi
  let current = input

  for (let iteration = 0; iteration < 6; iteration += 1) {
    const next = current.replace(listPattern, (_match, listType, content) => {
      const rawContent = String(content)
      const parts = rawContent.split(/\[\*\]/gi).map((part) => part.trim()).filter((part) => part.length > 0)
      if (parts.length === 0) {
        return ''
      }

      const tag = String(listType).toLowerCase() === 'olist' ? 'ol' : 'ul'
      const items = parts.map((part) => `<li>${part}</li>`).join('')
      return `<${tag} class="steam-list">${items}</${tag}>`
    })
    if (next === current) {
      return current
    }
    current = next
  }

  return current
}

function collapseBreaksAroundBlocks(input: string): string {
  return input
    .replaceAll(/<br \/>(\s*)<(\/)?(ul|ol|li|table|tr|th|td|blockquote|pre|h1|h2|h3|h4|h5|h6|hr|img|p)/gi, '<$2$3')
    .replaceAll(/<\/(ul|ol|li|table|tr|th|td|blockquote|pre|h1|h2|h3|h4|h5|h6|hr|img|p)>(\s*)<br \/>/gi, '</$1>')
}

function linkifyPlainUrls(input: string): string {
  const plainURL = /(?<!["'=])(https?:\/\/[^\s<\]]+|steam:\/\/[^\s<\]]+)/gi
  const parts = input.split(/(<[^>]+>)/g)
  return parts
    .map((part) => {
      if (part.startsWith('<') && part.endsWith('>')) {
        return part
      }
      return part.replace(plainURL, (urlText) => {
        const safeHref = toSafeUrl(urlText, false)
        if (!safeHref) {
          return urlText
        }
        return `<a href="${escapeHtml(safeHref)}" target="_blank" rel="noreferrer noopener">${urlText}</a>`
      })
    })
    .join('')
}

export function toDisplayImageSrc(raw: string): string {
  return toSafeUrl(raw, true)
}

export function renderRichDescriptionHtml(raw: string): string {
  const escaped = escapeHtml(raw || '').replaceAll('\r\n', '\n').replaceAll('\r', '\n')

  let html = escaped
  html = replaceTag(html, 'b', 'strong')
  html = replaceTag(html, 'i', 'em')
  html = replaceTag(html, 'u', 'u')
  html = replaceTag(html, 's', 's')
  html = replaceTag(html, 'strike', 's')
  html = replaceTag(html, 'quote', 'blockquote')
  html = replaceTag(html, 'code', 'pre')
  html = replaceTag(html, 'h1', 'h1')
  html = replaceTag(html, 'h2', 'h2')
  html = replaceTag(html, 'h3', 'h3')
  html = replaceTag(html, 'h4', 'h4')
  html = replaceTag(html, 'h5', 'h5')
  html = replaceTag(html, 'h6', 'h6')

  html = replaceTables(html)

  html = replaceLists(html)
  html = replaceSimpleTag(html, 'hr', 'hr')

  html = html.replace(/\[url=([^\]]+)\]([\s\S]*?)\[\/url\]/gi, (_match, href, text) => {
    const safeHref = toSafeUrl(String(href), false)
    return safeHref
      ? `<a href="${escapeHtml(safeHref)}" target="_blank" rel="noreferrer noopener">${text}</a>`
      : text
  })

  html = html.replace(/\[url\]([\s\S]*?)\[\/url\]/gi, (_match, hrefText) => {
    const safeHref = toSafeUrl(String(hrefText), false)
    const label = String(hrefText)
    return safeHref
      ? `<a href="${escapeHtml(safeHref)}" target="_blank" rel="noreferrer noopener">${label}</a>`
      : label
  })

  html = html.replace(/\[img(?:=[^\]]+)?\]([\s\S]*?)\[\/img\]/gi, (_match, src) => {
    const safeSrc = toSafeUrl(String(src), true)
    return safeSrc ? `<img class="steam-desc-image" src="${escapeHtml(safeSrc)}" alt="Workshop image" loading="lazy" />` : ''
  })

  html = linkifyPlainUrls(html)

  html = html.replaceAll('\n', '<br />')
  html = collapseBreaksAroundBlocks(html)
  return html
}

export function renderSteamDescriptionHtml(raw: string): string {
  return renderRichDescriptionHtml(raw)
}

