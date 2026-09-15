import React, { Suspense } from "react"
import Loading from "../Loading/Loading"

const CodeEditorBasic = React.lazy(() => import("./CodeEditorBasic"))

const CodeEditorLazy = (props) => (
  <Suspense fallback={<Loading />}>
    <CodeEditorBasic {...props} />
  </Suspense>
)

export default CodeEditorLazy
