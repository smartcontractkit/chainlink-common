// NOTE: loosely based on: https://github.com/tetratelabs/wazero/blob/1353ca24fef0a57a3a342d75f20357a6e9d3be35/internal/wasip1/errno.go#L14
package host

type Errno = int32

// Note: Below prefers POSIX symbol names over WASI ones, even if the docs are from WASI.
// See https://linux.die.net/man/3/errno
// See https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#variants-1
const (
	// ErrnoSuccess No error occurred. System call completed successfully.
	ErrnoSuccess Errno = iota
	// Errno2big Argument list too long.
	Errno2big
	// ErrnoAcces Permission denied.
	ErrnoAcces
	// ErrnoAddrinuse Address in use.
	ErrnoAddrinuse
	// ErrnoAddrnotavail Address not available.
	ErrnoAddrnotavail
	// ErrnoAfnosupport Address family not supported.
	ErrnoAfnosupport
	// ErrnoAgain Resource unavailable, or operation would block.
	ErrnoAgain
	// ErrnoAlready Connection already in progress.
	ErrnoAlready
	// ErrnoBadf Bad file descriptor.
	ErrnoBadf
	// ErrnoBadmsg Bad message.
	ErrnoBadmsg
	// ErrnoBusy Device or resource busy.
	ErrnoBusy
	// ErrnoCanceled Operation canceled.
	ErrnoCanceled
	// ErrnoChild No child processes.
	ErrnoChild
	// ErrnoConnaborted Connection aborted.
	ErrnoConnaborted
	// ErrnoConnrefused Connection refused.
	ErrnoConnrefused
	// ErrnoConnreset Connection reset.
	ErrnoConnreset
	// ErrnoDeadlk Resource deadlock would occur.
	ErrnoDeadlk
	// ErrnoDestaddrreq Destination address required.
	ErrnoDestaddrreq
	// ErrnoDom Mathematics argument out of domain of function.
	ErrnoDom
	// ErrnoDquot Reserved.
	ErrnoDquot
	// ErrnoExist File exists.
	ErrnoExist
	// ErrnoFault Bad address.
	ErrnoFault
	// ErrnoFbig File too large.
	ErrnoFbig
	// ErrnoHostunreach Host is unreachable.
	ErrnoHostunreach
	// ErrnoIdrm Identifier removed.
	ErrnoIdrm
	// ErrnoIlseq Illegal byte sequence.
	ErrnoIlseq
	// ErrnoInprogress Operation in progress.
	ErrnoInprogress
	// ErrnoIntr Interrupted function.
	ErrnoIntr
	// ErrnoInval Invalid argument.
	ErrnoInval
	// ErrnoIo I/O error.
	ErrnoIo
	// ErrnoIsconn Socket is connected.
	ErrnoIsconn
	// ErrnoIsdir Is a directory.
	ErrnoIsdir
	// ErrnoLoop Too many levels of symbolic links.
	ErrnoLoop
	// ErrnoMfile File descriptor value too large.
	ErrnoMfile
	// ErrnoMlink Too many links.
	ErrnoMlink
	// ErrnoMsgsize Message too large.
	ErrnoMsgsize
	// ErrnoMultihop Reserved.
	ErrnoMultihop
	// ErrnoNametoolong Filename too long.
	ErrnoNametoolong
	// ErrnoNetdown Network is down.
	ErrnoNetdown
	// ErrnoNetreset Connection aborted by network.
	ErrnoNetreset
	// ErrnoNetunreach Network unreachable.
	ErrnoNetunreach
	// ErrnoNfile Too many files open in system.
	ErrnoNfile
	// ErrnoNobufs No buffer space available.
	ErrnoNobufs
	// ErrnoNodev No such device.
	ErrnoNodev
	// ErrnoNoent No such file or directory.
	ErrnoNoent
	// ErrnoNoexec Executable file format error.
	ErrnoNoexec
	// ErrnoNolck No locks available.
	ErrnoNolck
	// ErrnoNolink Reserved.
	ErrnoNolink
	// ErrnoNomem Not enough space.
	ErrnoNomem
	// ErrnoNomsg No message of the desired type.
	ErrnoNomsg
	// ErrnoNoprotoopt No message of the desired type.
	ErrnoNoprotoopt
	// ErrnoNospc No space left on device.
	ErrnoNospc
	// ErrnoNosys function not supported.
	ErrnoNosys
	// ErrnoNotconn The socket is not connected.
	ErrnoNotconn
	// ErrnoNotdir Not a directory or a symbolic link to a directory.
	ErrnoNotdir
	// ErrnoNotempty Directory not empty.
	ErrnoNotempty
	// ErrnoNotrecoverable State not recoverable.
	ErrnoNotrecoverable
	// ErrnoNotsock Not a socket.
	ErrnoNotsock
	// ErrnoNotsup Not supported, or operation not supported on socket.
	ErrnoNotsup
	// ErrnoNotty Inappropriate I/O control operation.
	ErrnoNotty
	// ErrnoNxio No such device or address.
	ErrnoNxio
	// ErrnoOverflow Value too large to be stored in data type.
	ErrnoOverflow
	// ErrnoOwnerdead Previous owner died.
	ErrnoOwnerdead
	// ErrnoPerm Operation not permitted.
	ErrnoPerm
	// ErrnoPipe Broken pipe.
	ErrnoPipe
	// ErrnoProto Protocol error.
	ErrnoProto
	// ErrnoProtonosupport Protocol error.
	ErrnoProtonosupport
	// ErrnoPrototype Protocol wrong type for socket.
	ErrnoPrototype
	// ErrnoRange Result too large.
	ErrnoRange
	// ErrnoRofs Read-only file system.
	ErrnoRofs
	// ErrnoSpipe Invalid seek.
	ErrnoSpipe
	// ErrnoSrch No such process.
	ErrnoSrch
	// ErrnoStale Reserved.
	ErrnoStale
	// ErrnoTimedout Connection timed out.
	ErrnoTimedout
	// ErrnoTxtbsy Text file busy.
	ErrnoTxtbsy
	// ErrnoXdev Cross-device link.
	ErrnoXdev

	// Note: ErrnoNotcapable was removed by WASI maintainers.
	// See https://github.com/WebAssembly/wasi-libc/pull/294
)
