import React, { useMemo } from "react"

export type MarkdownHtmlClassNames = {
  a?: string
  blockquote?: string
  code?: string
  em?: string
  h1?: string
  h2?: string
  h3?: string
  h4?: string
  h5?: string
  h6?: string
  hr?: string
  img?: string
  li?: string
  ol?: string
  p?: string
  pre?: string
  strong?: string
  table?: string
  tbody?: string
  td?: string
  th?: string
  thead?: string
  tr?: string
  ul?: string
}

// An <img> src from untrusted brain output fires an out-of-origin request the
// moment it renders (a tracking/exfil pixel needs no click), so only same-origin
// sources are rendered as images. Relative URLs resolve to the current origin;
// anything else (absolute cross-origin, protocol-relative, data:, or unparseable)
// is treated as foreign. When there's no window (SSR) we can't verify origin, so
// we err on the side of not rendering.
const isSameOriginImage = (src: string): boolean => {
  if (typeof window === "undefined") {
    return false
  }
  try {
    return new URL(src, window.location.href).origin === window.location.origin
  } catch {
    return false
  }
}

const InlineMarkdown: React.FC<{
  text: string
  classNames: MarkdownHtmlClassNames
}> = ({ text, classNames }) => {
  const nodes: React.ReactNode[] = []
  // Underscore emphasis only fires at word boundaries (GFM intraword rule), so
  // `snake_case_word` stays literal; asterisks fire anywhere. URLs allow one
  // level of balanced parens so `.../Foo_(bar)` doesn't truncate at the `)`.
  const pattern =
    /(`[^`]+`)|(!\[[^\]]*\]\((?:[^()]|\([^()]*\))*\))|(\*\*.+?\*\*|(?<![A-Za-z0-9])__.+?__(?![A-Za-z0-9]))|(\*.+?\*|(?<![A-Za-z0-9])_.+?_(?![A-Za-z0-9]))|(\[[^\]]+\]\((?:[^()]|\([^()]*\))*\))/g
  let lastIndex = 0
  let key = 0
  let match: RegExpExecArray | null

  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(text.slice(lastIndex, match.index))
    }
    const [full, code, image, bold, italic, link] = match
    if (code) {
      nodes.push(
        <code key={key} className={classNames.code}>
          {code.slice(1, -1)}
        </code>,
      )
    } else if (image) {
      const parts = /!\[([^\]]*)\]\(((?:[^()]|\([^()]*\))*)\)/.exec(image)
      if (parts) {
        const alt = parts[1] ?? ""
        const src = parts[2] ?? ""
        if (isSameOriginImage(src)) {
          nodes.push(
            <img key={key} src={src} alt={alt} className={classNames.img} />,
          )
        } else {
          // Cross-origin src: render the raw markdown as a placeholder instead
          // of firing an out-of-origin request.
          nodes.push(full)
        }
      } else {
        nodes.push(full)
      }
    } else if (bold) {
      nodes.push(
        <strong key={key} className={classNames.strong}>
          <InlineMarkdown text={bold.slice(2, -2)} classNames={classNames} />
        </strong>,
      )
    } else if (italic) {
      nodes.push(
        <em key={key} className={classNames.em}>
          <InlineMarkdown text={italic.slice(1, -1)} classNames={classNames} />
        </em>,
      )
    } else if (link) {
      const parts = /\[([^\]]+)\]\(((?:[^()]|\([^()]*\))*)\)/.exec(link)
      if (parts) {
        const label = parts[1] ?? ""
        const url = parts[2] ?? ""
        const clickable = /^https?:\/\//.test(url)
        nodes.push(
          <React.Fragment key={key}>
            {`[${label}](`}
            {clickable ? (
              <a
                href={url}
                target="_blank"
                rel="noreferrer"
                className={classNames.a}
              >
                {url}
              </a>
            ) : (
              url
            )}
            {")"}
          </React.Fragment>,
        )
      } else {
        nodes.push(full)
      }
    }
    lastIndex = pattern.lastIndex
    key += 1
  }

  if (lastIndex < text.length) {
    nodes.push(text.slice(lastIndex))
  }
  return <>{nodes}</>
}

// A GFM table delimiter row separates the header from the body, e.g.
// `| --- | :--: | ---: |`. Each cell is dashes with optional alignment colons.
const isTableDelimiterRow = (line: string): boolean =>
  line.includes("|") && /^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$/.test(line)

// A table begins where a header row (containing a pipe) is immediately followed
// by a delimiter row.
const isTableStart = (lines: string[], index: number): boolean =>
  (lines[index] ?? "").includes("|") &&
  isTableDelimiterRow(lines[index + 1] ?? "")

const isBlockStart = (line: string): boolean =>
  /^```/.test(line) ||
  /^#{1,6}\s/.test(line) ||
  /^>\s?/.test(line) ||
  /^(-{3,}|\*{3,}|_{3,})\s*$/.test(line) ||
  /^\s*\d+\.\s+/.test(line) ||
  /^\s*[-*+]\s+/.test(line)

// Split a table row into trimmed cells, dropping the optional leading/trailing
// pipes.
const parseTableRow = (row: string): string[] =>
  row
    .trim()
    .replace(/^\|/, "")
    .replace(/\|$/, "")
    .split("|")
    .map((cell) => cell.trim())

export type MarkdownViewProps = {
  content: string
  className?: string
  htmlClassNames?: MarkdownHtmlClassNames
}

const MarkdownView: React.FC<MarkdownViewProps> = ({
  content,
  className,
  htmlClassNames,
}) => {
  const blocks = useMemo<React.ReactNode[]>(() => {
    const classNames = htmlClassNames || {}
    const lines = content.split("\n")
    const result: React.ReactNode[] = []
    let lineIndex = 0

    while (lineIndex < lines.length) {
      const line = lines[lineIndex] ?? ""
      // Each block starts at a distinct, increasing line index, so the start
      // index is a stable key.
      const key = lineIndex

      // Skip blank lines between blocks.
      if (line.trim() === "") {
        lineIndex += 1
        continue
      }

      // Fenced code block.
      const fence = /^```(\w*)\s*$/.exec(line)
      if (fence) {
        const language = fence[1] ?? ""
        const codeLines: string[] = []
        lineIndex += 1
        while (
          lineIndex < lines.length &&
          !/^```\s*$/.test(lines[lineIndex] ?? "")
        ) {
          codeLines.push(lines[lineIndex] ?? "")
          lineIndex += 1
        }
        lineIndex += 1 // closing fence
        // Surface the fence language as a `language-*` class so downstream
        // highlighters (and callers) can key off it instead of losing it.
        const codeClassName =
          [classNames.code, language && `language-${language}`]
            .filter(Boolean)
            .join(" ") || undefined
        result.push(
          <pre key={key} className={classNames.pre}>
            <code className={codeClassName}>{codeLines.join("\n")}</code>
          </pre>,
        )
        continue
      }

      // Heading.
      const heading = /^(#{1,6})\s+(.*)$/.exec(line)
      if (heading) {
        const level = (heading[1] ?? "").length
        const headingText = heading[2] ?? ""
        const inline = (
          <InlineMarkdown text={headingText} classNames={classNames} />
        )
        switch (level) {
          case 1:
            result.push(
              <h1 key={key} className={classNames.h1}>
                {inline}
              </h1>,
            )
            break
          case 2:
            result.push(
              <h2 key={key} className={classNames.h2}>
                {inline}
              </h2>,
            )
            break
          case 3:
            result.push(
              <h3 key={key} className={classNames.h3}>
                {inline}
              </h3>,
            )
            break
          case 4:
            result.push(
              <h4 key={key} className={classNames.h4}>
                {inline}
              </h4>,
            )
            break
          case 5:
            result.push(
              <h5 key={key} className={classNames.h5}>
                {inline}
              </h5>,
            )
            break
          default:
            result.push(
              <h6 key={key} className={classNames.h6}>
                {inline}
              </h6>,
            )
            break
        }
        lineIndex += 1
        continue
      }

      // Horizontal rule.
      if (/^(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
        result.push(<hr key={key} className={classNames.hr} />)
        lineIndex += 1
        continue
      }

      // Blockquote.
      if (/^>\s?/.test(line)) {
        const quoteLines: string[] = []
        while (
          lineIndex < lines.length &&
          /^>\s?/.test(lines[lineIndex] ?? "")
        ) {
          quoteLines.push((lines[lineIndex] ?? "").replace(/^>\s?/, ""))
          lineIndex += 1
        }
        result.push(
          <blockquote key={key} className={classNames.blockquote}>
            <InlineMarkdown
              text={quoteLines.join(" ")}
              classNames={classNames}
            />
          </blockquote>,
        )
        continue
      }

      // Ordered or unordered list.
      const ordered = /^\s*\d+\.\s+/.test(line)
      const unordered = /^\s*[-*+]\s+/.test(line)
      if (ordered || unordered) {
        const items: string[] = []
        const itemPattern = ordered ? /^\s*\d+\.\s+/ : /^\s*[-*+]\s+/
        while (
          lineIndex < lines.length &&
          itemPattern.test(lines[lineIndex] ?? "")
        ) {
          items.push((lines[lineIndex] ?? "").replace(itemPattern, ""))
          lineIndex += 1
        }
        const listItems = items.map((item, idx) => (
          <li key={idx} className={classNames.li}>
            <InlineMarkdown text={item} classNames={classNames} />
          </li>
        ))
        result.push(
          ordered ? (
            <ol key={key} className={classNames.ol}>
              {listItems}
            </ol>
          ) : (
            <ul key={key} className={classNames.ul}>
              {listItems}
            </ul>
          ),
        )
        continue
      }

      // Table.
      if (isTableStart(lines, lineIndex)) {
        const headerCells = parseTableRow(lines[lineIndex] ?? "")
        const alignments = parseTableRow(lines[lineIndex + 1] ?? "").map(
          (cell): React.CSSProperties["textAlign"] => {
            const left = cell.startsWith(":")
            const right = cell.endsWith(":")
            if (left && right) {
              return "center"
            }
            if (right) {
              return "right"
            }
            if (left) {
              return "left"
            }
            return undefined
          },
        )
        lineIndex += 2

        const bodyRows: string[][] = []
        while (
          lineIndex < lines.length &&
          (lines[lineIndex] ?? "").includes("|") &&
          (lines[lineIndex] ?? "").trim() !== ""
        ) {
          bodyRows.push(parseTableRow(lines[lineIndex] ?? ""))
          lineIndex += 1
        }

        result.push(
          <table key={key} className={classNames.table}>
            <thead className={classNames.thead}>
              <tr className={classNames.tr}>
                {headerCells.map((cell, idx) => (
                  <th
                    key={idx}
                    className={classNames.th}
                    style={{ textAlign: alignments[idx] }}
                  >
                    <InlineMarkdown text={cell} classNames={classNames} />
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className={classNames.tbody}>
              {bodyRows.map((row, rowIdx) => (
                <tr key={rowIdx} className={classNames.tr}>
                  {row.map((cell, idx) => (
                    <td
                      key={idx}
                      className={classNames.td}
                      style={{ textAlign: alignments[idx] }}
                    >
                      <InlineMarkdown text={cell} classNames={classNames} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>,
        )
        continue
      }

      // Paragraph: collect consecutive non-blank lines until the next block.
      const paraLines: string[] = []
      while (
        lineIndex < lines.length &&
        (lines[lineIndex] ?? "").trim() !== "" &&
        !isBlockStart(lines[lineIndex] ?? "") &&
        !isTableStart(lines, lineIndex)
      ) {
        paraLines.push(lines[lineIndex] ?? "")
        lineIndex += 1
      }
      result.push(
        <p key={key} className={classNames.p}>
          <InlineMarkdown text={paraLines.join(" ")} classNames={classNames} />
        </p>,
      )
    }

    return result
  }, [content, htmlClassNames])

  return <div className={className}>{blocks}</div>
}

export default MarkdownView
