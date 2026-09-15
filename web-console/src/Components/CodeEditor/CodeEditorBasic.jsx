// editor.all registers context menu, suggest, hover, find — not included in editor.api alone.
import "monaco-editor/esm/vs/editor/editor.all.js"
import * as monaco from "monaco-editor/esm/vs/editor/editor.api"
import "monaco-editor/esm/vs/basic-languages/html/html.contribution"
import "monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution"
import "monaco-editor/esm/vs/basic-languages/xml/xml.contribution"
import "monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution"
import "monaco-editor/esm/vs/language/json/monaco.contribution"
import "monaco-editor/esm/vs/language/typescript/monaco.contribution"
import Editor, { loader } from "@monaco-editor/react"
import React from "react"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker"

// Workers are bundled locally (no CDN) so the editor works air-gapped.
self.MonacoEnvironment = {
  getWorker(_workerId, label) {
    if (label === "json") {
      return new jsonWorker()
    }
    if (label === "html" || label === "handlebars" || label === "razor") {
      return new htmlWorker()
    }
    if (label === "typescript" || label === "javascript") {
      return new tsWorker()
    }
    return new editorWorker()
  },
}

const consoleTheme = {
  base: "vs-dark",
  inherit: true,
  rules: [
    { token: "number", foreground: "ace12e" },
    { token: "type", foreground: "73bcf7" },
    { token: "string", foreground: "f0ab00" },
    { token: "keyword", foreground: "cbc0ff" },
  ],
  colors: {
    "editor.background": "#151515",
    "editorGutter.background": "#292e34",
    "editorLineNumber.activeForeground": "#fff",
    "editorLineNumber.foreground": "#f0f0f0",
  },
}

// load required files from local, not from CDN registry
// see: https://github.com/suren-atoyan/monaco-react/blob/master/README.md#loader-config
// https://github.com/suren-atoyan/monaco-react/issues/327
loader.config({ monaco })

loader
  .init()
  .then((monaco) => {
    monaco.editor.defineTheme("console", consoleTheme)
  })
  .catch((error) => console.error("An error occurred during initialization of Monaco: ", error))

const defaultOptions = {
  selectOnLineNumbers: true,
  scrollBeyondLastLine: false,
  contextmenu: true,
  autoIndent: "full",
  cursorBlinking: "phase",
  smoothScrolling: true,
  tabSize: 2,
  fontSize: 15,
  minimap: {
    enabled: false,
  },
  readOnly: true,
}

const editor = ({
  language = "yaml",
  options = {},
  height = "71vh",
  data = "",
  handleEditorOnMount = () => {},
}) => {
  const finalOptions = {
    ...defaultOptions, // default options
    ...options, // supplied options
  }
  return (
    <Editor
      height={height}
      language={language}
      theme="console"
      value={data}
      options={finalOptions}
      onMount={handleEditorOnMount}
    />
  )
}

export default editor
