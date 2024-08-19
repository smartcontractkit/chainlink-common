{
  commit: {
    author: {
      email: 'jmank88@gmail.com',
      name: 'Jordan Krage',
      username: 'jmank88'
    },
    committer: {
      email: 'noreply@github.com',
      name: 'GitHub',
      username: 'web-flow'
    },
    distinct: true,
    id: 'b648e7569c8b83c6301364824a8b252e990a9559',
    message: 'script/lint.sh: fix output (#708)',
    timestamp: '2024-08-16T16:14:41-05:00',
    tree_id: '80d5d98261d8e50d46e92f733477b0958a54e9fa',
    url: 'https://github.com/smartcontractkit/chainlink-common/commit/b648e7569c8b83c6301364824a8b252e990a9559'
  },
  date: 1724085658709,
  tool: 'go',
  benches: [
    {
      name: 'BenchmarkKeystore_Sign/nop/in-process',
      value: 486.3,
      unit: 'ns/op',
      extra: '2587179 times\n4 procs'
    },
    {
      name: 'BenchmarkKeystore_Sign/hex/in-process',
      value: 567.8,
      unit: 'ns/op',
      extra: '2217537 times\n4 procs'
    },
    {
      name: 'BenchmarkKeystore_Sign/ed25519/in-process',
      value: 28311,
      unit: 'ns/op',
      extra: '42441 times\n4 procs'
    }
  ]
}
