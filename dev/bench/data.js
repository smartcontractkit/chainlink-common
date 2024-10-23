window.BENCHMARK_DATA = {
  "lastUpdate": 1729693550650,
  "repoUrl": "https://github.com/smartcontractkit/chainlink-common",
  "entries": {
    "Benchmark": [
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "54e825b236467c615d36431922336eb1a07186f8",
          "message": "updating benchmark action to use gh-pages (#747)",
          "timestamp": "2024-09-02T13:42:19-04:00",
          "tree_id": "c977893b533a50aca136143463627dd91d2fe479",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/54e825b236467c615d36431922336eb1a07186f8"
        },
        "date": 1725298993139,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465.8,
            "unit": "ns/op",
            "extra": "2612760 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.8,
            "unit": "ns/op",
            "extra": "2318792 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28207,
            "unit": "ns/op",
            "extra": "42346 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "121895364+martin-cll@users.noreply.github.com",
            "name": "martin-cll",
            "username": "martin-cll"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6488292a85e36d58e832b6de7af75fbecd035d22",
          "message": "MERC-6190: Remove bid/ask fields from Mercury v4 schema (#736)\n\n* Remove bid/ask fields from Mercury v4 schema\r\n\r\n* Add back deleted fields as deprecated",
          "timestamp": "2024-09-03T14:42:00-04:00",
          "tree_id": "0580aaed6f02476864e3a5f436cdcd5b725fec0d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6488292a85e36d58e832b6de7af75fbecd035d22"
        },
        "date": 1725388973376,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.5,
            "unit": "ns/op",
            "extra": "2213182 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.2,
            "unit": "ns/op",
            "extra": "2316528 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28276,
            "unit": "ns/op",
            "extra": "42405 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "32529249+silaslenihan@users.noreply.github.com",
            "name": "Silas Lenihan",
            "username": "silaslenihan"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "00ac29d259a7287e782e55a313238fc9c1b6253c",
          "message": "ChainComponents Generalized Interface Tests (#664)\n\n* Added ChainWriter to Generalized CR Tests\r\n\r\n* started uint setting\r\n\r\n* Enabled uint writes / historical testing via ChainWriter\r\n\r\n* enabled batch writing\r\n\r\n* slowed transaction checking timer\r\n\r\n* cleanup\r\n\r\n* Removed GenerateBlockstilConfidenceLevel in favor on ChainWriter's finality tracking\r\n\r\n* Added concurrent batch sending\r\n\r\n* Reverted batch concurrency\r\n\r\n* lint fix\r\n\r\n* Fixed local common tests\r\n\r\n* lint fix\r\n\r\n* Refactored chainreadertests to be chaincomponentstests\r\n\r\n* cleanup\r\n\r\n* added dirty\r\ncontracts back in\r\n\r\n* icnrease context timeout for waitForTransactionStatus\r\n\r\n* reverted refactor in loop tester\r\n\r\n* Delete gotest.log\r\n\r\n* refactored chain components to contract reader\r\n\r\n* Added comments and removed print statements\r\n\r\n* Delete gotest.log\r\n\r\n* moved method names to structs\r\n\r\n* Added comment to DirtyContracts() method signature",
          "timestamp": "2024-09-04T09:57:53-04:00",
          "tree_id": "cbfd501832716613ddf57e9122acb2594dd7fbd3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/00ac29d259a7287e782e55a313238fc9c1b6253c"
        },
        "date": 1725458325828,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.7,
            "unit": "ns/op",
            "extra": "2588266 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 547.2,
            "unit": "ns/op",
            "extra": "2307762 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28311,
            "unit": "ns/op",
            "extra": "42404 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1702865+kidambisrinivas@users.noreply.github.com",
            "name": "Sri Kidambi",
            "username": "kidambisrinivas"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1128f33dc70bd0fc787679c6c862079bd694697e",
          "message": "Expose ContractReader and ChainWriter of relayer in relayerSet (#749)\n\n* Embed relayer::relayer in relayerset::relayer to expose chainReader\r\n\r\n* Add NewContractReader and NewChainWriter functionality in relayerset\r\n\r\n* Minor change\r\n\r\n* Addressed PR comment\r\n\r\n* Use .mockery.yaml instead of go:generate\r\n\r\n* Add documentation and comments",
          "timestamp": "2024-09-05T14:16:01+01:00",
          "tree_id": "dc77244db77eb66b004f683df9837f99904f967c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1128f33dc70bd0fc787679c6c862079bd694697e"
        },
        "date": 1725542215097,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.4,
            "unit": "ns/op",
            "extra": "2287370 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.4,
            "unit": "ns/op",
            "extra": "2073802 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28302,
            "unit": "ns/op",
            "extra": "42394 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1229e6bc456fc2bf56be07823c3f554e99fe8f01",
          "message": "Remove redundant interface type (#750)",
          "timestamp": "2024-09-05T14:52:07+01:00",
          "tree_id": "4786246f4f79b504c5f9d114eec9f8f098b36df3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1229e6bc456fc2bf56be07823c3f554e99fe8f01"
        },
        "date": 1725544380701,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.8,
            "unit": "ns/op",
            "extra": "2626786 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 537.3,
            "unit": "ns/op",
            "extra": "2338347 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42394 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2ff0f9628f4d5957d52061507971c7820080caee",
          "message": "Generate the mocks for capabilities (#725)",
          "timestamp": "2024-09-05T10:59:27-04:00",
          "tree_id": "f3deb4d456fa4f88a87f2ed92c24961010175c05",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2ff0f9628f4d5957d52061507971c7820080caee"
        },
        "date": 1725548438072,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.3,
            "unit": "ns/op",
            "extra": "2400298 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.5,
            "unit": "ns/op",
            "extra": "2343837 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28288,
            "unit": "ns/op",
            "extra": "42452 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "14a5c7af361f61afb7e02360e286012770edf8f0",
          "message": "Change the Execute Capability API to sync (#748)\n\nChange capability API to synchronous call",
          "timestamp": "2024-09-06T14:22:54+01:00",
          "tree_id": "9c08cc7ef26d9c983ffa136950a9a73cb1993eb6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/14a5c7af361f61afb7e02360e286012770edf8f0"
        },
        "date": 1725629037009,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.9,
            "unit": "ns/op",
            "extra": "2593006 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521.3,
            "unit": "ns/op",
            "extra": "2304212 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28584,
            "unit": "ns/op",
            "extra": "42316 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "blaz@mxxn.io",
            "name": "Blaž Hrastnik",
            "username": "archseer"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b759a57ce2590354ec01a0fe0641cd6e813d8005",
          "message": "capabilities: mercury_trigger: Expose meta so it can be set by mock-trigger (#692)",
          "timestamp": "2024-09-09T17:55:00+09:00",
          "tree_id": "e1fe0267ffcc382f8310b8182e9d7054cd54a803",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b759a57ce2590354ec01a0fe0641cd6e813d8005"
        },
        "date": 1725872196011,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.7,
            "unit": "ns/op",
            "extra": "2594806 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.5,
            "unit": "ns/op",
            "extra": "2315995 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28478,
            "unit": "ns/op",
            "extra": "42538 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "663388d38293604368eaf6c76508729e3659c259",
          "message": "ContractReader Multiple Read Addresses (#603)\n\n* Contract Reader Multiple Address Support\r\n\r\nComplete support for multiple address bindings across all `ContractReader` methods.",
          "timestamp": "2024-09-09T09:12:52-05:00",
          "tree_id": "cecad331274333fdf830a8645200b685cc98c0c1",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/663388d38293604368eaf6c76508729e3659c259"
        },
        "date": 1725891243459,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 478.6,
            "unit": "ns/op",
            "extra": "2494620 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.9,
            "unit": "ns/op",
            "extra": "2318055 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28235,
            "unit": "ns/op",
            "extra": "42548 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1b0938c4678b7a4bae8f8fb93cda2a66cb90bdff",
          "message": "pkg/loop: use beholder (#696)",
          "timestamp": "2024-09-09T12:21:56-05:00",
          "tree_id": "5422bbcca3f0aead1d9a7e0718b8f60f23ca8d39",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1b0938c4678b7a4bae8f8fb93cda2a66cb90bdff"
        },
        "date": 1725902578689,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.2,
            "unit": "ns/op",
            "extra": "2173896 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507.7,
            "unit": "ns/op",
            "extra": "2354368 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28921,
            "unit": "ns/op",
            "extra": "42279 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3c6df3a1efcec288c776d76310e74a68cc4ad8e5",
          "message": "Add unit test runner for workflows (#751)",
          "timestamp": "2024-09-09T14:53:56-04:00",
          "tree_id": "f396dd500959f9c60fb0ce245617a2f9ccf006bd",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3c6df3a1efcec288c776d76310e74a68cc4ad8e5"
        },
        "date": 1725908101399,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.4,
            "unit": "ns/op",
            "extra": "2627319 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.6,
            "unit": "ns/op",
            "extra": "2204683 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28251,
            "unit": "ns/op",
            "extra": "42492 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5d42fb7622b71d61f634ddccd8f85ba12caa3f32",
          "message": "code cleanup on contract reader (#753)\n\n* code cleanup on contract reader\r\n\r\n* run make generate",
          "timestamp": "2024-09-09T17:34:13-05:00",
          "tree_id": "ca21a632e1832f440be624287924d7cc295a9fd4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5d42fb7622b71d61f634ddccd8f85ba12caa3f32"
        },
        "date": 1725921317550,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.8,
            "unit": "ns/op",
            "extra": "2623986 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508,
            "unit": "ns/op",
            "extra": "2335743 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28320,
            "unit": "ns/op",
            "extra": "42421 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3736fe7fda7a9477bf1db2fb5730a99a80315fca",
          "message": "fix install protoc (#755)\n\ninstall-protoc.sh needs to check the `protoc` version that will exceute, rather than an explicit path to a particular binary.",
          "timestamp": "2024-09-10T06:46:36-05:00",
          "tree_id": "c8098f63dd237446e96ba0dc1bb810a30345320d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3736fe7fda7a9477bf1db2fb5730a99a80315fca"
        },
        "date": 1725968853784,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.5,
            "unit": "ns/op",
            "extra": "2615790 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508,
            "unit": "ns/op",
            "extra": "2348409 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28316,
            "unit": "ns/op",
            "extra": "42470 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1416262+bolekk@users.noreply.github.com",
            "name": "Bolek",
            "username": "bolekk"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ce25c4b28676b90cd4467a02b36309b8b9df763a",
          "message": "[KS-365] New config fields for trigger event batching (#757)\n\nmaxBatchSize + batchCollectionPeriod",
          "timestamp": "2024-09-10T07:49:26-07:00",
          "tree_id": "10f74c7d7ea6035f3dbd635c9f5a772a935efb89",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ce25c4b28676b90cd4467a02b36309b8b9df763a"
        },
        "date": 1725979824476,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 481.1,
            "unit": "ns/op",
            "extra": "2562690 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.3,
            "unit": "ns/op",
            "extra": "2257256 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28276,
            "unit": "ns/op",
            "extra": "42415 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "makramkd@users.noreply.github.com",
            "name": "Makram",
            "username": "makramkd"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "33f91788deb60920bdd9a5fc930f3b7f487e066d",
          "message": "pkg/types/ccipocr3: add rmn sigs to report (#758)\n\n* pkg/types/ccipocr3: add rmn sigs to report\r\n\r\nAdditionally, mark NewCommitPluginReport as deprecated.\r\n\r\n* add ccip-offchain as codeowners of ccip types\r\n\r\n* remove usage of deprecated func, update tests\r\n\r\n* bump doc",
          "timestamp": "2024-09-10T19:34:10+04:00",
          "tree_id": "8abb79eb727cabe2d55e2a16f57f78deb93bca66",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/33f91788deb60920bdd9a5fc930f3b7f487e066d"
        },
        "date": 1725982513437,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463,
            "unit": "ns/op",
            "extra": "2576943 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 538.9,
            "unit": "ns/op",
            "extra": "2346670 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42408 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "57732589+ilija42@users.noreply.github.com",
            "name": "ilija42",
            "username": "ilija42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ed9f50de73222b8a2154366ce9a72c95aa2f97d8",
          "message": "Rename Chain Reader to Contract Reader (#759)\n\nCo-authored-by: Jordan Krage <jmank88@gmail.com>",
          "timestamp": "2024-09-10T17:40:10+02:00",
          "tree_id": "10cadb482a66f669be55670e444d51262c232b53",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ed9f50de73222b8a2154366ce9a72c95aa2f97d8"
        },
        "date": 1725982914261,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.9,
            "unit": "ns/op",
            "extra": "2589003 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.7,
            "unit": "ns/op",
            "extra": "2339268 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29001,
            "unit": "ns/op",
            "extra": "42430 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9446397802fe4abcfa2ad1fe87e642d571f0506f",
          "message": "[CAPPL-37] Implement Compute runner (#752)\n\n* [CAPPL-37] Add runner implementation\r\n\r\n* lint errors",
          "timestamp": "2024-09-11T14:48:39+01:00",
          "tree_id": "52e7528bf4aeaf7bfb66fd186bcd5f1c2e2c40bb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9446397802fe4abcfa2ad1fe87e642d571f0506f"
        },
        "date": 1726062622197,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465.7,
            "unit": "ns/op",
            "extra": "2559214 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.3,
            "unit": "ns/op",
            "extra": "2291344 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28422,
            "unit": "ns/op",
            "extra": "42436 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "57732589+ilija42@users.noreply.github.com",
            "name": "ilija42",
            "username": "ilija42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "27a338bd3e60957e230c096234640b93f6b79b4c",
          "message": "[BCF-3381] - Add LatestHead to ChainService (#760)\n\n* Add LatestHead to ChainService\r\n\r\n* Rename chain agnostic head struct Identifier field to height\r\n\r\n* Add comment for chain agnostic Head struct Timestamp field",
          "timestamp": "2024-09-11T15:58:47+02:00",
          "tree_id": "5397d61026300e94ee60125892090ee4d4c3a7f9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/27a338bd3e60957e230c096234640b93f6b79b4c"
        },
        "date": 1726063192723,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 484.8,
            "unit": "ns/op",
            "extra": "2533032 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.7,
            "unit": "ns/op",
            "extra": "2308606 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28285,
            "unit": "ns/op",
            "extra": "42412 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4836d1d7f16bdd92d561c72bc5094ed2a412aeb2",
          "message": "Update to a newer version of my go-jsonschema fork to support creating byte types. Also remove ID and TriggerType from streams, since it's not part of the payload. (#754)",
          "timestamp": "2024-09-11T11:28:14-04:00",
          "tree_id": "e5415e9031cf233de0b43fef6a29ad95be8c5fc9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4836d1d7f16bdd92d561c72bc5094ed2a412aeb2"
        },
        "date": 1726068594502,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.1,
            "unit": "ns/op",
            "extra": "2639503 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.4,
            "unit": "ns/op",
            "extra": "2327658 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28475,
            "unit": "ns/op",
            "extra": "42519 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d00d5184ffaa33e91ae3e491db377d5a842572ef",
          "message": "pkg/loop: include tracing attributes when enabled (#763)",
          "timestamp": "2024-09-11T13:18:00-05:00",
          "tree_id": "0d38c50ec017e257695fd6f99fb75eadaa3349f3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d00d5184ffaa33e91ae3e491db377d5a842572ef"
        },
        "date": 1726078739713,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.5,
            "unit": "ns/op",
            "extra": "2611869 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.3,
            "unit": "ns/op",
            "extra": "2286018 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28272,
            "unit": "ns/op",
            "extra": "42433 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "clement.erena78@gmail.com",
            "name": "Clement",
            "username": "Atrax1"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "36feb2504f38af28b633793d9b8d12fffb0b2a13",
          "message": "feat(observability-lib): deploy logic in grafana module (#761)",
          "timestamp": "2024-09-12T10:53:11+02:00",
          "tree_id": "ba6628b61f9bb313e4ca00fe81a3351dcf0c8b13",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/36feb2504f38af28b633793d9b8d12fffb0b2a13"
        },
        "date": 1726131250526,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457.8,
            "unit": "ns/op",
            "extra": "2566989 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.3,
            "unit": "ns/op",
            "extra": "2343920 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28245,
            "unit": "ns/op",
            "extra": "42501 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "dimitrios.kouveris@smartcontract.com",
            "name": "dimitris",
            "username": "dimkouv"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "aa383c8c46947660775549c480c937bf3e387bde",
          "message": "ccip - Add RMNCrypto interface + Remove unused constructor (#766)\n\n* add RMNCrypto interface\r\n\r\n* re-use ecdsa sigs struct\r\n\r\n* use Bytes\r\n\r\n* add comments\r\n\r\n* rm unused constructor",
          "timestamp": "2024-09-13T14:04:29+04:00",
          "tree_id": "d000ac369f7044db965cbeb7e183c26149159791",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aa383c8c46947660775549c480c937bf3e387bde"
        },
        "date": 1726221928950,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.6,
            "unit": "ns/op",
            "extra": "2602830 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.3,
            "unit": "ns/op",
            "extra": "2326138 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28250,
            "unit": "ns/op",
            "extra": "42456 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ce5d667907ce901f088b1ac04fb06ea6295b8c9f",
          "message": "embeddable implementation of contract reader (#680)\n\n* embeddable implementation of contract reader that returns unimplemented errors\r\n\r\n* update with latest reader changes\r\n\r\n* added must embed to interface to enforce implementation\r\n\r\n* apply unimplemented to existing contract readers",
          "timestamp": "2024-09-13T11:19:26-05:00",
          "tree_id": "4046687768b599b6a242cfb243c34653b1aaa339",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ce5d667907ce901f088b1ac04fb06ea6295b8c9f"
        },
        "date": 1726244438042,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.7,
            "unit": "ns/op",
            "extra": "2595174 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.1,
            "unit": "ns/op",
            "extra": "2312088 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28210,
            "unit": "ns/op",
            "extra": "40132 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "50029043+aalu1418@users.noreply.github.com",
            "name": "Aaron Lu",
            "username": "aalu1418"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "44d96950c886f211a184f18da973ea9baa4f88a1",
          "message": "monitoring: add jitter to source polling (#768)\n\n* monitoring: add jitter to source polling\r\n\r\n* use services.TickerConfig\r\n\r\n* replace deprecated method",
          "timestamp": "2024-09-13T13:19:49-06:00",
          "tree_id": "156a86e2a2164b6dceabdd69f3f45f4bf0551569",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/44d96950c886f211a184f18da973ea9baa4f88a1"
        },
        "date": 1726255259965,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.4,
            "unit": "ns/op",
            "extra": "2518092 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513,
            "unit": "ns/op",
            "extra": "2354822 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28247,
            "unit": "ns/op",
            "extra": "42488 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e1fc24838e09f88e60ca8b3f7b305562df11761a",
          "message": "adding general capability metrics (#756)\n\n* adding general capability metrics\r\n\r\n* adding first pass dashboard\r\n\r\n* feat(observability-lib): add dashboard with panels + test for capabilities\r\n\r\n* minor fix to metric and panel naming\r\n\r\n* chore(observability-lib): update capabilities test\r\n\r\n---------\r\n\r\nCo-authored-by: Clement Erena <clement.erena@smartcontract.com>",
          "timestamp": "2024-09-14T20:28:31-04:00",
          "tree_id": "7cf207cf03b1ec54a2b1d67ac264efdbccb63819",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e1fc24838e09f88e60ca8b3f7b305562df11761a"
        },
        "date": 1726360179709,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.7,
            "unit": "ns/op",
            "extra": "2682027 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 498.7,
            "unit": "ns/op",
            "extra": "2346651 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28399,
            "unit": "ns/op",
            "extra": "42294 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "96e50c64ed1173e3a0efb3c2c9a45225dc8b38f7",
          "message": "[CAPPL-51] Remove error from Runner.Run (#764)",
          "timestamp": "2024-09-16T12:20:56+01:00",
          "tree_id": "5996846c0d14de4f75c857ba74d72799a1c8125b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/96e50c64ed1173e3a0efb3c2c9a45225dc8b38f7"
        },
        "date": 1726485717012,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 478.7,
            "unit": "ns/op",
            "extra": "2611305 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 502.2,
            "unit": "ns/op",
            "extra": "2384001 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28434,
            "unit": "ns/op",
            "extra": "41090 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "36cb47701edf36bcd8de8e32f0ade21f1e534971",
          "message": "Seperate sdk from workflow to as part of an effort to shrink the WASM binary size (#765)",
          "timestamp": "2024-09-16T11:03:42-04:00",
          "tree_id": "dc3e9c455fded04167dfcbe91152e25ec89ed5f6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/36cb47701edf36bcd8de8e32f0ade21f1e534971"
        },
        "date": 1726499088728,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.8,
            "unit": "ns/op",
            "extra": "2595808 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.3,
            "unit": "ns/op",
            "extra": "2410183 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28386,
            "unit": "ns/op",
            "extra": "42262 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1702865+kidambisrinivas@users.noreply.github.com",
            "name": "Sri Kidambi",
            "username": "kidambisrinivas"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b8f8ccc4ecb3f0f9ccebdafcddd3128ba680b076",
          "message": "Add LatestHead relayer method to gRPC client and server implementations (#767)\n\n* Add LatestHead method to relayer interface\r\n\r\n* Add LatestHead implementation to relayerset server and clients\r\n\r\n* Fix error message",
          "timestamp": "2024-09-16T17:33:59+01:00",
          "tree_id": "a19bcb56fd532def16905f3b42fcf5f6f8b97c59",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b8f8ccc4ecb3f0f9ccebdafcddd3128ba680b076"
        },
        "date": 1726504498645,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457.4,
            "unit": "ns/op",
            "extra": "2606354 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 506.4,
            "unit": "ns/op",
            "extra": "2337867 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28396,
            "unit": "ns/op",
            "extra": "42072 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "47eac983684dde08adce4032eec5afe4280cdfb8",
          "message": "[CAPPL-38] GetWorkflowSpec tidy up (#770)\n\n* [CAPPL-51] Remove error from Runner.Run (#764)\r\n\r\n* [CAPPL-38] Add more tests for GetWorkflowSpec",
          "timestamp": "2024-09-17T10:00:32+01:00",
          "tree_id": "874f726aaedd58445e0d0f5deb23369c9f82d1f6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/47eac983684dde08adce4032eec5afe4280cdfb8"
        },
        "date": 1726563698421,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.9,
            "unit": "ns/op",
            "extra": "2236711 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 527.1,
            "unit": "ns/op",
            "extra": "2353958 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28252,
            "unit": "ns/op",
            "extra": "42452 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d6de1326f383c32be93350fce94b54e251dcb397",
          "message": "Allow remote refs in the generator (#769)",
          "timestamp": "2024-09-17T11:59:53-04:00",
          "tree_id": "36019551a1c6f2800328b2bd437ebe6237ee84fe",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d6de1326f383c32be93350fce94b54e251dcb397"
        },
        "date": 1726588894537,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.5,
            "unit": "ns/op",
            "extra": "2637542 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507.9,
            "unit": "ns/op",
            "extra": "2384773 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28309,
            "unit": "ns/op",
            "extra": "42340 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "fc3e154217dde977e93a7264a78108236ac86927",
          "message": "Remove mustEmbed from ContractReader (#772)\n\nTemporarily removing mustEmbed from `ContractReader` interface due to complications\r\ninvolving mockery. To use the mustEmbed, a new interface mocking style must be implemented\r\nthat doesn't directly reference `ContractReader` in common. This would negatively impact\r\nthe timeline to CCIP at the current time.",
          "timestamp": "2024-09-17T11:44:47-05:00",
          "tree_id": "9c302e90e65919fca4a164cc3ab49768a32a602e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/fc3e154217dde977e93a7264a78108236ac86927"
        },
        "date": 1726591544092,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.3,
            "unit": "ns/op",
            "extra": "2427393 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 498.5,
            "unit": "ns/op",
            "extra": "2419476 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28541,
            "unit": "ns/op",
            "extra": "41846 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e78a0de3f6847845b18da1797e5dccf9521abed2",
          "message": "[CAPPL-58] Some cleanup; add sandboxing tests (#773)",
          "timestamp": "2024-09-18T11:32:07+01:00",
          "tree_id": "7c8956a2cc2d0d463bf0b52d75a642a44c48a540",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e78a0de3f6847845b18da1797e5dccf9521abed2"
        },
        "date": 1726655590381,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 472.8,
            "unit": "ns/op",
            "extra": "2553610 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 496.6,
            "unit": "ns/op",
            "extra": "2399362 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29791,
            "unit": "ns/op",
            "extra": "42066 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "57732589+ilija42@users.noreply.github.com",
            "name": "ilija42",
            "username": "ilija42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "564164004d06594f1ae66adc9962fe8d82dbe442",
          "message": "[BCFR-203] - Improve CR ValComparators to take in arbitrary value (#689)\n\n* Improve CR ValComparators to use any instead of string\r\n\r\n* Add interface tests for QueryKey value comparators\r\n\r\n* Remove CR querying for nested fields test case\r\n\r\n* Rearrange CR TestStruct fields for easier EVM testing\r\n\r\n* lint\r\n\r\n* Update QueryKey Val Comp test case to use Chain Writer\r\n\r\n* Minor fixes and lint\r\n\r\n* FIx testing helper ComparisonOperator Compare function\r\n\r\n* run generate\r\n\r\n* Fix QueryKey filter conversion from proto",
          "timestamp": "2024-09-18T23:05:34+02:00",
          "tree_id": "537964cd27966dbb221077bda09bab708cfb81e9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/564164004d06594f1ae66adc9962fe8d82dbe442"
        },
        "date": 1726693590401,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.4,
            "unit": "ns/op",
            "extra": "2621308 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 504.3,
            "unit": "ns/op",
            "extra": "2369784 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28544,
            "unit": "ns/op",
            "extra": "41917 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "dimitrios.kouveris@smartcontract.com",
            "name": "dimitris",
            "username": "dimkouv"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "53e784c2e420f838f9b8925ae81809e7fa8003ba",
          "message": "add limit to seq num range (#781)",
          "timestamp": "2024-09-19T12:24:17+03:00",
          "tree_id": "ddf8392d15fe9f7ea428cc734ae337d5db603dfa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/53e784c2e420f838f9b8925ae81809e7fa8003ba"
        },
        "date": 1726737912792,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.7,
            "unit": "ns/op",
            "extra": "2508822 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 501.7,
            "unit": "ns/op",
            "extra": "2396382 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28564,
            "unit": "ns/op",
            "extra": "42033 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "34e8551279c436bd44b91e11a531f06f475e7101",
          "message": "[chore] Handle aliases in slices (#784)\n\n* [chore] Handle aliases in slices\r\n\r\n* More aliasing tests\r\n\r\n* Lint fix\r\n\r\n* Fix test\r\n\r\n---------\r\n\r\nCo-authored-by: Sri Kidambi <1702865+kidambisrinivas@users.noreply.github.com>",
          "timestamp": "2024-09-20T11:49:03+01:00",
          "tree_id": "b54f606437e8439cb2afd8e3c72e9db079a90f7e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/34e8551279c436bd44b91e11a531f06f475e7101"
        },
        "date": 1726829402160,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.9,
            "unit": "ns/op",
            "extra": "2650323 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 499.9,
            "unit": "ns/op",
            "extra": "2392999 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28584,
            "unit": "ns/op",
            "extra": "42037 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "clement.erena78@gmail.com",
            "name": "Clement",
            "username": "Atrax1"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8f5c155769f5c4054d9ac35632b1d2e076ee607b",
          "message": "feat(observability-lib): legendoptions + improvement on node general dashboard (#785)",
          "timestamp": "2024-09-23T12:10:18+02:00",
          "tree_id": "0b001896130f602d9f9168aad1b053e1e8fe61aa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8f5c155769f5c4054d9ac35632b1d2e076ee607b"
        },
        "date": 1727086283637,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.3,
            "unit": "ns/op",
            "extra": "2626467 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.4,
            "unit": "ns/op",
            "extra": "2403264 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28613,
            "unit": "ns/op",
            "extra": "41968 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "61c2ccba2f58d4124aae064bbb637f8af403ab13",
          "message": "[CAPPL-58] Correctly stub out clock_time_get and poll_oneoff (#778)\n\n* [CAPPL-58] Further cleanup\r\n\r\n* [CAPPL-58] Add support for compression",
          "timestamp": "2024-09-23T12:30:02+01:00",
          "tree_id": "df00fa350dba5581f39c784494c813480af963e7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/61c2ccba2f58d4124aae064bbb637f8af403ab13"
        },
        "date": 1727091103085,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.2,
            "unit": "ns/op",
            "extra": "2634062 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507,
            "unit": "ns/op",
            "extra": "2371747 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28790,
            "unit": "ns/op",
            "extra": "42038 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1702865+kidambisrinivas@users.noreply.github.com",
            "name": "Sri Kidambi",
            "username": "kidambisrinivas"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5d125850fa8092d6d00a90ec06b020cd2c4890c7",
          "message": "More alias handling in Unwrap functionality of Value  (#792)\n\n* Generic case to handle both pointer type and raw type and simplify int unwrap\r\n\r\n* Handling interface and default\r\n\r\n* Small test fix\r\n\r\n---------\r\n\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>",
          "timestamp": "2024-09-23T12:50:45+01:00",
          "tree_id": "cd5819f2d68d7491943bd77d1402c6da3dbdd05a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5d125850fa8092d6d00a90ec06b020cd2c4890c7"
        },
        "date": 1727092304952,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 471.6,
            "unit": "ns/op",
            "extra": "2600544 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 500.2,
            "unit": "ns/op",
            "extra": "2385236 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28554,
            "unit": "ns/op",
            "extra": "42028 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "26df9abc1e1a7d70afc2e010258991c5d86aae45",
          "message": "Fix alias typing and tests (#788)\n\n* Fix alias typing and tests\r\n\r\n* Fix ints\r\n\r\n* errors.new instead of fmt\r\n\r\n* Add array support to slice (#789)",
          "timestamp": "2024-09-23T15:26:41+01:00",
          "tree_id": "6493dc5dd53b99b65bfcdd2361ae709965543c81",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/26df9abc1e1a7d70afc2e010258991c5d86aae45"
        },
        "date": 1727101659842,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.9,
            "unit": "ns/op",
            "extra": "2651382 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 500.6,
            "unit": "ns/op",
            "extra": "2399055 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28581,
            "unit": "ns/op",
            "extra": "41979 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d4cf7aff85cb3f2a0630bc2a11b42cfcbb80c40c",
          "message": "Replace fmt.Errorf with errors.New where possible (#795)",
          "timestamp": "2024-09-23T10:43:42-04:00",
          "tree_id": "9b94ee205da2fa59988cc1834dc27df159edcd9b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d4cf7aff85cb3f2a0630bc2a11b42cfcbb80c40c"
        },
        "date": 1727102682635,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.5,
            "unit": "ns/op",
            "extra": "2671917 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 500.2,
            "unit": "ns/op",
            "extra": "2392772 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29936,
            "unit": "ns/op",
            "extra": "41856 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "5597260+MStreet3@users.noreply.github.com",
            "name": "Street",
            "username": "MStreet3"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c01f105fa51a16b0e0777556ee7446063aa6c7c3",
          "message": "chore(workflows): adds unit test to utils (#782)",
          "timestamp": "2024-09-23T11:36:27-04:00",
          "tree_id": "88bd993eaf29936a7d233bec65f1e3fecd3bb8ac",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c01f105fa51a16b0e0777556ee7446063aa6c7c3"
        },
        "date": 1727105851914,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.3,
            "unit": "ns/op",
            "extra": "2628511 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 502.2,
            "unit": "ns/op",
            "extra": "2377428 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28552,
            "unit": "ns/op",
            "extra": "41804 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9f7db4ed0dfc383ecb602473f75a3d869bf0b7b2",
          "message": "Have the mock runner register with capabilites (#783)",
          "timestamp": "2024-09-23T12:37:38-04:00",
          "tree_id": "a61538be47fa0de09c75b128810b5581444536f6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9f7db4ed0dfc383ecb602473f75a3d869bf0b7b2"
        },
        "date": 1727109514824,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 470.4,
            "unit": "ns/op",
            "extra": "2622174 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.9,
            "unit": "ns/op",
            "extra": "2339164 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28427,
            "unit": "ns/op",
            "extra": "42258 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "14086514727b667dc37edd4968b46c5a49222080",
          "message": "Add binary + config to custom compute (#794)\n\n* Add binary + config to custom compute\r\n\r\n* Add binary + config to custom compute",
          "timestamp": "2024-09-23T18:56:42+01:00",
          "tree_id": "7900823c6362810e822119da3b8d59dc556a265d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/14086514727b667dc37edd4968b46c5a49222080"
        },
        "date": 1727114265784,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.3,
            "unit": "ns/op",
            "extra": "2455132 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.9,
            "unit": "ns/op",
            "extra": "2373352 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28551,
            "unit": "ns/op",
            "extra": "41997 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2cc8993e6fe8809cee06394669bcebaff4a13c89",
          "message": "fix lint issues (#786)",
          "timestamp": "2024-09-24T09:39:09-05:00",
          "tree_id": "c313a2c5e79de90878af1fe4546562309827dfa0",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2cc8993e6fe8809cee06394669bcebaff4a13c89"
        },
        "date": 1727188805978,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.3,
            "unit": "ns/op",
            "extra": "2620765 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.5,
            "unit": "ns/op",
            "extra": "2380339 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28532,
            "unit": "ns/op",
            "extra": "41853 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "96611a2a09bde9a84d464611cb946cbd0d1c6160",
          "message": "execution factory constructor updated to take two providers, chainIDs, and source token address (#641)\n\n* execution factory constructor updated to take two providers and chain IDs\r\n\r\n(cherry picked from commit 6ad1f08d26810df5eaeed76a0f74e20be1908658)\r\n\r\n* Adding source token address to execution factory constructor",
          "timestamp": "2024-09-24T19:39:14-04:00",
          "tree_id": "0b600fd335943a0a5766a7f0909ae87dd9161d27",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/96611a2a09bde9a84d464611cb946cbd0d1c6160"
        },
        "date": 1727221225930,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465.1,
            "unit": "ns/op",
            "extra": "2575639 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 550.4,
            "unit": "ns/op",
            "extra": "2138781 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28808,
            "unit": "ns/op",
            "extra": "41583 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "aded1b263ecc96050483a55502793d3fe666783d",
          "message": "Support passing in a values.Value to the chainreader GetLatestValue method (#779)\n\n* add support for passing in a values.Value type to the contract readers GetLatestValue and QueryKey methods\r\n\r\n---------\r\n\r\nCo-authored-by: Sri Kidambi <1702865+kidambisrinivas@users.noreply.github.com>\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>",
          "timestamp": "2024-09-25T09:52:18+01:00",
          "tree_id": "e186a04c051e452541ca7e07dc7c816da2f2d003",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aded1b263ecc96050483a55502793d3fe666783d"
        },
        "date": 1727254397320,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 477,
            "unit": "ns/op",
            "extra": "2485322 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 546.7,
            "unit": "ns/op",
            "extra": "2127074 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28477,
            "unit": "ns/op",
            "extra": "42250 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "5597260+MStreet3@users.noreply.github.com",
            "name": "Street",
            "username": "MStreet3"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "10282bf15d4a0db9f29cca6268a1941021bfad39",
          "message": "feat(values): support float64 values (#799)",
          "timestamp": "2024-09-25T15:32:34-04:00",
          "tree_id": "883347d4615ff5fd188125f134dfd8ae675abcfc",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/10282bf15d4a0db9f29cca6268a1941021bfad39"
        },
        "date": 1727292827897,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 474.2,
            "unit": "ns/op",
            "extra": "2334978 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.3,
            "unit": "ns/op",
            "extra": "2299010 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28428,
            "unit": "ns/op",
            "extra": "41890 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c7122148f846cb5fd5b91dbf8e7b592dccebdc95",
          "message": "confidence level from string (#802)",
          "timestamp": "2024-09-26T09:58:36+01:00",
          "tree_id": "59c824df07e9499feeda5a0cbb67bde008df3c09",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c7122148f846cb5fd5b91dbf8e7b592dccebdc95"
        },
        "date": 1727341173240,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.9,
            "unit": "ns/op",
            "extra": "2478061 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 541.5,
            "unit": "ns/op",
            "extra": "2207108 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28569,
            "unit": "ns/op",
            "extra": "42012 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "192f940806bdc3c03a5c6cabe04bca875cc0f0c1",
          "message": "Float32/Float64 wrapping (#804)",
          "timestamp": "2024-09-26T12:04:54+01:00",
          "tree_id": "a981ebbefa89dcc88d0b71da458573d7ee6b00af",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/192f940806bdc3c03a5c6cabe04bca875cc0f0c1"
        },
        "date": 1727348757095,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.9,
            "unit": "ns/op",
            "extra": "2535540 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 543.8,
            "unit": "ns/op",
            "extra": "2176690 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28588,
            "unit": "ns/op",
            "extra": "42008 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "gaboparadiso@gmail.com",
            "name": "Gabriel Paradiso",
            "username": "agparadiso"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "eb2be16689074105a4b540a246c70394cfc41303",
          "message": "feat: implement sdk logger (#762)",
          "timestamp": "2024-09-26T14:11:50+02:00",
          "tree_id": "cf6fe0b2b684aed9c2520ff35159865762138575",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eb2be16689074105a4b540a246c70394cfc41303"
        },
        "date": 1727352777354,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.6,
            "unit": "ns/op",
            "extra": "2434596 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 527.7,
            "unit": "ns/op",
            "extra": "2147551 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28575,
            "unit": "ns/op",
            "extra": "42057 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7a9a88aee28f9948d7c9fcc74a615f9b5552ca88",
          "message": "Add MustEmbed Constraint to Contract Reader (#801)\n\nReintroducing the must embed constraint to `ContractReader` implementations to\r\nensure that all implementations of `ContractReader` embed the `UnimplementedContractReader`.\r\n\r\nIf an implementation contains the unemplemented struct, changes to the interface\r\nwill flow down to all implementations without introducing breaking changes.",
          "timestamp": "2024-09-26T09:46:09-05:00",
          "tree_id": "dc50c9b45a8a9b6153d7717b11a205137334198c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7a9a88aee28f9948d7c9fcc74a615f9b5552ca88"
        },
        "date": 1727362038865,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.4,
            "unit": "ns/op",
            "extra": "2404717 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526.2,
            "unit": "ns/op",
            "extra": "2212948 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28894,
            "unit": "ns/op",
            "extra": "40987 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1416262+bolekk@users.noreply.github.com",
            "name": "Bolek",
            "username": "bolekk"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "84ed150bf0bc4353cfd246518113c9a8e800d0bf",
          "message": "[CAPPL-60] Dynamic encoder selection in OCR consensus aggregator (#780)\n\nCo-authored-by: Cedric <cedric.cordenier@smartcontract.com>",
          "timestamp": "2024-09-26T16:32:11+01:00",
          "tree_id": "5ec512ba9b28c0a8ab3805dbed5646f25abb867d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/84ed150bf0bc4353cfd246518113c9a8e800d0bf"
        },
        "date": 1727364789402,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465.3,
            "unit": "ns/op",
            "extra": "2442145 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.7,
            "unit": "ns/op",
            "extra": "2307429 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28553,
            "unit": "ns/op",
            "extra": "41988 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "32529249+silaslenihan@users.noreply.github.com",
            "name": "Silas Lenihan",
            "username": "silaslenihan"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "0784a13b25368ec80ba621094477102839f2069b",
          "message": "Updated TestStruct to enable advanced querying (#798)\n\n* Updated TestStruct to enable advanced querying\r\n\r\n* linting fixes\r\n\r\n* Update pkg/codec/encodings/type_codec_test.go\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\n\r\n* Update pkg/codec/encodings/type_codec_test.go\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\n\r\n* Fixed codec tests\r\n\r\n---------\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>",
          "timestamp": "2024-09-26T14:01:10-04:00",
          "tree_id": "bba4411bd757b4ffb49ec6da1c30b9f91eb230e3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0784a13b25368ec80ba621094477102839f2069b"
        },
        "date": 1727373732892,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.5,
            "unit": "ns/op",
            "extra": "2445597 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.5,
            "unit": "ns/op",
            "extra": "2269346 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28603,
            "unit": "ns/op",
            "extra": "42001 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "20630b333f5737d23e1bad362900a27c1a22d680",
          "message": "Properly support the range of uint64 and allow big int to unwrap into smaller integer types (#810)",
          "timestamp": "2024-09-27T12:24:47-04:00",
          "tree_id": "8a4c28b8e67841211bd4c228749dd08d00af8543",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/20630b333f5737d23e1bad362900a27c1a22d680"
        },
        "date": 1727454349963,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.4,
            "unit": "ns/op",
            "extra": "2235145 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 531.6,
            "unit": "ns/op",
            "extra": "2184152 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28673,
            "unit": "ns/op",
            "extra": "41961 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4cca03442d82ebae2396f31114ed6038c512f60e",
          "message": "Extract expirable cache abstraction for reuse (#807)\n\n* expirable_cache",
          "timestamp": "2024-09-30T10:25:29+01:00",
          "tree_id": "a847c4f39c45d97b3ce78432e11f7007372f0f75",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4cca03442d82ebae2396f31114ed6038c512f60e"
        },
        "date": 1727688391767,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467.2,
            "unit": "ns/op",
            "extra": "2493256 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526.2,
            "unit": "ns/op",
            "extra": "2192206 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28598,
            "unit": "ns/op",
            "extra": "41940 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "33d83298df3784cd00735d0e4338d7ff84d859f7",
          "message": "remove cache (#812)",
          "timestamp": "2024-09-30T12:18:21+01:00",
          "tree_id": "8a4c28b8e67841211bd4c228749dd08d00af8543",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/33d83298df3784cd00735d0e4338d7ff84d859f7"
        },
        "date": 1727695160443,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461,
            "unit": "ns/op",
            "extra": "2516266 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.3,
            "unit": "ns/op",
            "extra": "2265940 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29130,
            "unit": "ns/op",
            "extra": "42007 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "mateusz.sekara@gmail.com",
            "name": "Mateusz Sekara",
            "username": "mateusz-sekara"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ef04dd443670b7e892611bfd4484b06a6217e12b",
          "message": "CCIP-3555 Attestation encoder interfaces (#813)\n\n* Attestation encoder interfaces\r\n\r\n* Attestation encoder interfaces\r\n\r\n* Attestation encoder interfaces\r\n\r\n* Comment",
          "timestamp": "2024-09-30T18:21:17+04:00",
          "tree_id": "c28119fd33c9048e596d21a3cf55364f22ea7eaa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ef04dd443670b7e892611bfd4484b06a6217e12b"
        },
        "date": 1727706144960,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 462.5,
            "unit": "ns/op",
            "extra": "2397535 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.3,
            "unit": "ns/op",
            "extra": "2288337 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28449,
            "unit": "ns/op",
            "extra": "42183 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "57732589+ilija42@users.noreply.github.com",
            "name": "ilija42",
            "username": "ilija42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f10ba2b23682b36574339a2738438142a40644f8",
          "message": "[BCF-3392]  - ContractReaderByIDs Wrapper (#797)\n\n* WIP\r\n\r\n* Update ContractReaderByIDs interface method names\r\n\r\n* Unexpose types.ContractReader in contractReaderByIDs\r\n\r\n* Add multiple contract address support to fakeContractReader for tests\r\n\r\n* Add GetLatestValue unit test for contractReaderByIDs\r\n\r\n* Add GetLatestValue unit test for QueryKey\r\n\r\n* Add BatchGetLatestValues unit test for CR by custom IdDs wrapper\r\n\r\n* Rm ContractReaderByIDs interface and export the struct\r\n\r\n* Change ContractReaderByIDs wrapper Unbind handling\r\n\r\n* Improve ContractReaderByIDs wrapper err handling\r\n\r\n* Remove mockery usage from ContractReaderByIDs tests\r\n\r\n* lint",
          "timestamp": "2024-09-30T19:03:30+02:00",
          "tree_id": "9ebce821913dac0321bcae74166c84fa9d93848d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f10ba2b23682b36574339a2738438142a40644f8"
        },
        "date": 1727715877348,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.2,
            "unit": "ns/op",
            "extra": "2399402 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 556.6,
            "unit": "ns/op",
            "extra": "2235771 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28469,
            "unit": "ns/op",
            "extra": "42158 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "makramkd@users.noreply.github.com",
            "name": "Makram",
            "username": "makramkd"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ac3da2ed53850e3b2570713b91a3ffb6a7198f0a",
          "message": "pkg/types/ccipocr3: add DestExecData to RampTokenAmount (#817)\n\n* pkg/types/ccipocr3: add DestExecData to RampTokenAmount\n\n* fix test",
          "timestamp": "2024-10-01T07:48:01-05:00",
          "tree_id": "27a31d3e9d3b644f8e1eb1834704e6ed86afecf9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ac3da2ed53850e3b2570713b91a3ffb6a7198f0a"
        },
        "date": 1727786950513,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.5,
            "unit": "ns/op",
            "extra": "2271177 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 518,
            "unit": "ns/op",
            "extra": "2286686 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28429,
            "unit": "ns/op",
            "extra": "42188 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "35be2fad06ec94d06d27607ebf8a8b37be136813",
          "message": "Allow the creation of maps from string to capbility outputs. (#815)",
          "timestamp": "2024-10-01T10:04:26-04:00",
          "tree_id": "b3da1cb79044d549c0be6da91a5b90ea07e01bb3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/35be2fad06ec94d06d27607ebf8a8b37be136813"
        },
        "date": 1727791525825,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467,
            "unit": "ns/op",
            "extra": "2531828 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523,
            "unit": "ns/op",
            "extra": "2293088 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28556,
            "unit": "ns/op",
            "extra": "42242 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "rstout610@gmail.com",
            "name": "Ryan Stout",
            "username": "rstout"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "dd59341432bd814166d98be1fb6bccfb0da6cbb5",
          "message": "Add the FeeValueJuels field to ccipocr3.Message (#819)",
          "timestamp": "2024-10-01T16:00:38-05:00",
          "tree_id": "5d39f28eccecebde9aecd9ad394cb6fb554fc3ec",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/dd59341432bd814166d98be1fb6bccfb0da6cbb5"
        },
        "date": 1727816503117,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 541.3,
            "unit": "ns/op",
            "extra": "2420666 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 607.1,
            "unit": "ns/op",
            "extra": "2087967 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 30450,
            "unit": "ns/op",
            "extra": "39322 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "clement.erena78@gmail.com",
            "name": "Clement",
            "username": "Atrax1"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e1435d915916aef78e985938f9cc98fe501507f2",
          "message": "feat(observability-lib): improve alerts rule (#803)\n\n* feat(observability-lib): improve alerts rule\r\n\r\n* chore(observability-lib): README + folder structure (#806)\r\n\r\n* chore(observability-lib): README + folder structure\r\n\r\n* feat(observability-lib): variable add current + includeAll options (#808)\r\n\r\n* chore(README): small corrections\r\n\r\n* chore(README): example improved\r\n\r\n* chore(README): add references to dashboards examples\r\n\r\n* feat(observability-lib): refactor exportable func + link to godoc\r\n\r\n* fix(observability-lib): cmd errors returns",
          "timestamp": "2024-10-02T11:49:39+02:00",
          "tree_id": "b60da4080a71c581a13562d80fcc3296af3b45e2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e1435d915916aef78e985938f9cc98fe501507f2"
        },
        "date": 1727862637989,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.4,
            "unit": "ns/op",
            "extra": "2522540 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.1,
            "unit": "ns/op",
            "extra": "2311569 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28581,
            "unit": "ns/op",
            "extra": "42048 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "943f18920813a1a2deacedc190dd057e59a5ea3f",
          "message": "enable errorf check (#826)",
          "timestamp": "2024-10-02T09:19:43-05:00",
          "tree_id": "0c725be0d592fa83206c28cb8c9e34662db7dc7f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/943f18920813a1a2deacedc190dd057e59a5ea3f"
        },
        "date": 1727878842965,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456,
            "unit": "ns/op",
            "extra": "2452474 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.1,
            "unit": "ns/op",
            "extra": "2313592 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28593,
            "unit": "ns/op",
            "extra": "41998 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "707d968168d4c1bfdc43264eed41da87f64d0e53",
          "message": "Fix map and ToListDefinition, add tests for them in the builder, add a way to create a list of any from inputs (#823)\n\n* Fix map and ToListDefinition, add tests for them in the builder, add a way to create a list of any from inputs\r\n\r\n* Fix any map test\r\n\r\n* Clarify comment int singleCapList Index",
          "timestamp": "2024-10-02T22:14:18+01:00",
          "tree_id": "a6f931ac792dbffb8b36cacc002b56e485f79961",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/707d968168d4c1bfdc43264eed41da87f64d0e53"
        },
        "date": 1727903717101,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 474.9,
            "unit": "ns/op",
            "extra": "2357803 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521,
            "unit": "ns/op",
            "extra": "2292445 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28437,
            "unit": "ns/op",
            "extra": "42206 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "alecpgard@gmail.com",
            "name": "Alec Gard",
            "username": "alecgard"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "298546dff699f20d9be574df38011759f6658316",
          "message": "Add metric offchain_aggregator_answers_latest_timestamp (#825)",
          "timestamp": "2024-10-04T05:57:06-05:00",
          "tree_id": "12f3f0b84837f3cf023fbbaff38ad26f05e6659d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/298546dff699f20d9be574df38011759f6658316"
        },
        "date": 1728039531164,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 452.8,
            "unit": "ns/op",
            "extra": "2663928 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 533.2,
            "unit": "ns/op",
            "extra": "2327762 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28521,
            "unit": "ns/op",
            "extra": "42256 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "93c2fb862aa950f2349664a817611c9ede779d1b",
          "message": "Fix a bug where schema validation looses type information if the input has an any in it (#827)",
          "timestamp": "2024-10-07T11:06:59-04:00",
          "tree_id": "5726c055322acc765a8c236ac3bf420cac0cd7bb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/93c2fb862aa950f2349664a817611c9ede779d1b"
        },
        "date": 1728313683792,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.8,
            "unit": "ns/op",
            "extra": "2639769 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.7,
            "unit": "ns/op",
            "extra": "2338688 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28545,
            "unit": "ns/op",
            "extra": "41959 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "justinkaseman@live.com",
            "name": "Justin Kaseman",
            "username": "justinkaseman"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d98e32024f8b8b7403a036a5b7704fdf99a7eb37",
          "message": "Add Capabilities team code owners to values library and /capabilities (#820)",
          "timestamp": "2024-10-07T19:04:13-04:00",
          "tree_id": "d16dc3006b5c1c1948aff8bf3afd578a3f647d47",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d98e32024f8b8b7403a036a5b7704fdf99a7eb37"
        },
        "date": 1728342312055,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.2,
            "unit": "ns/op",
            "extra": "2697048 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.1,
            "unit": "ns/op",
            "extra": "2335945 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28499,
            "unit": "ns/op",
            "extra": "42115 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "5597260+MStreet3@users.noreply.github.com",
            "name": "Street",
            "username": "MStreet3"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c03afeeb7d6df77ef669b1c6d81fdb0e1919ee85",
          "message": "feat(wasm): override random_get (#831)\n\n* feat(wasm): override random_get\r\n\r\n* chore(wasm): adds a deterministic config to modules\r\n\r\n* feat(wasm): require DAG creation is deterministic\r\n\r\n* chore(host): move random_get override",
          "timestamp": "2024-10-08T12:05:44-04:00",
          "tree_id": "3335cab21ac9bf395101016e752bfaf8800643d4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c03afeeb7d6df77ef669b1c6d81fdb0e1919ee85"
        },
        "date": 1728403601555,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.3,
            "unit": "ns/op",
            "extra": "2710537 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.6,
            "unit": "ns/op",
            "extra": "2354833 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28434,
            "unit": "ns/op",
            "extra": "42225 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "deividas.karzinauskas@gmail.com",
            "name": "Deividas Karžinauskas",
            "username": "DeividasK"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8bfcea33a98dce20a28f5205407eff129a04915c",
          "message": "[KS-430] Provide an OracleFactory to StandardCapabilities (#738)",
          "timestamp": "2024-10-08T20:04:07+03:00",
          "tree_id": "81ab6f90fec133ce592aa77bcbacc95f9150b4a8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8bfcea33a98dce20a28f5205407eff129a04915c"
        },
        "date": 1728407149690,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.2,
            "unit": "ns/op",
            "extra": "2652418 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.9,
            "unit": "ns/op",
            "extra": "2339522 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29354,
            "unit": "ns/op",
            "extra": "42265 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "gaboparadiso@gmail.com",
            "name": "Gabriel Paradiso",
            "username": "agparadiso"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a3ff1166f0dd80b998eb792368803e1ed4bdf900",
          "message": "[CAPPL-41] SDK Fetch import (#814)\n\n* feat: draft implementation of fetch\r\n\r\n* feat: use proto for the guest <> host communication\r\n\r\n* chore: nop implementation by default\r\n\r\n* chore: adjust errors returned\r\n\r\n* Pass responseSizeBytes via Compute call\r\n\r\n* Handle errors\r\n\r\n* fix: expose fetch and logger functions on runtime sdk\r\n\r\n* test: add test coverage for err while fetching and runtime cfg\r\n\r\n* test: validate response instead of log\r\n\r\n* chore: address comments\r\n\r\n---------\r\n\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>",
          "timestamp": "2024-10-08T18:35:18+01:00",
          "tree_id": "08008ce6df355f48f32297dd4cbd8ac66c340abb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a3ff1166f0dd80b998eb792368803e1ed4bdf900"
        },
        "date": 1728408976767,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.4,
            "unit": "ns/op",
            "extra": "2635040 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.5,
            "unit": "ns/op",
            "extra": "2339486 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28247,
            "unit": "ns/op",
            "extra": "42456 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "2677789+asoliman92@users.noreply.github.com",
            "name": "Abdelrahman Soliman (Boda)",
            "username": "asoliman92"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "167715aa8613ebdaa17f1b5b4b003a31c120d455",
          "message": "Mirror on-chain data structures (#833)\n\n* Mirror on-chain data structures\r\n\r\n* Revert \"Mirror on-chain data structures\"\r\n\r\nThis reverts commit b647b125f34e1f3ae6b64c17af4ac9d6acbb132b.\r\n\r\n* Mirror on-chain data structures\r\n\r\n* address comments\r\n\r\n---------\r\n\r\nCo-authored-by: Makram Kamaleddine <makramkd@users.noreply.github.com>",
          "timestamp": "2024-10-08T21:52:10+04:00",
          "tree_id": "c395aa1fbc5de85e36ebcc8737fabec4f88ffb30",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/167715aa8613ebdaa17f1b5b4b003a31c120d455"
        },
        "date": 1728409987101,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.2,
            "unit": "ns/op",
            "extra": "2572830 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.2,
            "unit": "ns/op",
            "extra": "2335609 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28417,
            "unit": "ns/op",
            "extra": "42259 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8166e6555b2330ba0abfc747a6a78d0685394d89",
          "message": "bump libocr; add context (#490)",
          "timestamp": "2024-10-09T07:48:09-05:00",
          "tree_id": "0e5686161e9fd777ff388b8a9bcba9d61a858ac3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8166e6555b2330ba0abfc747a6a78d0685394d89"
        },
        "date": 1728478188814,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.1,
            "unit": "ns/op",
            "extra": "2634718 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.5,
            "unit": "ns/op",
            "extra": "2278089 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28250,
            "unit": "ns/op",
            "extra": "42421 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "1416262+bolekk@users.noreply.github.com",
            "name": "Bolek",
            "username": "bolekk"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "46370848a78919a4974f33871de164a93c49e772",
          "message": "[CM-380] Identical Aggregator (#771)\n\n* [CM-380] Identical Aggregator\r\n\r\n* [CAPPL-60] Dynamic encoder selection in OCR consensus aggregator\r\n\r\n* extract encoder name and config\r\n\r\n* Add more tests\r\n\r\n* add limit to seq num range (#781)\r\n\r\n* [chore] Handle aliases in slices (#784)\r\n\r\n* [chore] Handle aliases in slices\r\n\r\n* More aliasing tests\r\n\r\n* Lint fix\r\n\r\n* Fix test\r\n\r\n---------\r\n\r\nCo-authored-by: Sri Kidambi <1702865+kidambisrinivas@users.noreply.github.com>\r\n\r\n* feat(observability-lib): legendoptions + improvement on node general dashboard (#785)\r\n\r\n* [CAPPL-58] Correctly stub out clock_time_get and poll_oneoff (#778)\r\n\r\n* [CAPPL-58] Further cleanup\r\n\r\n* [CAPPL-58] Add support for compression\r\n\r\n* More alias handling in Unwrap functionality of Value  (#792)\r\n\r\n* Generic case to handle both pointer type and raw type and simplify int unwrap\r\n\r\n* Handling interface and default\r\n\r\n* Small test fix\r\n\r\n---------\r\n\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>\r\n\r\n* Fix alias typing and tests (#788)\r\n\r\n* Fix alias typing and tests\r\n\r\n* Fix ints\r\n\r\n* errors.new instead of fmt\r\n\r\n* Add array support to slice (#789)\r\n\r\n* Replace fmt.Errorf with errors.New where possible (#795)\r\n\r\n* chore(workflows): adds unit test to utils (#782)\r\n\r\n* Have the mock runner register with capabilites (#783)\r\n\r\n* Add binary + config to custom compute (#794)\r\n\r\n* Add binary + config to custom compute\r\n\r\n* Add binary + config to custom compute\r\n\r\n* fix lint issues (#786)\r\n\r\n* execution factory constructor updated to take two providers, chainIDs, and source token address (#641)\r\n\r\n* execution factory constructor updated to take two providers and chain IDs\r\n\r\n(cherry picked from commit 6ad1f08d26810df5eaeed76a0f74e20be1908658)\r\n\r\n* Adding source token address to execution factory constructor\r\n\r\n* Support passing in a values.Value to the chainreader GetLatestValue method (#779)\r\n\r\n* add support for passing in a values.Value type to the contract readers GetLatestValue and QueryKey methods\r\n\r\n---------\r\n\r\nCo-authored-by: Sri Kidambi <1702865+kidambisrinivas@users.noreply.github.com>\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>\r\n\r\n* [CAPPL-31] feat(values): adds support for time.Time as value (#787)\r\n\r\n* feat(values): adds support for time.Time as value\r\n\r\n* chore(deps): updates .tool-versions\r\n\r\n* refactor(values): uses primitive type in protos\r\n\r\n* feat(values): support float64 values (#799)\r\n\r\n* confidence level from string (#802)\r\n\r\n* Float32/Float64 wrapping (#804)\r\n\r\n* feat: implement sdk logger (#762)\r\n\r\n* Add MustEmbed Constraint to Contract Reader (#801)\r\n\r\nReintroducing the must embed constraint to `ContractReader` implementations to\r\nensure that all implementations of `ContractReader` embed the `UnimplementedContractReader`.\r\n\r\nIf an implementation contains the unemplemented struct, changes to the interface\r\nwill flow down to all implementations without introducing breaking changes.\r\n\r\n* Updated TestStruct to enable advanced querying (#798)\r\n\r\n* Updated TestStruct to enable advanced querying\r\n\r\n* linting fixes\r\n\r\n* Update pkg/codec/encodings/type_codec_test.go\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\n\r\n* Update pkg/codec/encodings/type_codec_test.go\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\n\r\n* Fixed codec tests\r\n\r\n---------\r\n\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\n\r\n* Properly support the range of uint64 and allow big int to unwrap into smaller integer types (#810)\r\n\r\n* Extract expirable cache abstraction for reuse (#807)\r\n\r\n* expirable_cache\r\n\r\n* remove cache (#812)\r\n\r\n* CCIP-3555 Attestation encoder interfaces (#813)\r\n\r\n* Attestation encoder interfaces\r\n\r\n* Attestation encoder interfaces\r\n\r\n* Attestation encoder interfaces\r\n\r\n* Comment\r\n\r\n* [BCF-3392]  - ContractReaderByIDs Wrapper (#797)\r\n\r\n* WIP\r\n\r\n* Update ContractReaderByIDs interface method names\r\n\r\n* Unexpose types.ContractReader in contractReaderByIDs\r\n\r\n* Add multiple contract address support to fakeContractReader for tests\r\n\r\n* Add GetLatestValue unit test for contractReaderByIDs\r\n\r\n* Add GetLatestValue unit test for QueryKey\r\n\r\n* Add BatchGetLatestValues unit test for CR by custom IdDs wrapper\r\n\r\n* Rm ContractReaderByIDs interface and export the struct\r\n\r\n* Change ContractReaderByIDs wrapper Unbind handling\r\n\r\n* Improve ContractReaderByIDs wrapper err handling\r\n\r\n* Remove mockery usage from ContractReaderByIDs tests\r\n\r\n* lint\r\n\r\n* pkg/types/ccipocr3: add DestExecData to RampTokenAmount (#817)\r\n\r\n* pkg/types/ccipocr3: add DestExecData to RampTokenAmount\r\n\r\n* fix test\r\n\r\n* Allow the creation of maps from string to capbility outputs. (#815)\r\n\r\n* Add the FeeValueJuels field to ccipocr3.Message (#819)\r\n\r\n* feat(observability-lib): improve alerts rule (#803)\r\n\r\n* feat(observability-lib): improve alerts rule\r\n\r\n* chore(observability-lib): README + folder structure (#806)\r\n\r\n* chore(observability-lib): README + folder structure\r\n\r\n* feat(observability-lib): variable add current + includeAll options (#808)\r\n\r\n* chore(README): small corrections\r\n\r\n* chore(README): example improved\r\n\r\n* chore(README): add references to dashboards examples\r\n\r\n* feat(observability-lib): refactor exportable func + link to godoc\r\n\r\n* fix(observability-lib): cmd errors returns\r\n\r\n* enable errorf check (#826)\r\n\r\n* Make overridding the encoder first-class\r\n\r\n* Update mocks\r\n\r\n* Mock updates\r\n\r\n* Adjust tests\r\n\r\n* Fix mock\r\n\r\n* Fix mock\r\n\r\n* Update mock\r\n\r\n* Linting\r\n\r\n---------\r\n\r\nCo-authored-by: Cedric Cordenier <cedric.cordenier@smartcontract.com>\r\nCo-authored-by: dimitris <dimitrios.kouveris@smartcontract.com>\r\nCo-authored-by: Sri Kidambi <1702865+kidambisrinivas@users.noreply.github.com>\r\nCo-authored-by: Clement <clement.erena78@gmail.com>\r\nCo-authored-by: Ryan Tinianov <tinianov@live.com>\r\nCo-authored-by: Street <5597260+MStreet3@users.noreply.github.com>\r\nCo-authored-by: Jordan Krage <jmank88@gmail.com>\r\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>\r\nCo-authored-by: Matthew Pendrey <matthew.pendrey@gmail.com>\r\nCo-authored-by: Gabriel Paradiso <gaboparadiso@gmail.com>\r\nCo-authored-by: Awbrey Hughlett <athughlett@gmail.com>\r\nCo-authored-by: Silas Lenihan <32529249+silaslenihan@users.noreply.github.com>\r\nCo-authored-by: Mateusz Sekara <mateusz.sekara@gmail.com>\r\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>\r\nCo-authored-by: Makram <makramkd@users.noreply.github.com>\r\nCo-authored-by: Ryan Stout <rstout610@gmail.com>",
          "timestamp": "2024-10-10T15:37:02+01:00",
          "tree_id": "a338140ff3a2ba8d6d8797313645d6b1b6dbd88b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/46370848a78919a4974f33871de164a93c49e772"
        },
        "date": 1728571079926,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449,
            "unit": "ns/op",
            "extra": "2599153 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.6,
            "unit": "ns/op",
            "extra": "2250219 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28872,
            "unit": "ns/op",
            "extra": "41188 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "athughlett@gmail.com",
            "name": "Awbrey Hughlett",
            "username": "EasterTheBunny"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2fd649133aced68d944e70f3f5a855c34c858d6d",
          "message": "created integration test that asserts cursor functionality (#811)\n\n* created integration test that asserts cursor functionality\r\n\r\n* fix query key with cursor over loop and grpc type conversion bug\r\n\r\n* Added finality to QueryKey Expression\r\n\r\n* Apply suggestions from code review\r\n\r\nCo-authored-by: Jordan Krage <jmank88@gmail.com>\r\n\r\n* Increased MaxWaitTimeforEvents for contractReader tests\r\n\r\n---------\r\n\r\nCo-authored-by: Silas Lenihan <sjl@lenihan.net>\r\nCo-authored-by: Silas Lenihan <32529249+silaslenihan@users.noreply.github.com>\r\nCo-authored-by: Jordan Krage <jmank88@gmail.com>",
          "timestamp": "2024-10-10T11:31:10-04:00",
          "tree_id": "69390df8ebe00b9a53e0859dc8e7ab0db094aeee",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2fd649133aced68d944e70f3f5a855c34c858d6d"
        },
        "date": 1728574323564,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.8,
            "unit": "ns/op",
            "extra": "2539358 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508,
            "unit": "ns/op",
            "extra": "2342630 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28351,
            "unit": "ns/op",
            "extra": "42517 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d1831b62389a1a7da70765a42a0db94f2b8f27b9",
          "message": "pkg/loop: clean up background contexts (#839)",
          "timestamp": "2024-10-10T13:54:45-05:00",
          "tree_id": "467ac00cb57fb1fb9f0dde6ba56a94569e6d9366",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d1831b62389a1a7da70765a42a0db94f2b8f27b9"
        },
        "date": 1728586556429,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.9,
            "unit": "ns/op",
            "extra": "2658181 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526,
            "unit": "ns/op",
            "extra": "2304422 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28273,
            "unit": "ns/op",
            "extra": "42500 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jin.bang@smartcontract.com",
            "name": "jinhoonbang",
            "username": "jinhoonbang"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b7d55eff04946ebf6fa65af13644520828eac990",
          "message": "parse workflow YAML as float, not decimal.Decimal (#841)\n\n* parse workflow YAML as float, not decimal.Decimal\r\n\r\n* support unwrapping as decimal.Decimal",
          "timestamp": "2024-10-11T08:29:40-07:00",
          "tree_id": "04f08e033d041fdee7164d1814142b4265c13f51",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b7d55eff04946ebf6fa65af13644520828eac990"
        },
        "date": 1728660653773,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.1,
            "unit": "ns/op",
            "extra": "2638287 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.4,
            "unit": "ns/op",
            "extra": "2316380 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28228,
            "unit": "ns/op",
            "extra": "42526 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "alec.gard@chainlinklabs.com",
            "name": "Alec Gard",
            "username": "alecgard"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5d432bcdc2e8c4e39369cb19183e3925afd37997",
          "message": "Fix Latest timestamp metric to return timestamp in seconds (#844)",
          "timestamp": "2024-10-11T11:09:13-05:00",
          "tree_id": "d48de52ee226c04f956e78b2695b98f4acfc6df7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5d432bcdc2e8c4e39369cb19183e3925afd37997"
        },
        "date": 1728663014387,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450,
            "unit": "ns/op",
            "extra": "2674831 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 541.4,
            "unit": "ns/op",
            "extra": "2182867 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28294,
            "unit": "ns/op",
            "extra": "42494 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tinianov@live.com",
            "name": "Ryan Tinianov",
            "username": "nolag"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3619db2c34a431b1459d349ddf52e6e57848b3ae",
          "message": "Enable consensus based on a value.Map (#840)\n\n* Enable consensus based on a value.Map\r\n\r\n* Add a method to get all encoder names",
          "timestamp": "2024-10-11T12:36:36-04:00",
          "tree_id": "27746a42ec7d619908d42e6e4460b2e41d8b5e29",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3619db2c34a431b1459d349ddf52e6e57848b3ae"
        },
        "date": 1728664668678,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.2,
            "unit": "ns/op",
            "extra": "2668617 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.6,
            "unit": "ns/op",
            "extra": "2187087 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28273,
            "unit": "ns/op",
            "extra": "42447 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "cdeea8fc821aa56809705e9ef5fd61b1e7bb8e0e",
          "message": "[chore] Some logging updates for feeds aggregator (#843)\n\n* [chore] Some logging updates for feeds aggregator\r\n\r\n* [chore] Some logging updates for feeds aggregator\r\n\r\nAlso pass the logger through the aggregator interface so that\r\nwe can inherit the logging tags, including executionID and workflowID\r\n\r\n* Use Errorw for consistency",
          "timestamp": "2024-10-14T10:43:07+01:00",
          "tree_id": "01c7ab91e40bb54a2324f5e6ddfe67dafc66fad4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cdeea8fc821aa56809705e9ef5fd61b1e7bb8e0e"
        },
        "date": 1728899051961,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 452.6,
            "unit": "ns/op",
            "extra": "2586141 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519,
            "unit": "ns/op",
            "extra": "2191023 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28285,
            "unit": "ns/op",
            "extra": "42472 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6c3cc4d0dc87f114328a309d786ae2dea8cc974f",
          "message": "[CAPPL-87] Support breaking a workflow from inside custom compute (#848)",
          "timestamp": "2024-10-14T13:28:10+01:00",
          "tree_id": "9d3637db9bc660a07402e0c18e558c1f1a66265a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6c3cc4d0dc87f114328a309d786ae2dea8cc974f"
        },
        "date": 1728908949055,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.1,
            "unit": "ns/op",
            "extra": "2363246 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512,
            "unit": "ns/op",
            "extra": "2336799 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28243,
            "unit": "ns/op",
            "extra": "42536 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "516fee04cae8deb4d8bf5063b612a9b18aaf251f",
          "message": "pkg/loop: swap eventually from assert to require (#851)",
          "timestamp": "2024-10-15T09:17:57-05:00",
          "tree_id": "d3b08bb9f5004f5d0a67f568da7a8f9ecbbb42f8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/516fee04cae8deb4d8bf5063b612a9b18aaf251f"
        },
        "date": 1729001940702,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 452.1,
            "unit": "ns/op",
            "extra": "2660906 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.8,
            "unit": "ns/op",
            "extra": "2270888 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28367,
            "unit": "ns/op",
            "extra": "42434 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "06ab6c310f4d1af42b0cdf47b0c86650baa28b93",
          "message": "keystone custom message proto (#828)\n\n* initial pass at keystone custom message proto\r\n\r\n* trying other proto defs\r\n\r\n* moving keystone to its own package\r\n\r\n* picking msg values proto version\r\n\r\n* KeystoneCustomMessage --> BaseCustomMessage w/ path refactoring\r\n\r\n* make generate\r\n\r\n---------\r\n\r\nCo-authored-by: Street <5597260+MStreet3@users.noreply.github.com>",
          "timestamp": "2024-10-15T17:26:58-04:00",
          "tree_id": "61f4e8d87c4873e44f9022a972ecee28e240af79",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/06ab6c310f4d1af42b0cdf47b0c86650baa28b93"
        },
        "date": 1729027678128,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449,
            "unit": "ns/op",
            "extra": "2684554 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 533.6,
            "unit": "ns/op",
            "extra": "2253849 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28291,
            "unit": "ns/op",
            "extra": "42506 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "62ca3f778e32f09dd22c6089f4afe533f80f839a",
          "message": "cleaning up CODEOWNERS (#853)",
          "timestamp": "2024-10-15T19:57:49-04:00",
          "tree_id": "74c84e094de884b03ae8005f7e185267558a4397",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/62ca3f778e32f09dd22c6089f4afe533f80f839a"
        },
        "date": 1729036735904,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.4,
            "unit": "ns/op",
            "extra": "2649942 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 536.6,
            "unit": "ns/op",
            "extra": "2336173 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28340,
            "unit": "ns/op",
            "extra": "42466 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "gaboparadiso@gmail.com",
            "name": "Gabriel Paradiso",
            "username": "agparadiso"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "935b2eeecf569610f556a42abfba702b94d27c79",
          "message": "[CAPPL-122] return error fetch handler (#852)\n\n* fix: return error instead of os.Exiting\r\n\r\n* fix: align FetchResponse to have ExecutionError instead of success field\r\n\r\n* fix: add error_message field",
          "timestamp": "2024-10-16T11:41:36+02:00",
          "tree_id": "10ca44fbe955863f640f37a62a1a8372ab7d5536",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/935b2eeecf569610f556a42abfba702b94d27c79"
        },
        "date": 1729071754314,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 452.9,
            "unit": "ns/op",
            "extra": "2568079 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.5,
            "unit": "ns/op",
            "extra": "2300023 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28293,
            "unit": "ns/op",
            "extra": "42372 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "87939adac36d8da9565ba7d9a76845d21084e32c",
          "message": "[chore] Add Beholder custom event for Workflows and Capabilities (#854)\n\n* [chore] Add Beholder custom event for Workflows and Capabilities\r\n\r\n* [chore] Add Beholder custom event for Workflows and Capabilities",
          "timestamp": "2024-10-16T13:31:51+01:00",
          "tree_id": "1c863e424f0db9b40570b554f031a36966f7d46a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/87939adac36d8da9565ba7d9a76845d21084e32c"
        },
        "date": 1729081979772,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.4,
            "unit": "ns/op",
            "extra": "2663198 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.2,
            "unit": "ns/op",
            "extra": "2334406 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28237,
            "unit": "ns/op",
            "extra": "42487 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c30fa4f97e9ca250fd195eb9960a21dbb4743d51",
          "message": "Update README.md (#857)\n\n* Update README.md\r\n\r\n* Update README.md",
          "timestamp": "2024-10-16T12:22:00-05:00",
          "tree_id": "3dc0bc8eda01ebd7546e66351833fc558d5dce9d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c30fa4f97e9ca250fd195eb9960a21dbb4743d51"
        },
        "date": 1729099379299,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 475.6,
            "unit": "ns/op",
            "extra": "2550999 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.4,
            "unit": "ns/op",
            "extra": "2327582 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28235,
            "unit": "ns/op",
            "extra": "42430 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b7b7f6310ac2815acf7a4deae0d3c5e33ac349d9",
          "message": "[chore] Add OracleSpecID to RelayArgs (#858)",
          "timestamp": "2024-10-16T18:35:14+01:00",
          "tree_id": "a17772635e36890af94712cdecf921eb67909ba8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b7b7f6310ac2815acf7a4deae0d3c5e33ac349d9"
        },
        "date": 1729100175618,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 470.7,
            "unit": "ns/op",
            "extra": "2626407 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.4,
            "unit": "ns/op",
            "extra": "2327421 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28510,
            "unit": "ns/op",
            "extra": "42538 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b6966277c23bf71bf68b4c4b4d3d832c4a9671ed",
          "message": "pkg/codec: add bix-framework & foundations as CODEOWNERS (#859)",
          "timestamp": "2024-10-16T17:04:10-04:00",
          "tree_id": "8ce32900802f9ef99684e19c3d37c715e72ccc2d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b6966277c23bf71bf68b4c4b4d3d832c4a9671ed"
        },
        "date": 1729112714426,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 475.5,
            "unit": "ns/op",
            "extra": "2564451 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.9,
            "unit": "ns/op",
            "extra": "2329320 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28995,
            "unit": "ns/op",
            "extra": "42513 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "bf362d3dd312f82031bbf3ac42b9268d4c4aaaef",
          "message": "Add a non-wasip1 version of NewRunner (#838)",
          "timestamp": "2024-10-17T10:36:06+01:00",
          "tree_id": "f0add02fd2467fcc6f9f24c097e3e84e414c03a8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bf362d3dd312f82031bbf3ac42b9268d4c4aaaef"
        },
        "date": 1729157839671,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 471.1,
            "unit": "ns/op",
            "extra": "2677143 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.3,
            "unit": "ns/op",
            "extra": "2341243 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28253,
            "unit": "ns/op",
            "extra": "42488 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "matthew.pendrey@gmail.com",
            "name": "Matthew Pendrey",
            "username": "ettec"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "80c6a3362575898e16c120ac9e24ece63816d909",
          "message": "contract reader api change for block meta data (#855)\n\n* contract reader api change for meta data\r\n\r\n* add to unimplemented contract reader\r\n\r\n* update to use Head struct for meta data\r\n\r\n* typo\r\n\r\n* another typo",
          "timestamp": "2024-10-17T10:55:39+01:00",
          "tree_id": "bc1fafea3017c0ef78144f1e3c0671a803332587",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/80c6a3362575898e16c120ac9e24ece63816d909"
        },
        "date": 1729159002165,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.9,
            "unit": "ns/op",
            "extra": "2507300 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.7,
            "unit": "ns/op",
            "extra": "2336270 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28257,
            "unit": "ns/op",
            "extra": "42488 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "02a8c3d034c76c948b1a6fb09237d382d0b5134a",
          "message": "[chore] use beholder/BaseMessage (#856)",
          "timestamp": "2024-10-17T11:16:09+01:00",
          "tree_id": "6765121a0a8e1a3073f5a4f36d886830e62b6716",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/02a8c3d034c76c948b1a6fb09237d382d0b5134a"
        },
        "date": 1729160233340,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 471.3,
            "unit": "ns/op",
            "extra": "2549640 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.2,
            "unit": "ns/op",
            "extra": "2354626 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28480,
            "unit": "ns/op",
            "extra": "42464 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "dimitrios.kouveris@smartcontract.com",
            "name": "dimitris",
            "username": "dimkouv"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b283b1e14fa6ae215d3b644d0c48a2b25edbea1e",
          "message": "change rmnreport struct (#861)",
          "timestamp": "2024-10-17T14:51:27+01:00",
          "tree_id": "208c7092b6622a261aee2ba3a23faf5586865fbf",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b283b1e14fa6ae215d3b644d0c48a2b25edbea1e"
        },
        "date": 1729173148668,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 472.4,
            "unit": "ns/op",
            "extra": "2522518 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.3,
            "unit": "ns/op",
            "extra": "2327547 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28275,
            "unit": "ns/op",
            "extra": "42387 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "dimitrios.kouveris@smartcontract.com",
            "name": "dimitris",
            "username": "dimkouv"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "92541641510fd337b88aefc2a229d4fa2d35eae4",
          "message": "Revert \"change rmnreport struct (#861)\" (#863)\n\nThis reverts commit b283b1e14fa6ae215d3b644d0c48a2b25edbea1e.",
          "timestamp": "2024-10-17T16:05:49+01:00",
          "tree_id": "6765121a0a8e1a3073f5a4f36d886830e62b6716",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/92541641510fd337b88aefc2a229d4fa2d35eae4"
        },
        "date": 1729177613929,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.3,
            "unit": "ns/op",
            "extra": "2515840 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.8,
            "unit": "ns/op",
            "extra": "2347476 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28284,
            "unit": "ns/op",
            "extra": "42454 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "309a9a3d51098ef1f5fb263428339a6faa64b45c",
          "message": "[CAPPL-66] Add custom_message package (#864)",
          "timestamp": "2024-10-18T11:12:31+01:00",
          "tree_id": "caf987fb7839dd87abb8edadb2fa1ae482cf0864",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/309a9a3d51098ef1f5fb263428339a6faa64b45c"
        },
        "date": 1729246414248,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.9,
            "unit": "ns/op",
            "extra": "2653500 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.5,
            "unit": "ns/op",
            "extra": "2347964 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28297,
            "unit": "ns/op",
            "extra": "42434 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "samsondav@protonmail.com",
            "name": "Sam",
            "username": "samsondav"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5248d7c4468aeddd5b42c383760174c6115e6416",
          "message": "Add types for Retirement Report (#835)\n\nCo-authored-by: Bruno Moura <brunotm@gmail.com>",
          "timestamp": "2024-10-18T10:37:28-04:00",
          "tree_id": "4e298e9afa58a66ab53dc068a3e1957d6ca85a65",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5248d7c4468aeddd5b42c383760174c6115e6416"
        },
        "date": 1729262318333,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 477.5,
            "unit": "ns/op",
            "extra": "2625934 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 527.3,
            "unit": "ns/op",
            "extra": "2315359 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28212,
            "unit": "ns/op",
            "extra": "42534 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "juan.farber@smartcontract.com",
            "name": "Juan Farber",
            "username": "Farber98"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a9f995ebb98b9ef7facfc7033c650512c5745cd1",
          "message": "[BCFR-147][common] - Add codec chain agnostic modifier for converting byte array address to string (#818)\n\n* initial ideation of chain agnostic modifier\r\n\r\n* fix existing tests and remove go-eth\r\n\r\n* fix ci lint\r\n\r\n* codec tests draft\r\n\r\n* added address field transf\r\n\r\n* fix codec tests\r\n\r\n* separate hooks\r\n\r\n* add solana support\r\n\r\n* fix solana and remove prints\r\n\r\n* cleanups\r\n\r\n* go mod dep\r\n\r\n* chain agnostic modifier\r\n\r\n* tidy and lint\r\n\r\n* addressing comments\r\n\r\n* reuse modifier base logic to simplify things + refactors\r\n\r\n---------\r\n\r\nCo-authored-by: Awbrey Hughlett <awbrey.hughlett@smartcontract.com>\r\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2024-10-18T18:30:14+02:00",
          "tree_id": "5a46c2d0bbc2d268b8828a53883959a7cbf66c67",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a9f995ebb98b9ef7facfc7033c650512c5745cd1"
        },
        "date": 1729269085751,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.3,
            "unit": "ns/op",
            "extra": "2546512 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521.7,
            "unit": "ns/op",
            "extra": "2313135 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28339,
            "unit": "ns/op",
            "extra": "42295 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "clement.erena78@gmail.com",
            "name": "Clement",
            "username": "Atrax1"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b3695e6094ac20ef03130d09069bbd051cdc4266",
          "message": "fix(observability-lib): improvements and fixes (#850)\n\n* fix(observability-lib): improvements and fixes\r\n\r\n* chore(observability-lib): refactor tests to compare json output\r\n\r\n* feat(observability-lib): can create alerts without attaching to a panel\r\n\r\n* feat(observability-lib): add colorscheme option to timeseries panel\r\n\r\n* feat(observability-lib): upgrade grafana sdk to latest version\r\n\r\n* fix(observability-lib): colorscheme for all panel type\r\n\r\n* fix(observability-lib): remove verbose flag test in Makefile\r\n\r\n* chore(observability-lib): change flag name for updating golden test file\r\n\r\n* chore(observability-lib): use t.cleanup instead of defer",
          "timestamp": "2024-10-19T12:16:24+02:00",
          "tree_id": "3b002eee8ca9dec61db76c7ebddbae8dc275f40e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b3695e6094ac20ef03130d09069bbd051cdc4266"
        },
        "date": 1729333047030,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.6,
            "unit": "ns/op",
            "extra": "2676412 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.3,
            "unit": "ns/op",
            "extra": "2291624 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28298,
            "unit": "ns/op",
            "extra": "42525 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cedric.cordenier@smartcontract.com",
            "name": "Cedric",
            "username": "cedric-cordenier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "39a6e78c028689698295c6b5d7ab208c4990f5e8",
          "message": "[CAPPL-132] Add secrets interpolation (#862)",
          "timestamp": "2024-10-21T11:35:00+01:00",
          "tree_id": "4cd69fa2ac348865dbbe794a2236accf220f9535",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/39a6e78c028689698295c6b5d7ab208c4990f5e8"
        },
        "date": 1729506972720,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.2,
            "unit": "ns/op",
            "extra": "2698375 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.2,
            "unit": "ns/op",
            "extra": "2351166 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28319,
            "unit": "ns/op",
            "extra": "40802 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vyzaldysanchez@gmail.com",
            "name": "Vyzaldy Sanchez",
            "username": "vyzaldysanchez"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "32bc8c118af44e58cda8e78dd4d5a0abc931c2df",
          "message": "Add `MetricsLabeler` to `custmsg` pkg (#869)\n\n* Adds `MetricsLabeler` to `custmsg` pkg\r\n\r\n* Moves labeler to correct pkg\r\n\r\n* Update pkg/monitoring/metrics_labeler.go\r\n\r\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>\r\n\r\n---------\r\n\r\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2024-10-21T10:22:10-06:00",
          "tree_id": "bc6a5f56f338c584cb319f8a6828e290ca25af7e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/32bc8c118af44e58cda8e78dd4d5a0abc931c2df"
        },
        "date": 1729527794354,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.4,
            "unit": "ns/op",
            "extra": "2670903 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.4,
            "unit": "ns/op",
            "extra": "2320904 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28656,
            "unit": "ns/op",
            "extra": "42507 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "177363085+pkcll@users.noreply.github.com",
            "name": "Pavel",
            "username": "pkcll"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4b45ad16ad7fa0b865decac91f60aa77b91bf147",
          "message": "TT-1303 INFOPLAT-1372 Add support for OTLP/HTTP exporters for beholder sdk (#830)\n\n* TT-1303 Add support for OTLP/HTTP exporters\r\n\r\n* TT-1303 Add support for OTLP/HTTP exporters: enable case when InsecureConnection:false and CACertFile is not set\r\n\r\n* insecureskipverify true\r\n\r\n* revert insecureSkipVerify true\r\n\r\n* Consolidate NewGRPCClient, NewHTTPClient into single constructor\r\n\r\n* Return nil on error from Client constructors\r\n\r\n* Add comment for used/unused context\r\n\r\n---------\r\n\r\nCo-authored-by: gheorghestrimtu <studentcuza@gmail.com>\r\nCo-authored-by: 4of9 <177086174+4of9@users.noreply.github.com>\r\nCo-authored-by: Geert G <117188496+cll-gg@users.noreply.github.com>\r\nCo-authored-by: Clement <clement.erena78@gmail.com>",
          "timestamp": "2024-10-21T20:25:00+02:00",
          "tree_id": "70996dbfecb4f518aeed6e31560db2f05734f493",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4b45ad16ad7fa0b865decac91f60aa77b91bf147"
        },
        "date": 1729535204200,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458,
            "unit": "ns/op",
            "extra": "2680203 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.5,
            "unit": "ns/op",
            "extra": "2360708 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28312,
            "unit": "ns/op",
            "extra": "42427 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "justinkaseman@live.com",
            "name": "Justin Kaseman",
            "username": "justinkaseman"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d48f9ab7b31ee1de80e487f24e90413cad037756",
          "message": "(fix): Unsupported types when nesting Value in ValuesMap (#866)\n\n* (test): Reproduce bug with Value into ValueMap to Proto\r\n\r\n* (fix): Correctly handle already wrapped BigInt, Bool, & Time nested in a Map",
          "timestamp": "2024-10-21T15:44:29-04:00",
          "tree_id": "58a91d21580a20aafb0e5edfa64e971d1f8f170f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d48f9ab7b31ee1de80e487f24e90413cad037756"
        },
        "date": 1729539929709,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 462.1,
            "unit": "ns/op",
            "extra": "2688417 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.1,
            "unit": "ns/op",
            "extra": "2327098 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28252,
            "unit": "ns/op",
            "extra": "42440 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "177086174+4of9@users.noreply.github.com",
            "name": "4of9",
            "username": "4of9"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2acfad0b9592ede3b6d0652b8907f6c55b930415",
          "message": "Beholder: Add domain and entity to metadata (#846)\n\n* Add BeholderDomain and BeholderEntity to Metadata\r\n\r\n* Panic on init error\r\n\r\n* Add additional domain & entity validation\r\n\r\n* Return error instead of panic",
          "timestamp": "2024-10-21T16:09:01-05:00",
          "tree_id": "842bf4262598cee1337a8fd80a10b88805bcb4be",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2acfad0b9592ede3b6d0652b8907f6c55b930415"
        },
        "date": 1729544995772,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.4,
            "unit": "ns/op",
            "extra": "2674351 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 506.2,
            "unit": "ns/op",
            "extra": "2378367 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28253,
            "unit": "ns/op",
            "extra": "42169 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "120329946+george-dorin@users.noreply.github.com",
            "name": "george-dorin",
            "username": "george-dorin"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f3cd964c341d2a27edd4d758fb0c45adc9c8bd2d",
          "message": "LOOPP Keystore (#837)\n\n* Initial draft\r\n\r\n* Add keystore service\r\n\r\n* Wire keystore factory\r\n\r\n* Update keystore proto namespace\r\n\r\n* Update keystore service\r\n\r\n* Add internal methods for keystores\r\n\r\n* Clean up Keystore GRPC methods\r\n\r\n* Add tests\r\n\r\n* Remove unused file\r\n\r\n* Update protoc version\r\n\r\n* Fix lint\r\n\r\n* Rename keystore interface methods\r\nExplain UDF method",
          "timestamp": "2024-10-22T10:56:05+03:00",
          "tree_id": "43ebbb302e26a17b0e0ce3fc76f2b385d0e21f7d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f3cd964c341d2a27edd4d758fb0c45adc9c8bd2d"
        },
        "date": 1729583839704,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.2,
            "unit": "ns/op",
            "extra": "2679843 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.5,
            "unit": "ns/op",
            "extra": "2315956 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28350,
            "unit": "ns/op",
            "extra": "42361 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "yepishevsanya@gmail.com",
            "name": "chudilka1",
            "username": "chudilka1"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ceeb47375a5679524c088fe04a66105efc1f7a71",
          "message": "Exclude tests and mocks from SonarQube coverage + Add LLM Error Reporter\n\n* Exclude tests and mocks from capabilities SonarQube coverage\r\n\r\n* Add LLM Action Error Reporter workflow\r\n\r\n* Fix Golangci-lint issues\r\n\r\n* Update golangci-version\r\n\r\n* Fix Golangci-lint issues",
          "timestamp": "2024-10-22T14:36:29+03:00",
          "tree_id": "aca8291d89e0bea4ff8c270a0fe1b7d6215ed008",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ceeb47375a5679524c088fe04a66105efc1f7a71"
        },
        "date": 1729597053385,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.7,
            "unit": "ns/op",
            "extra": "2641339 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.8,
            "unit": "ns/op",
            "extra": "2362430 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28274,
            "unit": "ns/op",
            "extra": "42349 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "5597260+MStreet3@users.noreply.github.com",
            "name": "Street",
            "username": "MStreet3"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9225bc1a857e6395f446094786c74cdd07cfa73e",
          "message": "[cappl-86] feat(workflows/wasm): emit msgs to beholder (#845)\n\n* wip(wasm): adds Emit to Runtime interface\r\n\r\nWIP on Runtime with panics\r\n\r\n* refactor(wasm): separte funcs out of NewRunner\r\n\r\n* refactor(wasm): shifts logging related funcs around\r\n\r\n* feat(wasm): adds custom pb message\r\n\r\n* feat(wasm): calls emit from guest runner\r\n\r\n* refactor(workflows): splits out emitter interface + docstring\r\n\r\n* feat(host): defines a beholder adapter for emitter\r\n\r\n* wip(host): implement host side emit\r\n\r\n* refactor(wasm/host): abstracts read and write to wasm\r\n\r\n* protos wip\r\n\r\n* feat(wasm): emits error response\r\n\r\n* refactor(wasm/host): write all failures from wasm to memory\r\n\r\n* feat(wasm): inject metadata into module\r\n\r\n* feat(events+wasm): pull emit md from req md\r\n\r\n* feat(custmsg): creates labels from map\r\n\r\n* feat(wasm): adds tests and validates labels\r\n\r\n* feat(wasm/host): use custmsg implementation for calling beholder\r\n\r\n* chore(wasm+host): docstrings and lint\r\n\r\n* chore(host): new emitter iface + private func types\r\n\r\n* chore(multi) review comments\r\n\r\n* chore(wasm): add id and md to config directly\r\n\r\n* refactor(custmsg+host): adapter labeler from config for emit\r\n\r\n* refactor(wasm): remove emitter from mod config\r\n\r\n* refactor(custmsg+wasm): expose emitlabeler on guest\r\n\r\n* refactor(wasm+sdk): EmitLabeler to MessageEmitter\r\n\r\n* refactor(wasm+events): share label keys\r\n\r\n* refactor(wasm+values): use map[string]string directly",
          "timestamp": "2024-10-22T17:02:10+01:00",
          "tree_id": "2424df4681667f7ae41421475dbb6d8b0cf8b6e6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9225bc1a857e6395f446094786c74cdd07cfa73e"
        },
        "date": 1729612994746,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457.6,
            "unit": "ns/op",
            "extra": "2654148 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.4,
            "unit": "ns/op",
            "extra": "2359779 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28357,
            "unit": "ns/op",
            "extra": "39499 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vyzaldysanchez@gmail.com",
            "name": "Vyzaldy Sanchez",
            "username": "vyzaldysanchez"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8b1c952d3911d4157b67523b75880648b5515e0b",
          "message": "Move labelers to new pkg (#875)\n\n* Moves labelers to new pkg\r\n\r\n* Moves labelers to top level pkgs",
          "timestamp": "2024-10-22T13:39:57-04:00",
          "tree_id": "9c43a53be655b931fea5168907fac16cf325d335",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8b1c952d3911d4157b67523b75880648b5515e0b"
        },
        "date": 1729618855375,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.8,
            "unit": "ns/op",
            "extra": "2652798 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 529.1,
            "unit": "ns/op",
            "extra": "2280459 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28416,
            "unit": "ns/op",
            "extra": "40540 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b772997e9a33a83a65a13ee192601610cf2782cf",
          "message": "removing values.Map from BaseMessage proto as Beholder only supports root level protos (#876)\n\n* removing values.Map from BaseMessage proto as Beholder only supports root level protos\r\n\r\n* fixing import\r\n\r\n* lint",
          "timestamp": "2024-10-22T13:55:31-07:00",
          "tree_id": "ae532661786a7e23880a75bababd4c06757e05d3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b772997e9a33a83a65a13ee192601610cf2782cf"
        },
        "date": 1729630590509,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.2,
            "unit": "ns/op",
            "extra": "2679882 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507.8,
            "unit": "ns/op",
            "extra": "2347501 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29650,
            "unit": "ns/op",
            "extra": "42439 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "84d3ef9662b7c36787231434286b2acdb9d83942",
          "message": "golangci-lint: add rules (#872)\n\n* golangci-lint: add rules\r\n\r\n* bump golangci-lint\r\n\r\n* Bump ci-lint-go action version to include only-new-issues config\r\n\r\n---------\r\n\r\nCo-authored-by: Alexandr Yepishev <alexandr.yepishev@smartcontract.com>",
          "timestamp": "2024-10-22T18:12:57-05:00",
          "tree_id": "f83067cc6c9da5a8450841b269f1197eb91e96e6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/84d3ef9662b7c36787231434286b2acdb9d83942"
        },
        "date": 1729638845652,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.8,
            "unit": "ns/op",
            "extra": "2677696 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 506.6,
            "unit": "ns/op",
            "extra": "2366748 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28323,
            "unit": "ns/op",
            "extra": "41473 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "patrick.huie@smartcontract.com",
            "name": "Patrick",
            "username": "patrickhuie19"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c5c856ee23d16d2ad55bd0d47e0b9e845cf37e78",
          "message": "Revert \"removing values.Map from BaseMessage proto as Beholder only supports root level protos (#876)\" (#881)\n\nThis reverts commit b772997e9a33a83a65a13ee192601610cf2782cf.",
          "timestamp": "2024-10-22T21:52:35-04:00",
          "tree_id": "f17df075cd0ae937c53e1175f69f8ed491c95cfb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c5c856ee23d16d2ad55bd0d47e0b9e845cf37e78"
        },
        "date": 1729648416295,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.3,
            "unit": "ns/op",
            "extra": "2617527 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517,
            "unit": "ns/op",
            "extra": "2347918 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28366,
            "unit": "ns/op",
            "extra": "42446 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jmank88@gmail.com",
            "name": "Jordan Krage",
            "username": "jmank88"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "485f3f97cdbd0abc80e3a6263364ddabefa6d7f0",
          "message": "pkg/utils: add NewSleeperTaskCtx(WorkerCtx) (#868)",
          "timestamp": "2024-10-23T09:24:48-05:00",
          "tree_id": "6000a766085f467bf0546afd35613d17a5daf981",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/485f3f97cdbd0abc80e3a6263364ddabefa6d7f0"
        },
        "date": 1729693550131,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 492.4,
            "unit": "ns/op",
            "extra": "2650452 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.5,
            "unit": "ns/op",
            "extra": "2333247 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28281,
            "unit": "ns/op",
            "extra": "42451 times\n4 procs"
          }
        ]
      }
    ]
  }
}