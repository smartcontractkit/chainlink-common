;; A module the host classifies as V2 (it imports env.version_v2) that calls a
;; host function reading guest memory, but deliberately exports no memory at
;; all. No Go compiler emits this shape, so it is hand-written wasm text.
(module
  (import "env" "version_v2" (func $version_v2))
  (import "env" "send_response" (func $send_response (param i32 i32) (result i32)))

  ;; NOTE: no `(memory (export "memory") 1)` here. That is the whole point.

  (func (export "_start")
    i32.const 0
    i32.const 0
    call $send_response
    drop))
