;; Same shape as ../missing/main.wat, except the module does export something
;; named "memory" -- a global, not linear memory. The host resolves that export
;; by name only, so a wrong-typed export is as unusable as an absent one.
(module
  (import "env" "version_v2" (func $version_v2))
  (import "env" "send_response" (func $send_response (param i32 i32) (result i32)))

  ;; An export named "memory" that is not a memory.
  (global (export "memory") i32 (i32.const 0))

  (func (export "_start")
    i32.const 0
    i32.const 0
    call $send_response
    drop))
