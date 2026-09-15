// PatternFly Topology (and other CJS browser builds) assign to Node's `global`.
if (typeof globalThis !== "undefined" && typeof globalThis.global === "undefined") {
  globalThis.global = globalThis
}
