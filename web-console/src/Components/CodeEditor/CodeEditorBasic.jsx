import * as monaco from "monaco-editor"
import Editor, { loader } from "@monaco-editor/react"
import React from "react"

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
