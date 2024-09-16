window.BENCHMARK_DATA = {
  "lastUpdate": 1726504499114,
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
      }
    ]
  }
}