window.BENCHMARK_DATA = {
  "lastUpdate": 1751451011714,
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
          "id": "989addce9e4330dc0b92a4c3e1bfb336d43f9d1c",
          "message": "Bump settings (#884)\n\n* Bump settings\r\n\r\n* Fix test\r\n\r\n---------\r\n\r\nCo-authored-by: Vyzaldy Sanchez <vyzaldysanchez@gmail.com>",
          "timestamp": "2024-10-23T13:21:44-04:00",
          "tree_id": "f4d1ce4170bb760897633755ca487b3d74dca25f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/989addce9e4330dc0b92a4c3e1bfb336d43f9d1c"
        },
        "date": 1729704162701,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 445.9,
            "unit": "ns/op",
            "extra": "2676320 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 528.2,
            "unit": "ns/op",
            "extra": "2224357 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42369 times\n4 procs"
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
          "id": "86c89e29937d98bbce13bb7673c072f6c4b53343",
          "message": "Fix Foundry Shared Tests CI (#882)",
          "timestamp": "2024-10-23T22:42:19+02:00",
          "tree_id": "194e5605eea13037c328efae2a3870674abc9d81",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/86c89e29937d98bbce13bb7673c072f6c4b53343"
        },
        "date": 1729716194995,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448,
            "unit": "ns/op",
            "extra": "2664765 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 528.9,
            "unit": "ns/op",
            "extra": "2353148 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28276,
            "unit": "ns/op",
            "extra": "42488 times\n4 procs"
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
          "id": "221839275fbdadbfa86238793bb7616220698b64",
          "message": "pkg/services: add StopChan.CtxWithTimeout() (#879)",
          "timestamp": "2024-10-23T15:56:01-05:00",
          "tree_id": "be8da2023c9b3a142f7adf27ba071db8297b3985",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/221839275fbdadbfa86238793bb7616220698b64"
        },
        "date": 1729717019016,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.2,
            "unit": "ns/op",
            "extra": "2702540 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.5,
            "unit": "ns/op",
            "extra": "2279122 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28253,
            "unit": "ns/op",
            "extra": "42460 times\n4 procs"
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
          "id": "780e3d4c3ea9c7d55ede72c42bdb2188d01babdc",
          "message": "refactor(custmsg+wasm): adjust message emitter iface (#885)",
          "timestamp": "2024-10-23T17:40:34-04:00",
          "tree_id": "d2ea33ef9b31f11dd7f9c2280886e28e75a13504",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/780e3d4c3ea9c7d55ede72c42bdb2188d01babdc"
        },
        "date": 1729719692268,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.9,
            "unit": "ns/op",
            "extra": "2685976 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 524.6,
            "unit": "ns/op",
            "extra": "2307256 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28183,
            "unit": "ns/op",
            "extra": "42445 times\n4 procs"
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
          "id": "eaedfe1e99c74bd75ab1683be6a72b90aae7fb25",
          "message": "updating beholder for new domain model (#889)",
          "timestamp": "2024-10-24T15:50:54-04:00",
          "tree_id": "692e9ddd38dc9b24b87e43ea71157a10dd113efb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eaedfe1e99c74bd75ab1683be6a72b90aae7fb25"
        },
        "date": 1729799519584,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.2,
            "unit": "ns/op",
            "extra": "2686844 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 536.6,
            "unit": "ns/op",
            "extra": "2316738 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28287,
            "unit": "ns/op",
            "extra": "42417 times\n4 procs"
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
          "id": "c968705809fc4f77fc08906954cc13305ff4bd73",
          "message": "fix(observability-lib):health uptime value correct within node alert (#887)",
          "timestamp": "2024-10-25T12:13:08+02:00",
          "tree_id": "3e16f455d63d118e59935ba82f4ae97d7e150f7e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c968705809fc4f77fc08906954cc13305ff4bd73"
        },
        "date": 1729851255657,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.5,
            "unit": "ns/op",
            "extra": "2691213 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.4,
            "unit": "ns/op",
            "extra": "2363804 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28310,
            "unit": "ns/op",
            "extra": "39860 times\n4 procs"
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
          "id": "cfad021395954c9dd55ffe5d8e8ac2eb41661b25",
          "message": "[CAPPL-121] context propagation (#883)\n\nfeat: context propagation",
          "timestamp": "2024-10-25T15:20:45+02:00",
          "tree_id": "0d98ae14a3c8e3f14f9aee432a2dec6d81513027",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cfad021395954c9dd55ffe5d8e8ac2eb41661b25"
        },
        "date": 1729862509642,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.8,
            "unit": "ns/op",
            "extra": "2690922 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 535.1,
            "unit": "ns/op",
            "extra": "2325992 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28382,
            "unit": "ns/op",
            "extra": "42391 times\n4 procs"
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
          "id": "d5e98824b251e163797205ab00db54b74d937441",
          "message": ".github: add pull request template (#880)",
          "timestamp": "2024-10-25T09:04:23-05:00",
          "tree_id": "1c72a49fea4935545aa9d61ff89a1261bde1efaa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d5e98824b251e163797205ab00db54b74d937441"
        },
        "date": 1729865126507,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.9,
            "unit": "ns/op",
            "extra": "2547292 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.7,
            "unit": "ns/op",
            "extra": "2350600 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28291,
            "unit": "ns/op",
            "extra": "42240 times\n4 procs"
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
          "id": "99d0b847a001140721c757c1caae4a9c52c281a0",
          "message": "Allow user structs to be generated as capability definitions. (#873)",
          "timestamp": "2024-10-25T10:26:36-04:00",
          "tree_id": "29afcfa71bf4f32112785d644e087f0516fa5a93",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/99d0b847a001140721c757c1caae4a9c52c281a0"
        },
        "date": 1729866454670,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.8,
            "unit": "ns/op",
            "extra": "2674561 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517,
            "unit": "ns/op",
            "extra": "2343962 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28358,
            "unit": "ns/op",
            "extra": "42345 times\n4 procs"
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
          "id": "eed6b5f1be5d1d7b947ea57de81858e3bd6e8e62",
          "message": "Reapply \"removing values.Map from BaseMessage proto as Beholder only supports root level protos (#876)\" (#881) (#886)\n\nThis reverts commit c5c856ee23d16d2ad55bd0d47e0b9e845cf37e78.",
          "timestamp": "2024-10-25T13:09:01-04:00",
          "tree_id": "d0ae0ed9ecebda8f686941a590efcfeb21a4865d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eed6b5f1be5d1d7b947ea57de81858e3bd6e8e62"
        },
        "date": 1729876204293,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.6,
            "unit": "ns/op",
            "extra": "2638561 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 516.6,
            "unit": "ns/op",
            "extra": "2314929 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28743,
            "unit": "ns/op",
            "extra": "39628 times\n4 procs"
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
          "id": "8529bcce7eee0ba24ecdf1f96d3de792e9e7e497",
          "message": "[CAPPL-132] Support secrets in the Builder SDK (#888)",
          "timestamp": "2024-10-28T11:10:26Z",
          "tree_id": "fba52b8514b1ba802d647773359463a60bfec60f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8529bcce7eee0ba24ecdf1f96d3de792e9e7e497"
        },
        "date": 1730113884530,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.7,
            "unit": "ns/op",
            "extra": "2691332 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.4,
            "unit": "ns/op",
            "extra": "2183187 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28284,
            "unit": "ns/op",
            "extra": "42291 times\n4 procs"
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
          "id": "fcae9bd87a4214fd1ae78307f5b5cf757d0ba84a",
          "message": "[CAPPL-128] limit the amount of fetch calls per request (#894)\n\n* feat: limit the amount of fetch calls per request\r\n\r\n* fix: move counter to a request level\r\n\r\n* fix: define defaultMaxFetchRequests as a const\r\n\r\n* fix: rename and change type of fetchRequestsCounter",
          "timestamp": "2024-10-28T13:02:14Z",
          "tree_id": "b84debd191db193fd6b2687750c350f26d4e3c0e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/fcae9bd87a4214fd1ae78307f5b5cf757d0ba84a"
        },
        "date": 1730120599244,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.7,
            "unit": "ns/op",
            "extra": "2694259 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.7,
            "unit": "ns/op",
            "extra": "2284320 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28293,
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
          "id": "97f539091c3ceb97487a631b652c7dec32d91373",
          "message": "pkg/services: add (*Engine).Tracer() (#878)\n\nCo-authored-by: Vyzaldy Sanchez <vyzaldysanchez@gmail.com>",
          "timestamp": "2024-10-28T10:17:07-05:00",
          "tree_id": "80dc3f0ed6a0021adc6c4a9c09294f0fdfa59a3e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/97f539091c3ceb97487a631b652c7dec32d91373"
        },
        "date": 1730128696731,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.5,
            "unit": "ns/op",
            "extra": "2567205 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.9,
            "unit": "ns/op",
            "extra": "2361620 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28603,
            "unit": "ns/op",
            "extra": "42501 times\n4 procs"
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
          "id": "7c8207d66824c48aa54d330b05b9818290d43a76",
          "message": "feat(observability-lib): improve alerting provisioning (#893)\n\n* feat(observability-lib): improve alerting provisioning\r\n\r\n* feat(observability-lib): node general dashboard add log panel\r\n\r\n* feat(observability-lib): add notification template for pagerduty\r\n\r\n* chore(observability-lib): update ref files for tests",
          "timestamp": "2024-10-29T00:12:30+01:00",
          "tree_id": "ebb16e4107f09d6c2e895b1e8121ff2c5640963c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7c8207d66824c48aa54d330b05b9818290d43a76"
        },
        "date": 1730157206864,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.4,
            "unit": "ns/op",
            "extra": "2685238 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 516.7,
            "unit": "ns/op",
            "extra": "2318488 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28292,
            "unit": "ns/op",
            "extra": "42346 times\n4 procs"
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
          "id": "3a12ebe6d7fe90a4fcd2cacac4877dd53b1ffb59",
          "message": "Fix a bug where only one file is used for user type generation (#897)",
          "timestamp": "2024-10-29T12:26:06-04:00",
          "tree_id": "c87e7e0fa8632f9f18d11ca70f4d415f1080a672",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3a12ebe6d7fe90a4fcd2cacac4877dd53b1ffb59"
        },
        "date": 1730219227992,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.2,
            "unit": "ns/op",
            "extra": "2675689 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.5,
            "unit": "ns/op",
            "extra": "2350359 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29595,
            "unit": "ns/op",
            "extra": "42415 times\n4 procs"
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
          "id": "f2d327f5ac17483904c18268a077f4b6b11e3c57",
          "message": "custmsg domain chainlink --> platform (#901)",
          "timestamp": "2024-10-29T15:48:00-04:00",
          "tree_id": "3149e49e9169015948aa9e8f73dfa80a302d54ae",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f2d327f5ac17483904c18268a077f4b6b11e3c57"
        },
        "date": 1730231347940,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.5,
            "unit": "ns/op",
            "extra": "2674964 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 510.1,
            "unit": "ns/op",
            "extra": "2355636 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28302,
            "unit": "ns/op",
            "extra": "42434 times\n4 procs"
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
          "id": "0c0f971b1e73fb7cf79372fadce77d20de34cbe7",
          "message": "fix consensus cap encoder debug message (#900)",
          "timestamp": "2024-10-30T14:09:22-04:00",
          "tree_id": "66eb5d76c48cb67523e3785fdade7887f5557791",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0c0f971b1e73fb7cf79372fadce77d20de34cbe7"
        },
        "date": 1730311821995,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.2,
            "unit": "ns/op",
            "extra": "2681024 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.4,
            "unit": "ns/op",
            "extra": "2265238 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28284,
            "unit": "ns/op",
            "extra": "42475 times\n4 procs"
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
          "id": "a98d27835a2f792aa5553c037f63f87be08c89d2",
          "message": "[CAPPL-214] Move secrets library; add decrypt function (#906)",
          "timestamp": "2024-10-30T21:38:00Z",
          "tree_id": "c3dc89c763ec2e95b5b23a752abc817b4226ffa7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a98d27835a2f792aa5553c037f63f87be08c89d2"
        },
        "date": 1730324342006,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.5,
            "unit": "ns/op",
            "extra": "2692136 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.7,
            "unit": "ns/op",
            "extra": "2357005 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28254,
            "unit": "ns/op",
            "extra": "42507 times\n4 procs"
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
          "id": "4dc0db60d95434d3d77fb37bde1757711c2f218a",
          "message": "[CAPPL-195] preserve errors across the WASM boundary (#899)\n\n* feat: preserve errors across the WASM boundary\r\n\r\n---------\r\n\r\nCo-authored-by: Cedric <cedric.cordenier@smartcontract.com>",
          "timestamp": "2024-10-31T14:49:29+01:00",
          "tree_id": "1ff861dbc599ed0ccd3e2132d84a05e6b0c006a1",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4dc0db60d95434d3d77fb37bde1757711c2f218a"
        },
        "date": 1730382626809,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.7,
            "unit": "ns/op",
            "extra": "2672193 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.9,
            "unit": "ns/op",
            "extra": "2120512 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28284,
            "unit": "ns/op",
            "extra": "42369 times\n4 procs"
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
          "id": "2c5eeefca3f057846e2e346f4fda4deaaedef8d0",
          "message": "(feat): New OCR3 Consensus Capability Aggregator: Reduce (#842)\n\n* (feat): New OCR3 consenus capability aggregator: Reduce\r\n\r\n* Refactor to be more dynamic\r\n\r\n* Add reduce consensus schema\r\n\r\n* Changes from review\r\n\r\n* (feat): Add subMaps\r\n\r\n* Add more unit tests\r\n\r\n* Use agg config in workflow sdk type\r\n\r\n* Pass through config values to workflow test\r\n\r\n* PoR example\r\n\r\n* Fix for upstream change\r\n\r\n* Changes from review: Keep state using OutputKeys, add new reportFormat value, more tests\r\n\r\n* fix test output type",
          "timestamp": "2024-10-31T12:10:16-04:00",
          "tree_id": "378c3c37fd44bb98c65f004c4e1281675df95a63",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2c5eeefca3f057846e2e346f4fda4deaaedef8d0"
        },
        "date": 1730391074198,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.8,
            "unit": "ns/op",
            "extra": "2652906 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 530.5,
            "unit": "ns/op",
            "extra": "2244826 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28284,
            "unit": "ns/op",
            "extra": "42466 times\n4 procs"
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
          "id": "bd86494c574b9e3819055653655c14f833772589",
          "message": "[chore] Move validate function to secrets package (#908)",
          "timestamp": "2024-10-31T16:26:15-04:00",
          "tree_id": "fa02af95589a70459b3c4f2ac03eb41c7622392f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bd86494c574b9e3819055653655c14f833772589"
        },
        "date": 1730406435442,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.4,
            "unit": "ns/op",
            "extra": "2672859 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.4,
            "unit": "ns/op",
            "extra": "2349866 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28504,
            "unit": "ns/op",
            "extra": "42475 times\n4 procs"
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
          "id": "33711d0c3de7406b3543fafc95c8d64134ab2ea0",
          "message": "[CAPPL-182] Add context to message emitter (#909)\n\n* [CAPPL-182] Add context to message emitter\r\n\r\n* Run tests against uncompressed binaries by default\r\n\r\n* [chore] Move validate function to secrets package (#908)\r\n\r\n- Run tests against uncompressed binaries by default",
          "timestamp": "2024-11-01T09:38:30Z",
          "tree_id": "7aa6c284a34de66ac73bea458f4cbc5cda47869d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/33711d0c3de7406b3543fafc95c8d64134ab2ea0"
        },
        "date": 1730453969435,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.5,
            "unit": "ns/op",
            "extra": "2694564 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.5,
            "unit": "ns/op",
            "extra": "2217003 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28260,
            "unit": "ns/op",
            "extra": "42452 times\n4 procs"
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
          "id": "4b0948d48f16a864c3e9106469ee3024608aa56e",
          "message": "pkg/capabilities/triggers: MercuryTriggerService.Name() use Logger.Name() (#907)",
          "timestamp": "2024-11-01T09:37:26-05:00",
          "tree_id": "7754fdd7ff7f2b3e289f6ff70568a726b020677a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4b0948d48f16a864c3e9106469ee3024608aa56e"
        },
        "date": 1730471907269,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 528.3,
            "unit": "ns/op",
            "extra": "1892893 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507.6,
            "unit": "ns/op",
            "extra": "2358188 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28286,
            "unit": "ns/op",
            "extra": "42351 times\n4 procs"
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
          "id": "2780401fa0194459154776709515529afa572f6d",
          "message": "inline documentation for working with the codec (#895)\n\ninline documentation for working with the codec\r\n\r\n---------\r\n\r\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2024-11-01T14:45:58-05:00",
          "tree_id": "df8aa18d058849b4bf76d3af38bbcc7322bdb450",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2780401fa0194459154776709515529afa572f6d"
        },
        "date": 1730490417900,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.3,
            "unit": "ns/op",
            "extra": "2580826 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 505.6,
            "unit": "ns/op",
            "extra": "2343070 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28529,
            "unit": "ns/op",
            "extra": "42406 times\n4 procs"
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
          "id": "9a99ee1eeb57b14f63ef4a613cf8d346a61a9521",
          "message": "Fix loop getAttributes (#902)\n\n* Rename getAttributes to getMap\r\n\r\n* Fix getMap",
          "timestamp": "2024-11-01T15:21:13-05:00",
          "tree_id": "6293b259302453aaee38dc1f13ce90d6fa7e4ab7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9a99ee1eeb57b14f63ef4a613cf8d346a61a9521"
        },
        "date": 1730492533128,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 487.8,
            "unit": "ns/op",
            "extra": "2339157 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.2,
            "unit": "ns/op",
            "extra": "2332786 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28265,
            "unit": "ns/op",
            "extra": "42416 times\n4 procs"
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
          "id": "3072d4cf1ba45f0711ec4afc33fa2c39dff33315",
          "message": "[chore] Small optimizations (#912)\n\n* [chore] Make sure we clean up all WASM resources\r\n\r\n* Add cache settings",
          "timestamp": "2024-11-04T11:07:37Z",
          "tree_id": "ca31aeab267ac7011e39a31b0ff71592fcf87d8c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3072d4cf1ba45f0711ec4afc33fa2c39dff33315"
        },
        "date": 1730718519048,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.6,
            "unit": "ns/op",
            "extra": "2688072 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.8,
            "unit": "ns/op",
            "extra": "2365987 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28288,
            "unit": "ns/op",
            "extra": "42465 times\n4 procs"
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
          "id": "abf966b1e082add66848cd0b1906e4bfe81d7d9f",
          "message": "parallelize tests in slow packages (#911)",
          "timestamp": "2024-11-04T06:54:47-07:00",
          "tree_id": "f4017067219389a622bcf96aeb3c3214a397dff8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/abf966b1e082add66848cd0b1906e4bfe81d7d9f"
        },
        "date": 1730728548653,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.1,
            "unit": "ns/op",
            "extra": "2663847 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.4,
            "unit": "ns/op",
            "extra": "2334397 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28657,
            "unit": "ns/op",
            "extra": "42364 times\n4 procs"
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
          "distinct": false,
          "id": "d9732b30cbefb9e698f515b93d96ff11f89a2f94",
          "message": "feat(observability-lib): multiple alert rules can be attached to panel (#913)\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2024-11-04T17:19:59+01:00",
          "tree_id": "024adcf34c59dd5338e97122c624dce389b816f6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d9732b30cbefb9e698f515b93d96ff11f89a2f94"
        },
        "date": 1730737264722,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.1,
            "unit": "ns/op",
            "extra": "2655692 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 555.3,
            "unit": "ns/op",
            "extra": "2358320 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42355 times\n4 procs"
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
          "id": "eed4b097bcca4b0739774f9461bc9de29b1f0267",
          "message": "INFOPLAT-1376 Be able to configure retries for beholder otel exporters (#867)\n\n* Enable retries for otel exporters\r\n\r\n* Remove Enabled field from Beholder RetryConfig\r\n\r\n* Use retry config in Beholder HTTP client\r\n\r\n* Use beholder retry config only if its set\r\n\r\n* Update pkg/beholder/config_test.go\r\n\r\nCo-authored-by: Jordan Krage <jmank88@gmail.com>\r\n\r\n* Rename retry config fields\r\n\r\n* Fix golangci-lint errors\r\n\r\n* Rename EmitterRetryConfig -> LogRetryConfig\r\n\r\n* Move LogRetryConfig\r\n\r\n---------\r\n\r\nCo-authored-by: Jordan Krage <jmank88@gmail.com>",
          "timestamp": "2024-11-04T17:28:59+01:00",
          "tree_id": "c8835b4fe8be73bd6a01251a678b4b48e0a4c4ed",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eed4b097bcca4b0739774f9461bc9de29b1f0267"
        },
        "date": 1730737806391,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.2,
            "unit": "ns/op",
            "extra": "2663882 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 505.5,
            "unit": "ns/op",
            "extra": "2197396 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28272,
            "unit": "ns/op",
            "extra": "42481 times\n4 procs"
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
          "id": "15c5bee0552195acd483d96cce1ac780333c07c9",
          "message": "Add codec wrapper modifier (#905)\n\n* Add codec wrapper modifier\r\n\r\n* Fix WrapperModifierConfig description comment\r\n\r\n* Improve comments for wrapper modifier",
          "timestamp": "2024-11-05T11:49:41+01:00",
          "tree_id": "745a2016b0de2bc7fe38ac998b1387af4b815f40",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/15c5bee0552195acd483d96cce1ac780333c07c9"
        },
        "date": 1730803846570,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.8,
            "unit": "ns/op",
            "extra": "2673114 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507.1,
            "unit": "ns/op",
            "extra": "2137224 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28389,
            "unit": "ns/op",
            "extra": "42421 times\n4 procs"
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
          "id": "e1b7c81d582aeb8d73642fc2c8e2cc04913569fe",
          "message": "Extend `FetchRequest` fields for beholder support (#917)\n\n* Extends `FetchRequest` fields\r\n\r\n* Fixes proto version\r\n\r\n* Exposes `workflowId`\r\n\r\n* Adds proto comment\r\n\r\n* Adds metadata",
          "timestamp": "2024-11-05T12:33:18-04:00",
          "tree_id": "c8f0b9647eaa8978ad8ad040ba3c2a4367b9bf04",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e1b7c81d582aeb8d73642fc2c8e2cc04913569fe"
        },
        "date": 1730824458425,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.7,
            "unit": "ns/op",
            "extra": "2641656 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.4,
            "unit": "ns/op",
            "extra": "2376664 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28531,
            "unit": "ns/op",
            "extra": "42415 times\n4 procs"
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
          "id": "f1c901fd6191bf28e839b3d0a5b30854baefd995",
          "message": "feat(observability-lib): builder to create independently grafana resources (#915)\n\n* feat(observability-lib): builder to create independently grafana resources\r\n\r\n* fix(observability-lib): notification policy matchers check",
          "timestamp": "2024-11-05T19:30:26+01:00",
          "tree_id": "55c2cbcaee95f6db1fcdc18050a3519713252942",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f1c901fd6191bf28e839b3d0a5b30854baefd995"
        },
        "date": 1730831491238,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.4,
            "unit": "ns/op",
            "extra": "2646678 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.5,
            "unit": "ns/op",
            "extra": "2376554 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28287,
            "unit": "ns/op",
            "extra": "42333 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cfal@users.noreply.github.com",
            "name": "cfal",
            "username": "cfal"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c7bded1c08ae54419c07501f396840763f3e609c",
          "message": "pkg/config/validate.go: check CanInterface on subitems correctly (#918)",
          "timestamp": "2024-11-06T22:20:51+08:00",
          "tree_id": "6eb5489cd9a6d9aea4e4e9fbc2032c433e709eaf",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c7bded1c08ae54419c07501f396840763f3e609c"
        },
        "date": 1730902922806,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.6,
            "unit": "ns/op",
            "extra": "2634972 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 505.8,
            "unit": "ns/op",
            "extra": "2367771 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28279,
            "unit": "ns/op",
            "extra": "42210 times\n4 procs"
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
          "id": "aa5186fe92b4af2378e6fab896addd8a8c886072",
          "message": "Register ContractReader gRPC service fixes and tests (#921)",
          "timestamp": "2024-11-06T19:28:50Z",
          "tree_id": "3b93e962d0ddbc28e702e7c421702313e46387bc",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aa5186fe92b4af2378e6fab896addd8a8c886072"
        },
        "date": 1730921389468,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.9,
            "unit": "ns/op",
            "extra": "2658676 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 503.7,
            "unit": "ns/op",
            "extra": "2228364 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28718,
            "unit": "ns/op",
            "extra": "42397 times\n4 procs"
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
          "id": "3b320ad9b7c45432e44dde9c5441b8d6ce03d33c",
          "message": "(fix): Enforce Consensus Capability config field key_id (#892)\n\n* (fix): Enforce Consensus Capability config field key_id\r\n\r\n* Generate\r\n\r\n* (test): Fix from merge, add key_id to another test\r\n\r\n---------\r\n\r\nCo-authored-by: Bolek <1416262+bolekk@users.noreply.github.com>",
          "timestamp": "2024-11-06T15:25:32-05:00",
          "tree_id": "41ad935712096685b34c7925f773bc55ed925810",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3b320ad9b7c45432e44dde9c5441b8d6ce03d33c"
        },
        "date": 1730924799454,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.6,
            "unit": "ns/op",
            "extra": "2584399 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.4,
            "unit": "ns/op",
            "extra": "2361644 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28613,
            "unit": "ns/op",
            "extra": "42397 times\n4 procs"
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
          "id": "25e45ecd73ba518f955995dd3aef97d1318cee17",
          "message": "Fix codec wrapper modifier config unmarshall and add tests (#922)",
          "timestamp": "2024-11-07T14:42:05+01:00",
          "tree_id": "4f4623ce99c0a6cc6f2e3dbc6f081ee45f00f564",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/25e45ecd73ba518f955995dd3aef97d1318cee17"
        },
        "date": 1730986992102,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.9,
            "unit": "ns/op",
            "extra": "2678180 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.2,
            "unit": "ns/op",
            "extra": "2235960 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28289,
            "unit": "ns/op",
            "extra": "42418 times\n4 procs"
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
          "id": "67dff0fbb60dbbb2dfb8d1ce6473217e1e9d0ced",
          "message": "pkg/loop: ClientConfig.SkipHostEnv=true (#924)",
          "timestamp": "2024-11-07T14:26:38-07:00",
          "tree_id": "311ada07da02cecc3c3595d83da544bde0f3fb4e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/67dff0fbb60dbbb2dfb8d1ce6473217e1e9d0ced"
        },
        "date": 1731014908275,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.9,
            "unit": "ns/op",
            "extra": "2684976 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.6,
            "unit": "ns/op",
            "extra": "2354568 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28247,
            "unit": "ns/op",
            "extra": "42463 times\n4 procs"
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
          "id": "1531008bdec9c034f02131383b62ef80b51aaeec",
          "message": "Add key id field to consensus wrappers (#923)",
          "timestamp": "2024-11-08T10:26:17Z",
          "tree_id": "3e9d71670de0c72d585bd02e51e9ddbace563517",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1531008bdec9c034f02131383b62ef80b51aaeec"
        },
        "date": 1731061631003,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.7,
            "unit": "ns/op",
            "extra": "2546547 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 504.2,
            "unit": "ns/op",
            "extra": "2387876 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28287,
            "unit": "ns/op",
            "extra": "42466 times\n4 procs"
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
          "id": "44ef01dbdeff940b342a8693cbfadb795f4ee1ca",
          "message": "[KS-507] Make Streams Trigger ID (name+version) configurable (#925)",
          "timestamp": "2024-11-08T06:38:08-08:00",
          "tree_id": "e5f2aad53a6bf6eb38e94d51d8068019fc9fd1bb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/44ef01dbdeff940b342a8693cbfadb795f4ee1ca"
        },
        "date": 1731076747858,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.4,
            "unit": "ns/op",
            "extra": "2712001 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.5,
            "unit": "ns/op",
            "extra": "2357679 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28346,
            "unit": "ns/op",
            "extra": "42482 times\n4 procs"
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
          "id": "4ae4553ff99a6e26fff032c0b5dde82fbc357910",
          "message": "Improve getMapsFromPath to handle ptrs to array/slice and map cleanup (#926)\n\n* Improve getMapsFromPath to handle ptrs to array/slice and add a test\r\n\r\n* minor improvement",
          "timestamp": "2024-11-08T18:17:39+01:00",
          "tree_id": "70ab7b7f29d9fabf27bd75073d7006e0e7ce8f36",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4ae4553ff99a6e26fff032c0b5dde82fbc357910"
        },
        "date": 1731086319419,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.4,
            "unit": "ns/op",
            "extra": "2503801 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508.8,
            "unit": "ns/op",
            "extra": "2347791 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28263,
            "unit": "ns/op",
            "extra": "42571 times\n4 procs"
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
          "id": "914b88b62cf277f057d396275a29131a27879057",
          "message": "Beholder CSA Authentication (#877)\n\n* Rename getAttributes to getMap\r\n\r\n* Fix getMap\r\n\r\n* Add Authenticator to Beholder\r\n\r\n* Use Authenticator in Beholder\r\n\r\n* Add Authenticator to Beholder global\r\n\r\n* Use Authenticator Headers in LOOP\r\n\r\n* Add authenticator to HTTP client\r\n\r\n* Fix config test\r\n\r\n* Add pub key getter to authenticator\r\n\r\n* Set CSA pub key on Otel resource\r\n\r\n* Add noop value to authenticator\r\n\r\n* Move auth tests to beholder package, unexport new auth\r\n\r\n* Simplify auth header approach\r\n\r\n* Remove duplicate test\r\n\r\n* Use ed25519 keys instead of signer\r\n\r\n* Remove pub key from args\r\n\r\n---------\r\n\r\nCo-authored-by: nanchano <nicolas.anchano@smartcontract.com>\r\nCo-authored-by: Pavel <177363085+pkcll@users.noreply.github.com>\r\nCo-authored-by: Geert G <117188496+cll-gg@users.noreply.github.com>",
          "timestamp": "2024-11-08T15:43:52-05:00",
          "tree_id": "4cc4684dfe5f72e91c1600388a73f48b63b86aeb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/914b88b62cf277f057d396275a29131a27879057"
        },
        "date": 1731098696208,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 443.4,
            "unit": "ns/op",
            "extra": "2502715 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.4,
            "unit": "ns/op",
            "extra": "2191346 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28235,
            "unit": "ns/op",
            "extra": "42360 times\n4 procs"
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
          "id": "af894848b3b461c607ee310cb488d3e449e6b55c",
          "message": "Enable batching for beholder emitter in LOOP Server (#927)\n\n* Enable batching for beholder emitter in LOOP Server\r\n\r\n* Rename config fields",
          "timestamp": "2024-11-08T19:22:40-05:00",
          "tree_id": "0eab23082134547df14532a18a2420189f47668d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/af894848b3b461c607ee310cb488d3e449e6b55c"
        },
        "date": 1731111817218,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 443.5,
            "unit": "ns/op",
            "extra": "2681552 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526,
            "unit": "ns/op",
            "extra": "2310855 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28257,
            "unit": "ns/op",
            "extra": "41635 times\n4 procs"
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
          "id": "9c172120302b50e5208a639221f498191da0d3c4",
          "message": "[fix] Marshal state deterministically (#928)",
          "timestamp": "2024-11-11T09:46:48Z",
          "tree_id": "26a8c0d4e28be628b25fc1975e82e4d3568f3de4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9c172120302b50e5208a639221f498191da0d3c4"
        },
        "date": 1731318467212,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.7,
            "unit": "ns/op",
            "extra": "2557333 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 505.1,
            "unit": "ns/op",
            "extra": "2377052 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28271,
            "unit": "ns/op",
            "extra": "42429 times\n4 procs"
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
          "id": "d37df61b04c3ecf8fe6881be5b185f208d1516f4",
          "message": "Apply timeout to host functions (#929)",
          "timestamp": "2024-11-11T10:45:04Z",
          "tree_id": "75614b74d2519709af6bdffea185cae4b6f030df",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d37df61b04c3ecf8fe6881be5b185f208d1516f4"
        },
        "date": 1731321957855,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 445.3,
            "unit": "ns/op",
            "extra": "2374700 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.5,
            "unit": "ns/op",
            "extra": "1975903 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42538 times\n4 procs"
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
          "id": "c61aebee0af93db7bf5a951d51724e4f0930049b",
          "message": "Default to version 1.1.0 of Streams Trigger (#932)",
          "timestamp": "2024-11-11T11:46:21-07:00",
          "tree_id": "e91bec43e386e8d6f04a46495cdb80cc65b74aaf",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c61aebee0af93db7bf5a951d51724e4f0930049b"
        },
        "date": 1731350838857,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.6,
            "unit": "ns/op",
            "extra": "2548206 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 505.6,
            "unit": "ns/op",
            "extra": "2374068 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28412,
            "unit": "ns/op",
            "extra": "42486 times\n4 procs"
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
          "id": "5d958d7a8b12e5252ecf93aadbb6ec64074047ae",
          "message": "pkg/services: update docs (#933)",
          "timestamp": "2024-11-12T05:32:25-06:00",
          "tree_id": "951fb7fd060f65b9fe9436e1119c7fe632d54f60",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5d958d7a8b12e5252ecf93aadbb6ec64074047ae"
        },
        "date": 1731411210762,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 494.6,
            "unit": "ns/op",
            "extra": "2593240 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 506.7,
            "unit": "ns/op",
            "extra": "2377080 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28252,
            "unit": "ns/op",
            "extra": "42397 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "pablolagreca@hotmail.com",
            "name": "pablolagreca",
            "username": "pablolagreca"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "0e2daed34ef6738ccce1362f53384460550e5bea",
          "message": "BCFR-967 - Basic support for method writing and reading - Add logic to enable/disable test cases for chain components common test suite (#829)\n\n* BCFR-967 - Basic support for method writing and reading - Add logic to enable/disable test cases for chain components common test suites\r\n\r\n* improving test cases IDs and grouping them\r\n\r\n---------\r\n\r\nCo-authored-by: Pablo La Greca <pablo.lagreca@msartcontract.com>\r\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2024-11-12T11:08:26-03:00",
          "tree_id": "2cd041b68e24fea9c586609bbd0e3ab31a34d54d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0e2daed34ef6738ccce1362f53384460550e5bea"
        },
        "date": 1731420568736,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447,
            "unit": "ns/op",
            "extra": "2686521 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 503.4,
            "unit": "ns/op",
            "extra": "2384342 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28336,
            "unit": "ns/op",
            "extra": "42453 times\n4 procs"
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
          "id": "ad26f6053786d93e0702eef2361062691b3ceb53",
          "message": "add get latest value with head data (#931)\n\n* add get latest value with head data\r\n\r\n* embed get latest value with head data test into get latest value test",
          "timestamp": "2024-11-13T12:11:43Z",
          "tree_id": "81d66334271f865f312dd37ed131911a41b2387d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ad26f6053786d93e0702eef2361062691b3ceb53"
        },
        "date": 1731499970457,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 469,
            "unit": "ns/op",
            "extra": "2705878 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 503.2,
            "unit": "ns/op",
            "extra": "2397190 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28355,
            "unit": "ns/op",
            "extra": "42342 times\n4 procs"
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
          "id": "8a7a997a03710e7a80d883a5aa822cb097bd5492",
          "message": "remove test to unbreak core (#936)",
          "timestamp": "2024-11-13T14:22:56Z",
          "tree_id": "4a8bf7fd411e1a7e33d2b67ac8ede3f0b4af1f92",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8a7a997a03710e7a80d883a5aa822cb097bd5492"
        },
        "date": 1731507838490,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.5,
            "unit": "ns/op",
            "extra": "2428928 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 502,
            "unit": "ns/op",
            "extra": "2369332 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28628,
            "unit": "ns/op",
            "extra": "42361 times\n4 procs"
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
          "id": "cb37b932100911bbfdadab21a8baa907d4636e10",
          "message": "[CAPPL-270/271] Fix Consensus bugs (#934)\n\n- Fix \"result is not a pointer error\" in the reduce aggregator\r\n- Continue rather than error if we encounter an aggregation error",
          "timestamp": "2024-11-13T14:45:57Z",
          "tree_id": "891209760ef4bcd469f4a096fd095a0296d93dfa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cb37b932100911bbfdadab21a8baa907d4636e10"
        },
        "date": 1731509221246,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 446.8,
            "unit": "ns/op",
            "extra": "2700645 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 508,
            "unit": "ns/op",
            "extra": "2329899 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28359,
            "unit": "ns/op",
            "extra": "42344 times\n4 procs"
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
          "id": "cc4f026925aeb98e766366c1d871ae7401a810e5",
          "message": "pkg/services: add *Engine IfStarted & IfNotStopped methods (#935)",
          "timestamp": "2024-11-13T10:14:00-06:00",
          "tree_id": "7797ce479d32a66f919339c6412935233371431a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cc4f026925aeb98e766366c1d871ae7401a810e5"
        },
        "date": 1731514512624,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 445.2,
            "unit": "ns/op",
            "extra": "2706406 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 502.4,
            "unit": "ns/op",
            "extra": "2378544 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28364,
            "unit": "ns/op",
            "extra": "42350 times\n4 procs"
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
          "id": "c220622d97884f7cd02d908ca380ebb398736e32",
          "message": "fix(observability-lib): updating alert rules + deleting associated alerts when deleting dashboard (#937)",
          "timestamp": "2024-11-13T18:39:07+01:00",
          "tree_id": "1d3dc39d4cc882344042e7f4f3ac4b6ea595a0cc",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c220622d97884f7cd02d908ca380ebb398736e32"
        },
        "date": 1731519620883,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 443.4,
            "unit": "ns/op",
            "extra": "2663725 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.6,
            "unit": "ns/op",
            "extra": "2220656 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28359,
            "unit": "ns/op",
            "extra": "42235 times\n4 procs"
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
          "id": "7f3fc0f974fce6009d8743b351c7b7adf3ed55ae",
          "message": "Add 'identical' aggregator to OCR3 Consensus capability json schema options (#940)\n\n* Add 'identical' aggregator to OCR3 Consensus capability json schema options\r\n\r\n* (test): update test schema",
          "timestamp": "2024-11-14T12:11:45Z",
          "tree_id": "c1ddc0c2af6f6f2d6d3ea72cc5fced289f3fc261",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7f3fc0f974fce6009d8743b351c7b7adf3ed55ae"
        },
        "date": 1731586374463,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 441.7,
            "unit": "ns/op",
            "extra": "2618528 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 524.9,
            "unit": "ns/op",
            "extra": "2385081 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28360,
            "unit": "ns/op",
            "extra": "42322 times\n4 procs"
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
          "id": "aadff98ef0688543aae154ba4712484b2df83d82",
          "message": "Changes required for remote action support  (#930)\n\n* reader capability changes\r\n\r\n* update test\r\n\r\n* review comments",
          "timestamp": "2024-11-14T13:48:22Z",
          "tree_id": "de5249653533d0450f36332cc67bf9521ab19b91",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aadff98ef0688543aae154ba4712484b2df83d82"
        },
        "date": 1731592168909,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.6,
            "unit": "ns/op",
            "extra": "2686148 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 497.2,
            "unit": "ns/op",
            "extra": "2401112 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28297,
            "unit": "ns/op",
            "extra": "42411 times\n4 procs"
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
          "id": "9557da03bad32971c6035d8a1bf5ffe7bf536284",
          "message": "Adds beholder logging (#938)",
          "timestamp": "2024-11-14T10:55:34-05:00",
          "tree_id": "b335396794bf3f3b00638c432aa5b1d314b3f21e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9557da03bad32971c6035d8a1bf5ffe7bf536284"
        },
        "date": 1731599802195,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 443.8,
            "unit": "ns/op",
            "extra": "2681691 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 500.5,
            "unit": "ns/op",
            "extra": "2383836 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28274,
            "unit": "ns/op",
            "extra": "42469 times\n4 procs"
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
          "id": "65bdfbc52ccf2e617e5f05f12b8aa0d6f8e6a7d2",
          "message": "[CAPPL-276] update current state only if should report is true (#939)\n\n* fix: update current state only if should report is true\r\n\r\n* fix: refactor initializeCurrentState to not initialize with ZeroValue in order to distinguish between zero and empty value\r\n\r\n* feat: only report DEVIATION_TYPE_NONE if the value has changed\r\n\r\n* feat: add DEVIATION_TYPE_ANY to check for any type of change\r\n\r\n* test: create test table to reduce duplication",
          "timestamp": "2024-11-15T11:29:43Z",
          "tree_id": "c40862f566b5a076a8debf140c0f24630daf72c8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/65bdfbc52ccf2e617e5f05f12b8aa0d6f8e6a7d2"
        },
        "date": 1731670247430,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 462.4,
            "unit": "ns/op",
            "extra": "2487567 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 501.1,
            "unit": "ns/op",
            "extra": "2370848 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28280,
            "unit": "ns/op",
            "extra": "42140 times\n4 procs"
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
          "id": "ad84e3712352552a7c8720860aed7942d01228c8",
          "message": "fix generate diff check (#943)",
          "timestamp": "2024-11-15T06:04:14-06:00",
          "tree_id": "74e878304049bff3fdf0141ca8f663d369348328",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ad84e3712352552a7c8720860aed7942d01228c8"
        },
        "date": 1731672319490,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.9,
            "unit": "ns/op",
            "extra": "2675240 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.4,
            "unit": "ns/op",
            "extra": "2197332 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28350,
            "unit": "ns/op",
            "extra": "42030 times\n4 procs"
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
          "id": "a6a70ec7692bd91bb15a55c99d74344975ca2957",
          "message": "[CAPPL-305] Fix typo in custom compute capability ID (#945)",
          "timestamp": "2024-11-20T11:17:40Z",
          "tree_id": "da4c14073ff0fa1901f72d422041d30e0816c673",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a6a70ec7692bd91bb15a55c99d74344975ca2957"
        },
        "date": 1732101526200,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 444.4,
            "unit": "ns/op",
            "extra": "2687307 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 497.9,
            "unit": "ns/op",
            "extra": "2320359 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28297,
            "unit": "ns/op",
            "extra": "42321 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "domino.valdano@smartcontract.com",
            "name": "Domino Valdano",
            "username": "reductionista"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "262c6d8a55e1bc7bc5f1d8a23ef3b0e8c4d96642",
          "message": "generate a mock DataSource for use in unit testing (#919)\n\n* Add sqltest package with no-op DataSource definition",
          "timestamp": "2024-11-20T11:06:13-08:00",
          "tree_id": "25c14893050fa2082d701559e1618a92adfaa204",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/262c6d8a55e1bc7bc5f1d8a23ef3b0e8c4d96642"
        },
        "date": 1732129632432,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.9,
            "unit": "ns/op",
            "extra": "2663347 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.1,
            "unit": "ns/op",
            "extra": "2043003 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28330,
            "unit": "ns/op",
            "extra": "42436 times\n4 procs"
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
          "id": "e0189e5db1ec1fa675cdacccdae5baaa40239c8a",
          "message": "Pass AuthHeaders from beholder config to to loop/Tracing (#948)",
          "timestamp": "2024-11-21T06:11:06-06:00",
          "tree_id": "2dd8c72de420cdaeb9adcd9777289b03a17f033c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e0189e5db1ec1fa675cdacccdae5baaa40239c8a"
        },
        "date": 1732191125796,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.7,
            "unit": "ns/op",
            "extra": "2667625 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.1,
            "unit": "ns/op",
            "extra": "2271579 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28370,
            "unit": "ns/op",
            "extra": "42462 times\n4 procs"
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
          "id": "97ceadb2072d3f60896922c260ae902ecb2d8c5d",
          "message": "[CAPPL-197/CAPPL-309] Small fixes to compute (#951)\n\n* [CAPPL-197/CAPPL-309] Small fixes to compute\r\n\r\n- Add a recovery handler to the runner's Run method. This means we\r\n  preserve stack traces and will help debugging.\r\n- Allow users to explicitly error with an error via ExitWithError.\r\n- Remove owner and name from the factory constructor.\r\n\r\n* [CAPPL-197/CAPPL-309] Small fixes to compute\r\n\r\n- Add a recovery handler to the runner's Run method. This means we\r\n  preserve stack traces and will help debugging.\r\n- Allow users to explicitly error with an error via ExitWithError.\r\n- Remove owner and name from the factory constructor.",
          "timestamp": "2024-11-25T15:06:08Z",
          "tree_id": "00e2c987d9689985219ed89065e856e39c58a03a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/97ceadb2072d3f60896922c260ae902ecb2d8c5d"
        },
        "date": 1732547238148,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 452.7,
            "unit": "ns/op",
            "extra": "2453799 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.1,
            "unit": "ns/op",
            "extra": "2380863 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28269,
            "unit": "ns/op",
            "extra": "42435 times\n4 procs"
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
          "id": "59c388b419b2d714f8134ba03607414dde8e3128",
          "message": "add QueryKey helper functions (#613)\n\n* add QueryKey helper functions",
          "timestamp": "2024-11-25T16:40:03-05:00",
          "tree_id": "7e3273910498e60d13a710b4cc1020be8a4983ca",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/59c388b419b2d714f8134ba03607414dde8e3128"
        },
        "date": 1732570869296,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.3,
            "unit": "ns/op",
            "extra": "2676764 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.6,
            "unit": "ns/op",
            "extra": "2300148 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28283,
            "unit": "ns/op",
            "extra": "42403 times\n4 procs"
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
          "id": "d8e086da888a5a56ae4d41984cee271eabdf216f",
          "message": "go 1.23 (#952)",
          "timestamp": "2024-11-26T06:57:16-06:00",
          "tree_id": "87eaeef13f6d398d6af966a81b55cd1dcff1bc1b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d8e086da888a5a56ae4d41984cee271eabdf216f"
        },
        "date": 1732625904962,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 447.7,
            "unit": "ns/op",
            "extra": "2665273 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 507,
            "unit": "ns/op",
            "extra": "2285928 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28275,
            "unit": "ns/op",
            "extra": "42511 times\n4 procs"
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
          "id": "1a1df2cf0f6139543868c44b910ec857b454c826",
          "message": "bump golangci-lint 1.62.2 (#954)",
          "timestamp": "2024-11-26T13:43:35-06:00",
          "tree_id": "5c9fa18f557916893a9655cd4c0dbfd1de4068e3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1a1df2cf0f6139543868c44b910ec857b454c826"
        },
        "date": 1732650273685,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.8,
            "unit": "ns/op",
            "extra": "2686425 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.9,
            "unit": "ns/op",
            "extra": "2263554 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28379,
            "unit": "ns/op",
            "extra": "42375 times\n4 procs"
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
          "id": "15b3598dc146c282c6c8dd0330c44ffceb42a3b4",
          "message": "feat(observability-lib): can create standalone alerts with alert group (#950)",
          "timestamp": "2024-11-27T15:59:10+01:00",
          "tree_id": "60dcf3fb2f8c6154c3555eb39e3361608366512b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/15b3598dc146c282c6c8dd0330c44ffceb42a3b4"
        },
        "date": 1732719610428,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.4,
            "unit": "ns/op",
            "extra": "2704198 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 493.5,
            "unit": "ns/op",
            "extra": "2423229 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28005,
            "unit": "ns/op",
            "extra": "41948 times\n4 procs"
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
          "id": "75cf18c4d0c4c3e563827ee1f8d339c001981a0d",
          "message": "feat(observability-lib): notification policy provisioning improvements (#955)",
          "timestamp": "2024-11-27T16:24:56+01:00",
          "tree_id": "109794974c925ddb6f9f18b987506a7cbc957b1a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/75cf18c4d0c4c3e563827ee1f8d339c001981a0d"
        },
        "date": 1732721159878,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 443.2,
            "unit": "ns/op",
            "extra": "2687457 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 496.2,
            "unit": "ns/op",
            "extra": "2341320 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28341,
            "unit": "ns/op",
            "extra": "39368 times\n4 procs"
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
          "id": "07aa781ee1f492bd806386e4ab4b4b0e987fdb96",
          "message": "Rename ChainWriter Chain Component to ContractWriter (#956)\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2024-11-27T17:26:36+01:00",
          "tree_id": "819a0356c66f222eb351d3ce91b905e8ea1aa45a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/07aa781ee1f492bd806386e4ab4b4b0e987fdb96"
        },
        "date": 1732724859109,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465.9,
            "unit": "ns/op",
            "extra": "2698802 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 493.4,
            "unit": "ns/op",
            "extra": "2398128 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28281,
            "unit": "ns/op",
            "extra": "41691 times\n4 procs"
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
          "id": "2c73d505ee33197542d1907b3b86d55e46d93206",
          "message": "(fix): Handle nested types as output of Compute capability (#949)\n\n* (fix): Handle nested types from output of Compute capability\r\n\r\n* Only allow structs input in CreateMapFromStruct\r\n\r\n* Update pkg/values/value.go\r\n\r\nCo-authored-by: Street <5597260+MStreet3@users.noreply.github.com>\r\n\r\n---------\r\n\r\nCo-authored-by: Street <5597260+MStreet3@users.noreply.github.com>",
          "timestamp": "2024-11-28T10:22:50Z",
          "tree_id": "3254efe04c40aae371ab25c458e272c2bcaa41ce",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2c73d505ee33197542d1907b3b86d55e46d93206"
        },
        "date": 1732789433364,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467.4,
            "unit": "ns/op",
            "extra": "2648060 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 496.7,
            "unit": "ns/op",
            "extra": "2395237 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28905,
            "unit": "ns/op",
            "extra": "42444 times\n4 procs"
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
          "id": "26d4a0b45b233961f6148439a000a86006413613",
          "message": "[CAPPL-205] Remove OCR3 capability from registry when closed (#953)\n\n* Add Remove method to capabilities registry interface\r\n\r\n* Remove OCR3 capability plugin from Capability Registry on close\r\n\r\n* Refactor to use context & remove from Registry within OCR3 Capability\r\n\r\n* (refactor): Simplify to passing ID as the input to Remove",
          "timestamp": "2024-12-02T12:24:04-05:00",
          "tree_id": "cae4eb13789295c31e0cbf2d9555dd419dc91267",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/26d4a0b45b233961f6148439a000a86006413613"
        },
        "date": 1733160312704,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.8,
            "unit": "ns/op",
            "extra": "2677824 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.1,
            "unit": "ns/op",
            "extra": "2324020 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28320,
            "unit": "ns/op",
            "extra": "42364 times\n4 procs"
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
          "id": "a946a573d600f6485948b7d8019a6c974386401c",
          "message": "pkg/types/interfacetests: simplify test names (#959)",
          "timestamp": "2024-12-04T07:57:43-06:00",
          "tree_id": "c9f0bbe2c67da6ca8d7a7b5f66aee4fdbebeac09",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a946a573d600f6485948b7d8019a6c974386401c"
        },
        "date": 1733320727894,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.4,
            "unit": "ns/op",
            "extra": "2699366 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 496,
            "unit": "ns/op",
            "extra": "2420175 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28596,
            "unit": "ns/op",
            "extra": "41894 times\n4 procs"
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
          "id": "3d96843f6b7dbc0d2792ad260113742b4f843aeb",
          "message": "pkg/loop/internal/net: treat Close specially from clientConn (#960)",
          "timestamp": "2024-12-04T11:32:18-06:00",
          "tree_id": "1255b3e3fcc8595d9fd6ec12c1e21006ee04ea9f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3d96843f6b7dbc0d2792ad260113742b4f843aeb"
        },
        "date": 1733333642670,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 468.8,
            "unit": "ns/op",
            "extra": "2612439 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.8,
            "unit": "ns/op",
            "extra": "2336398 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28176,
            "unit": "ns/op",
            "extra": "42606 times\n4 procs"
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
          "id": "29871ced7b4de1ccbedb96c0771ef5dfac7c28b8",
          "message": "[chore] Add function to generate workflowID (#962)\n\n* [chore] Add function to generate workflowID\r\n\r\n* [chore] Add function to generate workflowID",
          "timestamp": "2024-12-04T18:45:25Z",
          "tree_id": "47ef8db2e4f7d79ed9681720dea0923e2c966843",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/29871ced7b4de1ccbedb96c0771ef5dfac7c28b8"
        },
        "date": 1733337986986,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460,
            "unit": "ns/op",
            "extra": "2631580 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 510.8,
            "unit": "ns/op",
            "extra": "2338092 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28432,
            "unit": "ns/op",
            "extra": "42603 times\n4 procs"
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
          "id": "b6684ee6508f89d0d3b7d8c8b9fb1b2db93c1e43",
          "message": "Adding View option to Beholder config (#958)\n\n* Adding View option to Beholder config\r\n\r\n* updating to a slice of metric views\r\n\r\n* fixing httm meter provider\r\n\r\n* attempting to fix ExampleConfig test",
          "timestamp": "2024-12-05T20:12:33-05:00",
          "tree_id": "ec1e26493482943a3e3616c717deea5daefbe8a7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b6684ee6508f89d0d3b7d8c8b9fb1b2db93c1e43"
        },
        "date": 1733447610253,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.3,
            "unit": "ns/op",
            "extra": "2341770 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 510,
            "unit": "ns/op",
            "extra": "2293284 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28191,
            "unit": "ns/op",
            "extra": "42580 times\n4 procs"
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
          "id": "70300ddcc77640b62c3dcd49c36b1eaf80844ac4",
          "message": "additional API method to support getting events of different types in index order (#944)",
          "timestamp": "2024-12-09T15:13:52Z",
          "tree_id": "add24a6ddf7586bb044ae0cff0ac1e0f9b880373",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/70300ddcc77640b62c3dcd49c36b1eaf80844ac4"
        },
        "date": 1733757298653,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 479.7,
            "unit": "ns/op",
            "extra": "2468149 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521,
            "unit": "ns/op",
            "extra": "2296135 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28217,
            "unit": "ns/op",
            "extra": "42602 times\n4 procs"
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
          "id": "182a3d1ef5af6c5e9a21bca84d904251847fc315",
          "message": "feat(observability-lib): remove consumers + refactor cmd (#964)",
          "timestamp": "2024-12-09T17:37:49+01:00",
          "tree_id": "415fb43e771fb7905ebac767ec26befb062f7d80",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/182a3d1ef5af6c5e9a21bca84d904251847fc315"
        },
        "date": 1733762331977,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 472,
            "unit": "ns/op",
            "extra": "2512638 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 516.3,
            "unit": "ns/op",
            "extra": "2321942 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28204,
            "unit": "ns/op",
            "extra": "42513 times\n4 procs"
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
          "id": "eb2f2bc67b8f3e8bc0bb3b3eed5f774e356899e6",
          "message": "[Keystone] Increase default OCR phase size limit (#969)",
          "timestamp": "2024-12-10T08:05:42-08:00",
          "tree_id": "ea325bcfb9859493656120d11fd2c0ed5b935bc5",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eb2f2bc67b8f3e8bc0bb3b3eed5f774e356899e6"
        },
        "date": 1733846807678,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.5,
            "unit": "ns/op",
            "extra": "2609665 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.1,
            "unit": "ns/op",
            "extra": "2296021 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28252,
            "unit": "ns/op",
            "extra": "42498 times\n4 procs"
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
          "id": "a9c706f99e83ac0ec0e3508930138e4e06d5b160",
          "message": "[INFOPLAT-1592] Address high CPU utilization when telemetry is enabled (#967)\n\n* [loop/EnvConfig] parse sets TelemetryEmitterBatchProcessor, TelemetryEmitterExportTimeout\r\n\r\n* [beholder/client] BatchProcessor ExportTimeout option is non-zero value\r\n\r\n* [loop/EnvConfig] Use maps.Equal in tests\r\n\r\n---------\r\n\r\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2024-12-10T14:26:53-05:00",
          "tree_id": "541a9c2107d89b528d00ac53cb66137120ebfe57",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a9c706f99e83ac0ec0e3508930138e4e06d5b160"
        },
        "date": 1733858883830,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 449.9,
            "unit": "ns/op",
            "extra": "2680748 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.1,
            "unit": "ns/op",
            "extra": "2346340 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 26890,
            "unit": "ns/op",
            "extra": "44739 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "lei.shi@smartcontract.com",
            "name": "Lei",
            "username": "shileiwill"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9087f5e8daf9a3e693a726839f789e6590e7ce09",
          "message": "add cron trigger and readcontract action (#971)\n\nSigned-off-by: Lei <lei.shi@smartcontract.com>",
          "timestamp": "2024-12-11T19:06:13Z",
          "tree_id": "ca5d4ebe6d26f5e6a1201b73d2599c17be74ed60",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9087f5e8daf9a3e693a726839f789e6590e7ce09"
        },
        "date": 1733944038899,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 475.2,
            "unit": "ns/op",
            "extra": "2427742 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 516.9,
            "unit": "ns/op",
            "extra": "2311503 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28312,
            "unit": "ns/op",
            "extra": "40534 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "34754799+dhaidashenko@users.noreply.github.com",
            "name": "Dmytro Haidashenko",
            "username": "dhaidashenko"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "0b03fa331a49577ad30b8b780e0bc8070bd58328",
          "message": "BCFR-1086 finality violation (#966)\n\n* define finality violation error\r\n\r\nSigned-off-by: Dmytro Haidashenko <dmytro.haidashenko@smartcontract.com>\r\n\r\n* rename finality violation\r\n\r\nSigned-off-by: Dmytro Haidashenko <dmytro.haidashenko@smartcontract.com>\r\n\r\n* Test ContainsError\r\n\r\nSigned-off-by: Dmytro Haidashenko <dmytro.haidashenko@smartcontract.com>\r\n\r\n---------\r\n\r\nSigned-off-by: Dmytro Haidashenko <dmytro.haidashenko@smartcontract.com>\r\nCo-authored-by: Domino Valdano <domino.valdano@smartcontract.com>",
          "timestamp": "2024-12-11T20:22:25+01:00",
          "tree_id": "303c1daeb26f62b20d8d895053c20bbb712c7ae8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0b03fa331a49577ad30b8b780e0bc8070bd58328"
        },
        "date": 1733945008868,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.4,
            "unit": "ns/op",
            "extra": "2590290 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.9,
            "unit": "ns/op",
            "extra": "2302418 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28238,
            "unit": "ns/op",
            "extra": "42352 times\n4 procs"
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
          "id": "525a5610c8775f1566802ddec651f1383e155df1",
          "message": "[CAPPL] Add mode quorum configuration option to Reduce Aggregator (#972)\n\n* Add 'majority' aggregation method to Reduce Aggregator\r\n\r\n* (refactor): Change implementation to 'ModeQuorum'\r\n\r\n* Only fill modeQuorum for method mode",
          "timestamp": "2024-12-12T08:22:05-08:00",
          "tree_id": "fc7132d00b4f277fe50047facd398a60a53df3bf",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/525a5610c8775f1566802ddec651f1383e155df1"
        },
        "date": 1734020592515,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458,
            "unit": "ns/op",
            "extra": "2482362 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.5,
            "unit": "ns/op",
            "extra": "2320495 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28250,
            "unit": "ns/op",
            "extra": "42482 times\n4 procs"
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
          "id": "6a43e61b9d4990e98ca80a8155cfa5287c5d67b6",
          "message": "[CAPPL-366/CAPPL-382] Miscellaneous fixes (#973)\n\n* [CAPPL-382] Normalize owner before comparing\r\n\r\n* [CAPPL-366] Add name to hash to generate workflowID",
          "timestamp": "2024-12-12T16:39:58Z",
          "tree_id": "b4f5d564555cdbbde29b5db43ae012ff27c015c6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6a43e61b9d4990e98ca80a8155cfa5287c5d67b6"
        },
        "date": 1734021663178,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467.9,
            "unit": "ns/op",
            "extra": "2613801 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.6,
            "unit": "ns/op",
            "extra": "2311908 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28376,
            "unit": "ns/op",
            "extra": "42555 times\n4 procs"
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
          "id": "dbebc0fc753a6cb6955fb08e9d2f53d8e401ed24",
          "message": "(feat): Add PreCodec modifier (#961)",
          "timestamp": "2024-12-13T17:49:39Z",
          "tree_id": "2f1a7c1e449576581af6c39f29b440c088120529",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/dbebc0fc753a6cb6955fb08e9d2f53d8e401ed24"
        },
        "date": 1734112242766,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457,
            "unit": "ns/op",
            "extra": "2604240 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 513.1,
            "unit": "ns/op",
            "extra": "2324736 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28240,
            "unit": "ns/op",
            "extra": "42519 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "domino.valdano@smartcontract.com",
            "name": "Domino Valdano",
            "username": "reductionista"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "edc5deed9ffd87fd980b153e8297660f8b541746",
          "message": "Add pkg/pg with dialects.go & txdb.go (#910)\n\n* Add pkg/pg with dialects.go & txdb.go\r\n\r\nNeither of these were in the actual pg package in chainlink repo.\r\ndialects.go came from core/store/dialects and txdb.go from\r\ncore/internal/testutils/pgtest, but neither of these seem like they\r\ndeserve their own package in chainlink-common--we can lump all the\r\npostgres specific common utilities under pkg/pg\r\n\r\n* Add TestTxDBDriver, NewSqlxDB, SkipShort, SkipShortDB and SkipFlakey\r\n\r\n* Add idempotency test of RegisterTxDb\r\n\r\n* Create ctx from testing context, instead of using context.Background\r\n\r\n* Only abort tx's when last connection is closed\r\n\r\nAlso: convert rest of panic isn't ordinary errors\r\n\r\n* go mod tidy\r\n\r\n* Split abort channel into one per connection object\r\n\r\nAll txdb connections share the same underlying connection to the\r\npostgres db. Calling NewSqlxDB() or NewConnection() with dialect=txdb\r\ndoesn't create a new pg connection, it just creates a new tx with\r\nBEGIN. Closing the connection with db.Close() issues ROLLBACK.\r\n\r\nBoth NewSqlxDB() and NewConneciton() choose random UUID's for their\r\ndsn string, so we shouldn't have a case where the same dsn is opened\r\nmore than once. If that did happen, then these two different txdb\r\n\"connections\" would be sharing the same transaction which would\r\nmean closing the abort channel due to a query sent over one of them\r\nwould affect the other. Hopefully that's not a problem? If it is\r\nI think our only option will be to go back to using context.Background\r\nfor all queries.\r\n\r\nBefore this commit, there was only one abort channel for the entire\r\ntxdb driver meaning that even two entirely different connections\r\nopened with different dsn's could interfere with each other's queries.\r\nThis should fix that case, which is presumably the only case we\r\ncare about. Since each dsn corresponds to a different call to\r\nNewSqlxDB() and the UUID's are generated randomly, there should no\r\nlonger be a conflict. Each txdb connection will have its own abort\r\nchannel.\r\n\r\n* Errorf -> Fatalf on failure to register txdb driver\r\n\r\n* Add in-memory DataSource using go-duckdb\r\n\r\n* Fall back to testing txdb with in-memory backed db if CL_DATABASE_URL is not set\r\n\r\nThis allows us to test most of it in CI, and all locally\r\n\r\n* Fix imports & fmt.Sprintf -> t.Log\r\n\r\n* Add concurrency test for RegisterTxDb()\r\n\r\n* Fix race condition\r\n\r\nThis showed up in some of the unit tests in the linked PR in chainlink repo\r\n\r\n* Remove pg.SkipDB(), add DbUrlOrInMemory()\r\n\r\n* pkg/pg -> pkg/sqlutil/pg\r\n\r\n* NewSqlxDB -> NewTestDB, DbUrlOrInMemory -> TestURL",
          "timestamp": "2024-12-13T13:05:22-08:00",
          "tree_id": "a50cda0f992b0387c03ff3c908aeaf00dee3ab94",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/edc5deed9ffd87fd980b153e8297660f8b541746"
        },
        "date": 1734124036620,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.3,
            "unit": "ns/op",
            "extra": "2638141 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526.5,
            "unit": "ns/op",
            "extra": "2340086 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28392,
            "unit": "ns/op",
            "extra": "42552 times\n4 procs"
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
          "id": "b403079b28054659d66944a44e6d7bae1fb662dc",
          "message": "(fix): Allow pointers to bytes in PreCodec modifier (#975)",
          "timestamp": "2024-12-14T10:58:18-05:00",
          "tree_id": "f0cffe4413cc8ee87eec9b04185bcfe7e7d4b5ca",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b403079b28054659d66944a44e6d7bae1fb662dc"
        },
        "date": 1734191950907,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 464.5,
            "unit": "ns/op",
            "extra": "2505398 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 523.4,
            "unit": "ns/op",
            "extra": "2309174 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28397,
            "unit": "ns/op",
            "extra": "42496 times\n4 procs"
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
          "id": "bbe318cd07609546b26b84bf3f43d622d2e0ea0c",
          "message": "add registration refresh and expiry to executable capability remote config (#968)\n\n* wip\r\n\r\n* tests",
          "timestamp": "2024-12-17T12:09:18Z",
          "tree_id": "94e90c1cc695b17aa8285e4856b7927e38e43885",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bbe318cd07609546b26b84bf3f43d622d2e0ea0c"
        },
        "date": 1734437428464,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 456.5,
            "unit": "ns/op",
            "extra": "2624692 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 537.5,
            "unit": "ns/op",
            "extra": "2333137 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28835,
            "unit": "ns/op",
            "extra": "41590 times\n4 procs"
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
          "id": "9728444fab6273123d4f1f59a917956a68abbc62",
          "message": "Fix `padWorkflowName()` (#977)\n\n* Fixes `padWorkflowName()`\r\n\r\n* Fixes `padWorkflowName()`\r\n\r\n* Fixes `padWorkflowName()`\r\n\r\n* Updates comments on `Metadata` struct\r\n\r\n* Fixes tests",
          "timestamp": "2024-12-19T15:57:27-04:00",
          "tree_id": "ff39d8e2fb914e41237298714ab0c8483c57b354",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9728444fab6273123d4f1f59a917956a68abbc62"
        },
        "date": 1734638317410,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.5,
            "unit": "ns/op",
            "extra": "2616327 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.4,
            "unit": "ns/op",
            "extra": "2301211 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28321,
            "unit": "ns/op",
            "extra": "42109 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "studentcuza@gmail.com",
            "name": "Gheorghe Strimtu",
            "username": "gheorghestrimtu"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7c7d06f0c7e2c160b9b7f2af994595f33e2fa42e",
          "message": "INFOPLAT-1539  Beholder Log Batch Processor More Settings (#957)\n\n* Beholder Log Batch Processor More Settings\r\n\r\n* add settings for message emitter\r\n\r\n* remove settings\r\n\r\n* add settings\r\n\r\n* more settings\r\n\r\n* add too httpclient and test\r\n\r\n* fix ExampleConfig test\r\n\r\n* add new lines for spacing in Config\r\n\r\n* Add all beholder config options to loop/EnvConfig; set beholder config options from loop EnvConfig\r\n\r\n* Set EmitterExportTimeout, LogExportTimeout to 30sec which is OTel default\r\n\r\n* Update comment for EmitterBatchProcessor config option\r\n\r\n* Dont set batch processor options with invalid values\r\n\r\n---------\r\n\r\nCo-authored-by: Pavel <177363085+pkcll@users.noreply.github.com>",
          "timestamp": "2024-12-20T06:33:42-06:00",
          "tree_id": "f84498367a8459e6f0898a3066cd6138bc30783b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7c7d06f0c7e2c160b9b7f2af994595f33e2fa42e"
        },
        "date": 1734698083738,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 472.8,
            "unit": "ns/op",
            "extra": "2596122 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521.3,
            "unit": "ns/op",
            "extra": "2307481 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28209,
            "unit": "ns/op",
            "extra": "42553 times\n4 procs"
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
          "id": "41f4bc066dcd996f03a59e6c1cfb94370a703fa8",
          "message": "feat(workflows): adds workflow name normalizer (#980)",
          "timestamp": "2024-12-20T12:50:45-05:00",
          "tree_id": "abe7e41d40a91c7154fc521c97354dbe41c87401",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/41f4bc066dcd996f03a59e6c1cfb94370a703fa8"
        },
        "date": 1734717108074,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.3,
            "unit": "ns/op",
            "extra": "2625446 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.4,
            "unit": "ns/op",
            "extra": "2313738 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28393,
            "unit": "ns/op",
            "extra": "42546 times\n4 procs"
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
          "id": "db7919d60550c76b37a0d7cc5e694a272cc54bdd",
          "message": "Removed flakey testcases and optimized for parallel test runs (#965)\n\n* Removed Finality Checks that expect errors in ChainComponents tests\"\r\n\r\n* Removed flakey testcases and optimized for parallel test runs\r\n\r\n* Updated new tests",
          "timestamp": "2024-12-23T09:39:29-05:00",
          "tree_id": "f09527d2e5047c20fd89263381fc79bc840c1fbe",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/db7919d60550c76b37a0d7cc5e694a272cc54bdd"
        },
        "date": 1734964840551,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 474.5,
            "unit": "ns/op",
            "extra": "2612947 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.3,
            "unit": "ns/op",
            "extra": "2309457 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28443,
            "unit": "ns/op",
            "extra": "42368 times\n4 procs"
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
          "id": "47a52b179fe3b35c66105acc806bca160bc8ac8c",
          "message": "feat(observability-lib): can specify tooltip on timeseries panels and enable by default (#982)",
          "timestamp": "2025-01-08T13:36:52+01:00",
          "tree_id": "93f262e32abde71963fde4d25f2275afbf94662d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/47a52b179fe3b35c66105acc806bca160bc8ac8c"
        },
        "date": 1736339941762,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.3,
            "unit": "ns/op",
            "extra": "2608291 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521.3,
            "unit": "ns/op",
            "extra": "2268507 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28201,
            "unit": "ns/op",
            "extra": "42502 times\n4 procs"
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
          "id": "2ebd63bbb16ec1bb0f0c5b7263a18411a849201a",
          "message": "[CAPPL-320] implement HexDecodeWorkflowName (#983)",
          "timestamp": "2025-01-08T14:43:20-05:00",
          "tree_id": "5d3b92a0b4cf1fb87abc7c9300a047500c12ba14",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2ebd63bbb16ec1bb0f0c5b7263a18411a849201a"
        },
        "date": 1736365463013,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.3,
            "unit": "ns/op",
            "extra": "2604100 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 532.5,
            "unit": "ns/op",
            "extra": "2306035 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28634,
            "unit": "ns/op",
            "extra": "42488 times\n4 procs"
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
          "id": "c2007b3df1b680db993a9fae43ce8583f1e20921",
          "message": "fix: newTimeout should be read as Uint64 (#987)",
          "timestamp": "2025-01-09T08:07:30-08:00",
          "tree_id": "80b60bf027e14e4fb0636e6f9d74c1d0547ac3dc",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c2007b3df1b680db993a9fae43ce8583f1e20921"
        },
        "date": 1736438913746,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.8,
            "unit": "ns/op",
            "extra": "2518530 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 542.6,
            "unit": "ns/op",
            "extra": "2297578 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28308,
            "unit": "ns/op",
            "extra": "42296 times\n4 procs"
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
          "id": "0f838d55ed83b2d8efd03c01a7fda06e0d036d49",
          "message": "Add ReportFormatEVMAbiEncodeUnpacked (#991)",
          "timestamp": "2025-01-09T12:19:04-05:00",
          "tree_id": "9db2722ed1130a0645cd9f4568c5be28b90c78d3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0f838d55ed83b2d8efd03c01a7fda06e0d036d49"
        },
        "date": 1736443204571,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.3,
            "unit": "ns/op",
            "extra": "2595044 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.9,
            "unit": "ns/op",
            "extra": "2274769 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28252,
            "unit": "ns/op",
            "extra": "42538 times\n4 procs"
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
          "id": "7dbb1b0863a38a649b7e049a89a2033ccc4588cd",
          "message": "[CRE-42] Fix partial or truncated writes (#989)\n\n* fix: check size and len(src) match to avoid partial or truncated writes\r\n\r\n* fix: return the number of bytes copied\r\n\r\n* chore: align test naming",
          "timestamp": "2025-01-10T10:10:35Z",
          "tree_id": "09e48701d7c9cf304007a09d2dd8eeaf6cbf18e2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7dbb1b0863a38a649b7e049a89a2033ccc4588cd"
        },
        "date": 1736503900867,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467.4,
            "unit": "ns/op",
            "extra": "2530446 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.1,
            "unit": "ns/op",
            "extra": "2185737 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28203,
            "unit": "ns/op",
            "extra": "42430 times\n4 procs"
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
          "id": "149f0847b70b8a25fb0fc3f1ada94e59325d6478",
          "message": "pkg/logger: docs (#985)",
          "timestamp": "2025-01-10T10:14:21-06:00",
          "tree_id": "2648463700241559ac01c62ee12bbd0e84f100de",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/149f0847b70b8a25fb0fc3f1ada94e59325d6478"
        },
        "date": 1736525725187,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 466.9,
            "unit": "ns/op",
            "extra": "2564686 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 531.3,
            "unit": "ns/op",
            "extra": "2260130 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28332,
            "unit": "ns/op",
            "extra": "42523 times\n4 procs"
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
          "id": "9b2f9ef755857edbee5a8187ef1bf41aaa7cbc33",
          "message": "feat(observability-lib): can specify max data points on panels (#981)",
          "timestamp": "2025-01-10T21:17:31+01:00",
          "tree_id": "f4c2b9985f8320b4dd1661bc7a72d1cb617671ac",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9b2f9ef755857edbee5a8187ef1bf41aaa7cbc33"
        },
        "date": 1736540311555,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.9,
            "unit": "ns/op",
            "extra": "2591137 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 525.3,
            "unit": "ns/op",
            "extra": "2288138 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28259,
            "unit": "ns/op",
            "extra": "42339 times\n4 procs"
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
          "id": "7c712f12dc6a7f46afae5947e5b233038e1966a5",
          "message": "feat(observability-lib): add timerange to alert rule (#979)",
          "timestamp": "2025-01-10T21:31:57+01:00",
          "tree_id": "57f2be320c7a76a035d6b7c7eeddb69285cc58e0",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7c712f12dc6a7f46afae5947e5b233038e1966a5"
        },
        "date": 1736541180370,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.8,
            "unit": "ns/op",
            "extra": "2588624 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526,
            "unit": "ns/op",
            "extra": "2294354 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28234,
            "unit": "ns/op",
            "extra": "42529 times\n4 procs"
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
          "id": "b34bea64641c9c5e336f0232683de2fc731d8b18",
          "message": "[CRE-43] fix slicing of events (#992)\n\n* fix: calculate the index of the slot instead of relying on the value of it\r\n\r\n* test: extract getSlot and unit test it\r\n\r\n* chore: renename offset + add test coverage",
          "timestamp": "2025-01-13T10:48:49Z",
          "tree_id": "93da6bcfa5dcfc0c3bd0de27a4f7b0c3861c9c68",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b34bea64641c9c5e336f0232683de2fc731d8b18"
        },
        "date": 1736765402666,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 504.9,
            "unit": "ns/op",
            "extra": "2370367 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 570.8,
            "unit": "ns/op",
            "extra": "2119292 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29534,
            "unit": "ns/op",
            "extra": "40131 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "lee.yikjiun@gmail.com",
            "name": "Lee Yik Jiun",
            "username": "leeyikjiun"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "373e8891c5abe758d640459bad2486c54f68f8dd",
          "message": "feat(observability-lib): various updates to observabilty library (#993)\n\n* Various updates to observabilty library\r\n\r\n- Add text panel\r\n- Make Title and Decimals in panel nullable\r\n- Add LineWidth and DrawStyle to time series panel\r\n- Add more configs to log panel\r\n- Add Hide to custom variable\r\n\r\n* Add timezone and tooltip to dashboard builder\r\n\r\n* Fix tests",
          "timestamp": "2025-01-13T18:03:29+01:00",
          "tree_id": "02d1f617d595a322a239815cd1fcdb39039850d6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/373e8891c5abe758d640459bad2486c54f68f8dd"
        },
        "date": 1736787880001,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.2,
            "unit": "ns/op",
            "extra": "2601374 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.1,
            "unit": "ns/op",
            "extra": "2294971 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28661,
            "unit": "ns/op",
            "extra": "42472 times\n4 procs"
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
          "id": "d266596f156041499402862de51ef56ad04c20c9",
          "message": "Serialization for ReportFormatEVMABIEncodeUnpacked (#995)",
          "timestamp": "2025-01-13T12:18:35-05:00",
          "tree_id": "33259b51be42c8144a01504e92bc11081d458928",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d266596f156041499402862de51ef56ad04c20c9"
        },
        "date": 1736788780567,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 458.1,
            "unit": "ns/op",
            "extra": "2589550 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.8,
            "unit": "ns/op",
            "extra": "2317011 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28232,
            "unit": "ns/op",
            "extra": "42510 times\n4 procs"
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
          "id": "42c3764c171e870bfd91443c6ae82a6e76bc6f1f",
          "message": "Add hex encoding to HashTruncateName util (#996)",
          "timestamp": "2025-01-13T10:34:10-08:00",
          "tree_id": "4d0f6e5374200361835b872288e4d6b49f6615f8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/42c3764c171e870bfd91443c6ae82a6e76bc6f1f"
        },
        "date": 1736793320150,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457.9,
            "unit": "ns/op",
            "extra": "2593540 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 522.9,
            "unit": "ns/op",
            "extra": "2310429 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28211,
            "unit": "ns/op",
            "extra": "37977 times\n4 procs"
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
          "id": "0cd7b49eb4786c8c0253a280a412eb94941626d4",
          "message": "[CRE-40] Check binary size before decompression (#994)\n\nfix: check binary size before decompression",
          "timestamp": "2025-01-14T15:11:39+01:00",
          "tree_id": "cec210e4c17a656fd32dd643e175a8b6ee735cbe",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0cd7b49eb4786c8c0253a280a412eb94941626d4"
        },
        "date": 1736863960644,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 475,
            "unit": "ns/op",
            "extra": "2583445 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 541.4,
            "unit": "ns/op",
            "extra": "2301604 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28909,
            "unit": "ns/op",
            "extra": "40776 times\n4 procs"
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
          "id": "4e61572bb9bdfd020ff85cafe6ae480da72f02c4",
          "message": "pkg/services: ErrorBuffer.Flush fix race (#998)",
          "timestamp": "2025-01-15T10:43:25+01:00",
          "tree_id": "39505132317629a658f85ced27f3d1852b1a5984",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4e61572bb9bdfd020ff85cafe6ae480da72f02c4"
        },
        "date": 1736934276633,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459,
            "unit": "ns/op",
            "extra": "2635430 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 529.6,
            "unit": "ns/op",
            "extra": "2301537 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28241,
            "unit": "ns/op",
            "extra": "42531 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "104409744+vreff@users.noreply.github.com",
            "name": "Chris Cushman",
            "username": "vreff"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e56b78c794ecb76b99aedb12fb64052afead8350",
          "message": "Add AggregationConfig (#988)",
          "timestamp": "2025-01-15T15:19:50Z",
          "tree_id": "2bc55b3130ae776cd985777698ed7fdf5193c730",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e56b78c794ecb76b99aedb12fb64052afead8350"
        },
        "date": 1736954460824,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460,
            "unit": "ns/op",
            "extra": "2267336 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 518.9,
            "unit": "ns/op",
            "extra": "2296142 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28232,
            "unit": "ns/op",
            "extra": "42495 times\n4 procs"
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
          "id": "5ef3235a3dc961892f0c88fa303f0881e0cdd15e",
          "message": "[chore] Add README documentation (#999)",
          "timestamp": "2025-01-16T11:57:19Z",
          "tree_id": "88491396536eedbde03af24e29a2af48c49eede9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5ef3235a3dc961892f0c88fa303f0881e0cdd15e"
        },
        "date": 1737028710931,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 459.7,
            "unit": "ns/op",
            "extra": "2616303 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 527.6,
            "unit": "ns/op",
            "extra": "2303362 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28604,
            "unit": "ns/op",
            "extra": "42553 times\n4 procs"
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
          "id": "8481a75ca8a94666851aecdb3e0e768f2012fd31",
          "message": "[CRE-39] (fix): Add more guards & nil checks to WASM compute (#984)\n\n* (fix): Add guards\n\n* (test): Add sdk unit tests\n\n* (test): fix expected error string\n\n* Replace with * instead of space\n\n* Allow spaces in log sanitization\n\n* Split out log sanitization fix",
          "timestamp": "2025-01-16T18:09:10Z",
          "tree_id": "331dbe7be19549d7334b9908b64ac2a5d39dcc4a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8481a75ca8a94666851aecdb3e0e768f2012fd31"
        },
        "date": 1737051022943,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 461.7,
            "unit": "ns/op",
            "extra": "2600439 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 521.8,
            "unit": "ns/op",
            "extra": "2315740 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28259,
            "unit": "ns/op",
            "extra": "41853 times\n4 procs"
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
          "id": "fe3ec4466fb5adfffd8fc77eef1cef67c4a918cc",
          "message": "[CAPPL-471] Handle possible nil versioned bytes (#1002)",
          "timestamp": "2025-01-16T19:30:07Z",
          "tree_id": "681d96139e50cbc635d2afa33ef615c88850f425",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/fe3ec4466fb5adfffd8fc77eef1cef67c4a918cc"
        },
        "date": 1737055870739,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 462.9,
            "unit": "ns/op",
            "extra": "2579654 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 534.7,
            "unit": "ns/op",
            "extra": "2289985 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28213,
            "unit": "ns/op",
            "extra": "42524 times\n4 procs"
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
          "id": "f49c5c27db51b1ec116cd8b4acad5cd269446e2c",
          "message": "pkg/loop: plugins report health to host [BCF-2709] (#206)\n\n* pkg/services/servicetest: add helper for testing HealthReport names\r\n\r\n* pkg/loop: plugins report health to host",
          "timestamp": "2025-01-16T15:48:55-06:00",
          "tree_id": "c44a803eac2d03945c1c8e90c0c4259c549a06cb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f49c5c27db51b1ec116cd8b4acad5cd269446e2c"
        },
        "date": 1737064196483,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 465,
            "unit": "ns/op",
            "extra": "2590003 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 529,
            "unit": "ns/op",
            "extra": "2261188 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28173,
            "unit": "ns/op",
            "extra": "41923 times\n4 procs"
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
          "id": "1922eef0bdd4bdb60669f64f0a41739fe89fe83c",
          "message": "[CRE-44] Add restricted config; validate WASM config (#1001)\n\n* [chore] Add README documentation\n\n* [CRE-44] Add restricted_config and restricted_keys to capability registry config\n\n* Use uint64 to describe min/max memory",
          "timestamp": "2025-01-17T10:15:54Z",
          "tree_id": "91c4651b8be123a9d8e53c20c8bae586fbaf5313",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1922eef0bdd4bdb60669f64f0a41739fe89fe83c"
        },
        "date": 1737109030535,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 460.4,
            "unit": "ns/op",
            "extra": "2640570 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 517.4,
            "unit": "ns/op",
            "extra": "2138074 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28301,
            "unit": "ns/op",
            "extra": "42361 times\n4 procs"
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
          "id": "2b05726309228e588596c9c9370d067d863fc39c",
          "message": "[CAPPL-471] Add more tests to verify that panic is handled (#1003)",
          "timestamp": "2025-01-17T14:44:51Z",
          "tree_id": "a23bce3349913f1872b50a0845854a04c3eec6ee",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2b05726309228e588596c9c9370d067d863fc39c"
        },
        "date": 1737125154903,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 527.4,
            "unit": "ns/op",
            "extra": "2563680 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.9,
            "unit": "ns/op",
            "extra": "2293833 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28233,
            "unit": "ns/op",
            "extra": "42501 times\n4 procs"
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
          "id": "62443f4b3c303f35631461f64e6de4790a21ba30",
          "message": "pkg/sqlutil/pg: create package; expand env config (#450)",
          "timestamp": "2025-01-21T08:19:17-06:00",
          "tree_id": "7799e1610c62e4f41f52ddeaac63febf7149fc25",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/62443f4b3c303f35631461f64e6de4790a21ba30"
        },
        "date": 1737469287480,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.5,
            "unit": "ns/op",
            "extra": "2611502 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519,
            "unit": "ns/op",
            "extra": "2331732 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28214,
            "unit": "ns/op",
            "extra": "42564 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "34754799+dhaidashenko@users.noreply.github.com",
            "name": "Dmytro Haidashenko",
            "username": "dhaidashenko"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3e179a73cb92553b19b0652ebe82b1e15f2d7c23",
          "message": "Query Primitives Any operator (#1004)",
          "timestamp": "2025-01-21T17:33:09+01:00",
          "tree_id": "8932a199fed3e9983548aacb2c0d51caa8e0ace9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3e179a73cb92553b19b0652ebe82b1e15f2d7c23"
        },
        "date": 1737477267637,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.1,
            "unit": "ns/op",
            "extra": "2611695 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 526.5,
            "unit": "ns/op",
            "extra": "2342374 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28178,
            "unit": "ns/op",
            "extra": "42548 times\n4 procs"
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
          "id": "95c9b2dcf46a9a869b201ec22ff96f6ab0c24872",
          "message": "Compute log sanitization (#1000)",
          "timestamp": "2025-01-21T11:11:34-08:00",
          "tree_id": "c19117a5b4c4d43919b30c258d93c54207b8b101",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/95c9b2dcf46a9a869b201ec22ff96f6ab0c24872"
        },
        "date": 1737486772706,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 467,
            "unit": "ns/op",
            "extra": "2649210 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 524.2,
            "unit": "ns/op",
            "extra": "2351820 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28197,
            "unit": "ns/op",
            "extra": "42634 times\n4 procs"
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
          "id": "42d2956d3284e019793f8cc05562fc5045b7ed1a",
          "message": "[fix] Update sanitization regex (#1007)",
          "timestamp": "2025-01-22T13:48:06Z",
          "tree_id": "6bfc98b610213cc577923e028eb9a694fa1c5cae",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/42d2956d3284e019793f8cc05562fc5045b7ed1a"
        },
        "date": 1737553765540,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 478.4,
            "unit": "ns/op",
            "extra": "2619588 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.5,
            "unit": "ns/op",
            "extra": "2351211 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28254,
            "unit": "ns/op",
            "extra": "42606 times\n4 procs"
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
          "id": "bcaa629eba00813508350d442d3acdd81c408a8b",
          "message": "bump golangci-lint v1.63.4; replace deprecated linter (#1006)",
          "timestamp": "2025-01-22T18:44:12+01:00",
          "tree_id": "966571ecaf9ddd2384b6daea8e50bb64f7c6ae8b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bcaa629eba00813508350d442d3acdd81c408a8b"
        },
        "date": 1737567933454,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.2,
            "unit": "ns/op",
            "extra": "2673404 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 510.3,
            "unit": "ns/op",
            "extra": "2358360 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28459,
            "unit": "ns/op",
            "extra": "42548 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "fergal.gribben@smartcontract.com",
            "name": "Fergal",
            "username": "ferglor"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a36a3f72228943aab811ad4f284c8131a9c447f9",
          "message": "Pass a runtime to the test runner (#1010)",
          "timestamp": "2025-01-24T16:31:22Z",
          "tree_id": "755cdd4807a09bf2b5a6374dfcb2f9c924b8384f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a36a3f72228943aab811ad4f284c8131a9c447f9"
        },
        "date": 1737736350915,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.1,
            "unit": "ns/op",
            "extra": "2652991 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.3,
            "unit": "ns/op",
            "extra": "2331381 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28334,
            "unit": "ns/op",
            "extra": "42116 times\n4 procs"
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
          "id": "94af7d0df176038687ca2274aeaf9fa4ca5a4c0a",
          "message": "fix(observability-lib): getAlertRules by group name (#1008)",
          "timestamp": "2025-01-27T13:25:17+01:00",
          "tree_id": "a0138cb466e4a4719856ed5051c6d4630b5ac61e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/94af7d0df176038687ca2274aeaf9fa4ca5a4c0a"
        },
        "date": 1737980782158,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.5,
            "unit": "ns/op",
            "extra": "2650617 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 514.4,
            "unit": "ns/op",
            "extra": "2342968 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28441,
            "unit": "ns/op",
            "extra": "42537 times\n4 procs"
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
          "id": "a8fa42cc0f360a7db51054fea0a284b5d2f8ac51",
          "message": "switch off native unwind info (requires wasmtime version update) (#1012)",
          "timestamp": "2025-01-27T12:55:41Z",
          "tree_id": "f56b3768777d0fab3f61ca2f3a890b6aa2fb9204",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a8fa42cc0f360a7db51054fea0a284b5d2f8ac51"
        },
        "date": 1737982684307,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.8,
            "unit": "ns/op",
            "extra": "2628187 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 516.7,
            "unit": "ns/op",
            "extra": "2325948 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28176,
            "unit": "ns/op",
            "extra": "42590 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "fergal.gribben@smartcontract.com",
            "name": "Fergal",
            "username": "ferglor"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b32b200b4c35630c680403c0be8d5bd303713ded",
          "message": "testutils: Convert the workflow spec to and from proto in ensureGraph (#1014)\n\n* Convert the workflow spec to and from proto\n\n* Add comment",
          "timestamp": "2025-01-28T11:38:11-08:00",
          "tree_id": "2dd72038f272bc89d172265af994879c397eef55",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b32b200b4c35630c680403c0be8d5bd303713ded"
        },
        "date": 1738093149128,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 455.7,
            "unit": "ns/op",
            "extra": "2227364 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.3,
            "unit": "ns/op",
            "extra": "2354854 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28230,
            "unit": "ns/op",
            "extra": "42496 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "david@makewhat.is",
            "name": "David Johansen",
            "username": "makewhatis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "bcca537b302e4863fc95825a3492c37124af992f",
          "message": "Add support for BarGauge panels (#1015)\n\n* add support for bargauge",
          "timestamp": "2025-01-29T16:56:04+01:00",
          "tree_id": "7bc3ef0c93571ef19c2933b1fe8fc88d856d8141",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bcca537b302e4863fc95825a3492c37124af992f"
        },
        "date": 1738166217967,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453,
            "unit": "ns/op",
            "extra": "2493921 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 509.6,
            "unit": "ns/op",
            "extra": "2344362 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28179,
            "unit": "ns/op",
            "extra": "42436 times\n4 procs"
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
          "id": "82e554262f7da72a87944c0b062e47396efea03a",
          "message": "[CAPPL-499] validate decompressed binary size (#1017)\n\nfeat: validate decompresed binary size",
          "timestamp": "2025-01-30T11:46:13+01:00",
          "tree_id": "3152ead216479248592ec1f1a6a1f01359622de0",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/82e554262f7da72a87944c0b062e47396efea03a"
        },
        "date": 1738234024481,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.1,
            "unit": "ns/op",
            "extra": "2619811 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 510,
            "unit": "ns/op",
            "extra": "2258566 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28229,
            "unit": "ns/op",
            "extra": "42547 times\n4 procs"
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
          "id": "64eba2d8808a28003742c6a84d457346528859e5",
          "message": "pkg/fee: remove unused (#1016)",
          "timestamp": "2025-01-30T06:51:49-06:00",
          "tree_id": "b9b84768660d5a5a63270333e175234c6dd88632",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/64eba2d8808a28003742c6a84d457346528859e5"
        },
        "date": 1738241573272,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 454.9,
            "unit": "ns/op",
            "extra": "2633236 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 519.5,
            "unit": "ns/op",
            "extra": "2357654 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28229,
            "unit": "ns/op",
            "extra": "41907 times\n4 procs"
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
          "id": "6f1f48342e36fa00a3e89d73695922c47aa94987",
          "message": "pkg/sqlutil/sqltest: add CreateOrReplace (#1018)",
          "timestamp": "2025-01-30T14:29:59-06:00",
          "tree_id": "4e0780150c8220ebb75e27d52ab2cc72505cb4ee",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6f1f48342e36fa00a3e89d73695922c47aa94987"
        },
        "date": 1738269052223,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 463.1,
            "unit": "ns/op",
            "extra": "2608483 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 515.1,
            "unit": "ns/op",
            "extra": "2339076 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28234,
            "unit": "ns/op",
            "extra": "42566 times\n4 procs"
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
          "id": "4974de13af77b88d083cff16d88153b129500112",
          "message": "chore: make MaxDecompressedBinarySize inclusive limit (#1019)",
          "timestamp": "2025-01-31T08:22:03-08:00",
          "tree_id": "de88e9b5a31b6e00fafe51553cd16ffb182e1417",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4974de13af77b88d083cff16d88153b129500112"
        },
        "date": 1738340573890,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 457.7,
            "unit": "ns/op",
            "extra": "2630871 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 511.7,
            "unit": "ns/op",
            "extra": "2348233 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28198,
            "unit": "ns/op",
            "extra": "42597 times\n4 procs"
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
          "id": "f22f85eca3d55adcfa914164c1921278a839b9a3",
          "message": "Nested Value Codec Access (#990)\n\n* rename modifier functional for nested fields\n\n* complete path traverse modifier update with backward compatibility\n\n* address comments\n\n---------\n\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2025-02-03T14:31:51+01:00",
          "tree_id": "537b227e232e6062bd63027863d7e167f8e2f960",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f22f85eca3d55adcfa914164c1921278a839b9a3"
        },
        "date": 1738589575409,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.8,
            "unit": "ns/op",
            "extra": "2658900 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 541.7,
            "unit": "ns/op",
            "extra": "2358926 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28201,
            "unit": "ns/op",
            "extra": "42379 times\n4 procs"
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
          "id": "aea9294a7d555844336a92c9ffe41219dfb26c68",
          "message": "Add typecheck to Precodec modifier decodeFieldMapAction (#1020)",
          "timestamp": "2025-02-03T17:29:07Z",
          "tree_id": "68743eb788dcc72dc7272a19c4fb5867deb4c96c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aea9294a7d555844336a92c9ffe41219dfb26c68"
        },
        "date": 1738603811549,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 453.6,
            "unit": "ns/op",
            "extra": "2612893 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 520.9,
            "unit": "ns/op",
            "extra": "2321566 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28248,
            "unit": "ns/op",
            "extra": "42511 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "yikjiun.lee@smartcontract.com",
            "name": "Lee Yik Jiun",
            "username": "leeyikjiun"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8f50d72601bb805041e3082501a9c0bf6138794c",
          "message": "Various updates to Observability Library (#1022)\n\n- Fix alert query type not instant when the query is instant\r\n- Add disable resolve message and uid to contact point\r\n- Add stacking mode to time series panel\r\n- Add multi and include all to custom variable\r\n- Use csv when the variable key and value is the same\r\n- Fix delete notification template status code should be 204\r\n- Add annotations to alerts\r\n- Derive notification templates name from file name",
          "timestamp": "2025-02-05T15:11:37+01:00",
          "tree_id": "5b7cc5644e0ba3af46fb1c7dcc3d44cc87c89f18",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8f50d72601bb805041e3082501a9c0bf6138794c"
        },
        "date": 1738764749110,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 462.5,
            "unit": "ns/op",
            "extra": "2458044 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 512.5,
            "unit": "ns/op",
            "extra": "2337364 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28391,
            "unit": "ns/op",
            "extra": "42544 times\n4 procs"
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
          "id": "760701bde4a15406116beab42844d6367edd6c1c",
          "message": "feat(observability-lib): add interval to panel option (#1024)",
          "timestamp": "2025-02-07T15:36:52+01:00",
          "tree_id": "817067feaed007093b92983970526b553e993590",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/760701bde4a15406116beab42844d6367edd6c1c"
        },
        "date": 1738939143132,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 450.3,
            "unit": "ns/op",
            "extra": "2566411 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 528,
            "unit": "ns/op",
            "extra": "2405812 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28209,
            "unit": "ns/op",
            "extra": "41899 times\n4 procs"
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
          "id": "40efdbab0277d8976cfbd4d5850047ae9675f99f",
          "message": "[CAPPL-308] Add non-data dependency to chain reader (#1025)",
          "timestamp": "2025-02-10T11:13:30Z",
          "tree_id": "83504eea2271720ee0f757ae87eb190795da5d6f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/40efdbab0277d8976cfbd4d5850047ae9675f99f"
        },
        "date": 1739186067430,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 448.3,
            "unit": "ns/op",
            "extra": "2695843 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 502,
            "unit": "ns/op",
            "extra": "2373164 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28295,
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
          "id": "d2aaa393ca5554abd4b5a6a88d81d63552a518ac",
          "message": "Move StepDependency to Inputs (#1026)",
          "timestamp": "2025-02-10T12:23:26Z",
          "tree_id": "dccc79bd687eceaaa1bdc273b6e92a67c2a714d0",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d2aaa393ca5554abd4b5a6a88d81d63552a518ac"
        },
        "date": 1739190263473,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 451.9,
            "unit": "ns/op",
            "extra": "2686108 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 570.1,
            "unit": "ns/op",
            "extra": "2282254 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28223,
            "unit": "ns/op",
            "extra": "42459 times\n4 procs"
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
          "id": "dc2073fe0d21eb352e37792b3ebdeda2616698ac",
          "message": "add bytes to string modifier for solana contracts (#1040)\n\n* add bytes to string modifier for solana contracts\n\n* property extractor path traversal\n\n* update modifiers and value extraction\n\n* fix bug on settable struct\n\n* export value util functions for values at path\n\n* add example to address modifier\n\n* move and rename examples for CI to pass\n\n* move example to test file",
          "timestamp": "2025-02-24T13:05:53-08:00",
          "tree_id": "bc0b172d92bd8df09addcb1e7b99ff1f141ff4b6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/dc2073fe0d21eb352e37792b3ebdeda2616698ac"
        },
        "date": 1740431283832,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.4,
            "unit": "ns/op",
            "extra": "3396902 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.3,
            "unit": "ns/op",
            "extra": "3007951 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28292,
            "unit": "ns/op",
            "extra": "42489 times\n4 procs"
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
          "id": "f0e1dd7b79421faa83946e194507b95eb95df999",
          "message": "fix: syncronize RegisterTrigger client and server using a first ack/err message (#1048)",
          "timestamp": "2025-02-25T11:06:21+01:00",
          "tree_id": "27e2c1671e3a8e781979aba753b9e485123c119a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f0e1dd7b79421faa83946e194507b95eb95df999"
        },
        "date": 1740478050868,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 369.7,
            "unit": "ns/op",
            "extra": "3442149 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 396.5,
            "unit": "ns/op",
            "extra": "3042092 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28224,
            "unit": "ns/op",
            "extra": "42439 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "yikjiun.lee@smartcontract.com",
            "name": "Lee Yik Jiun",
            "username": "leeyikjiun"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "12124a68b50cda36dfc2644be5d9130185854719",
          "message": "feat(obs-lib): add filterable and min width property (#1044)",
          "timestamp": "2025-02-25T21:03:02+08:00",
          "tree_id": "cdd5528bcd67e639a5280ec2f20b0cab3f86d3b7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/12124a68b50cda36dfc2644be5d9130185854719"
        },
        "date": 1740488644099,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 348.7,
            "unit": "ns/op",
            "extra": "3435444 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.5,
            "unit": "ns/op",
            "extra": "2715170 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28217,
            "unit": "ns/op",
            "extra": "42459 times\n4 procs"
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
          "id": "1cc8c4f04c3f696e21fbe90c57fa7340908cfb4c",
          "message": "[CAPPL-595] LLO-compatible trigger event structs (#1051)\n\n* [CAPPL-595] LLO-compatible trigger service (stubs only)\n\nTo onblock LLO-side changes\n\n* Update pkg/capabilities/datastreams/types.go\n\n* Formatting\n\n---------\n\nCo-authored-by: Sam <samsondav@protonmail.com>",
          "timestamp": "2025-02-27T08:50:39-05:00",
          "tree_id": "49ea23c33dc14b3d1c20bf862e79e0761a41e293",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1cc8c4f04c3f696e21fbe90c57fa7340908cfb4c"
        },
        "date": 1740664314788,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 351.7,
            "unit": "ns/op",
            "extra": "3193326 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 397.9,
            "unit": "ns/op",
            "extra": "3024002 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28407,
            "unit": "ns/op",
            "extra": "41950 times\n4 procs"
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
          "id": "2537a8c226bba0b2fe4ce7a0780f3fa6e6967a8e",
          "message": "Add interface support for MedianPluginOption (#1039)\n\n* Add support for MedianPluginOption\n\n* Try loopifying it",
          "timestamp": "2025-02-27T15:30:31-05:00",
          "tree_id": "64407dbb94a55a2da56b370b7664e0c1cb590564",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2537a8c226bba0b2fe4ce7a0780f3fa6e6967a8e"
        },
        "date": 1740688297460,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 407.8,
            "unit": "ns/op",
            "extra": "3287083 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 396.9,
            "unit": "ns/op",
            "extra": "2892010 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28068,
            "unit": "ns/op",
            "extra": "42597 times\n4 procs"
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
          "id": "8777dbcefd5c24350012ac96d5a5ff1b92b4b531",
          "message": "Switch timestamp to uint64 (#1054)",
          "timestamp": "2025-02-28T10:15:13-05:00",
          "tree_id": "d864a7d9080d339c3fe45ad4e8f1614ff7229edd",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8777dbcefd5c24350012ac96d5a5ff1b92b4b531"
        },
        "date": 1740755780001,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 370.6,
            "unit": "ns/op",
            "extra": "3384793 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 395.4,
            "unit": "ns/op",
            "extra": "3007392 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28113,
            "unit": "ns/op",
            "extra": "42639 times\n4 procs"
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
          "id": "00be1cbabe48ef0555d9182b572f96efa32671b5",
          "message": "bumping benchmark action (#1082)",
          "timestamp": "2025-03-24T21:47:59-04:00",
          "tree_id": "67343a0df04e42fd3303475b7c5f939c555db95c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/00be1cbabe48ef0555d9182b572f96efa32671b5"
        },
        "date": 1742867408466,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 385.5,
            "unit": "ns/op",
            "extra": "3244710 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.4,
            "unit": "ns/op",
            "extra": "2770761 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28139,
            "unit": "ns/op",
            "extra": "42884 times\n4 procs"
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
          "id": "c05266d7568233283bbb30fb72ac26a8ec337e38",
          "message": "CRE-293 Add Metering Detail to Capabilities Response (#1080)\n\n* Add Metering Detail to Capabilities Response\n\nEvery node must independently report metering information while the engine expects an aggregated list of all nodes.\nInstead of having two different response types: one for a node and one as an aggregate, a `CapabilityResponse` can\ncontain metadata with an array of values. A node is expected to send a single entity in the array while an aggregated\nresponse would have multiple.\n\nThis change gives the ability to surface metering data from 1 or many nodes without moving to more complex types.\n\n* fix capability loop test\n\n* reduce metadata name\n\n* make another test pass\n\n* fix another test",
          "timestamp": "2025-03-26T07:35:41+01:00",
          "tree_id": "219e93a2829c09732010f77b586863fd589b95b9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c05266d7568233283bbb30fb72ac26a8ec337e38"
        },
        "date": 1742971009291,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.8,
            "unit": "ns/op",
            "extra": "3302904 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.5,
            "unit": "ns/op",
            "extra": "2918049 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28056,
            "unit": "ns/op",
            "extra": "42784 times\n4 procs"
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
          "id": "8485f36ebe7e3e749c46e9e171ca5c55c42bcf08",
          "message": "[chore] Add CreStepTimeout to WriteTarget (#1073)\n\n* [chore] Add CreStepTimeout to WriteTarget\n\n* [chore] Add CreStepTimeout to WriteTarget",
          "timestamp": "2025-03-26T11:01:08Z",
          "tree_id": "945e7767a3c9abcccb2c07180f8e5769b59725e1",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8485f36ebe7e3e749c46e9e171ca5c55c42bcf08"
        },
        "date": 1742986938360,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.6,
            "unit": "ns/op",
            "extra": "3348494 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415,
            "unit": "ns/op",
            "extra": "2848659 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28184,
            "unit": "ns/op",
            "extra": "42784 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "39bc061d09ded8c6b87ff95ffaea53110a742f87",
          "message": "feat:(capabilities) OCR aggregator for LLO-based data feeds (#1079)\n\n* [CAPPL-597] OCR aggregator for LLO-based data feeds\n\n\n* tests working\n\n* fix partial staleness\n\n* cleanup; add example\n\n* linter, fix example, use decimal.Decimal consistently\n\n* typos, docs\n\n* fixing encoded outcome\n\n* cleanup\n\n* json schema validation\n\n* cleanup; rm OCRTriggerEvent to simplify the dependency and layering\n\n* address comments\n\n* fix test\n\n* maintain compatability; depreciate OCREvent\n\n---------\n\nCo-authored-by: Bolek Kulbabinski <1416262+bolekk@users.noreply.github.com>",
          "timestamp": "2025-03-26T12:37:26-06:00",
          "tree_id": "cee820d8e47cd9658c426a16ae596572e25e79b5",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/39bc061d09ded8c6b87ff95ffaea53110a742f87"
        },
        "date": 1743014319380,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 367.6,
            "unit": "ns/op",
            "extra": "3273548 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.1,
            "unit": "ns/op",
            "extra": "2846638 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28861,
            "unit": "ns/op",
            "extra": "42285 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "dylan.tinianov@smartcontract.com",
            "name": "Dylan Tinianov",
            "username": "DylanTinianov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8dad34f94aa5f30528de595b9db71465531398cf",
          "message": "Add Replay to Relayer (#1072)\n\n* Add Replay\n\n* Add ReplayStatus strings\n\n* Update relayer.go\n\n* Add Replay to Relayer\n\n* Update relayer.pb.go\n\n* Add replay to tests\n\n* Update relayer.go\n\n* Use string fromBlock\n\n* trigger CI\n\n* trigger CI\n\n* trigger CI\n\n* Remove ReplayStatus",
          "timestamp": "2025-03-26T20:43:43-04:00",
          "tree_id": "b3f25228e3aeb27dc086dc0862cb731c80c80453",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8dad34f94aa5f30528de595b9db71465531398cf"
        },
        "date": 1743036304160,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 373.9,
            "unit": "ns/op",
            "extra": "3325424 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.3,
            "unit": "ns/op",
            "extra": "2905951 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28108,
            "unit": "ns/op",
            "extra": "42535 times\n4 procs"
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
          "id": "e4827f4b9cb5d2e51ff72c5d593051cdd65aa79c",
          "message": "removing global state for beholder tester (#1085)\n\n* removing global state for beholder tester\n\n* adding doc\n\n---------\n\nCo-authored-by: Vyzaldy Sanchez <vyzaldysanchez@gmail.com>",
          "timestamp": "2025-03-27T09:38:35-04:00",
          "tree_id": "305c77ae0efbff1237655b19d361a8b893022620",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e4827f4b9cb5d2e51ff72c5d593051cdd65aa79c"
        },
        "date": 1743082798387,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 363,
            "unit": "ns/op",
            "extra": "3318004 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 447.2,
            "unit": "ns/op",
            "extra": "2853153 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28020,
            "unit": "ns/op",
            "extra": "42721 times\n4 procs"
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
          "id": "ef0e2432cdca9d1ec630a7ac7a3e3cbba49b014c",
          "message": "tests.Context(t) --> t.Context(t) (#1086)",
          "timestamp": "2025-03-27T10:26:01-04:00",
          "tree_id": "1b175fa4a01d2b4647be2327d299930070ee0cf3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ef0e2432cdca9d1ec630a7ac7a3e3cbba49b014c"
        },
        "date": 1743085634624,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 371.1,
            "unit": "ns/op",
            "extra": "3267722 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 421.7,
            "unit": "ns/op",
            "extra": "2820186 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28033,
            "unit": "ns/op",
            "extra": "42747 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c11ba1cf36826b86449aa1c12514fa91f12c0894",
          "message": "fix llo aggregator timestamp for cache contract encoding (#1089)\n\n* fix llo aggregator timestamp for cache contract encoding\n\n* linter",
          "timestamp": "2025-03-27T11:49:12-06:00",
          "tree_id": "979b6b156d5d5a443a8d00911368ac0f3e3e7cc4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c11ba1cf36826b86449aa1c12514fa91f12c0894"
        },
        "date": 1743097823221,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366,
            "unit": "ns/op",
            "extra": "3270852 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 447.5,
            "unit": "ns/op",
            "extra": "2479548 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28278,
            "unit": "ns/op",
            "extra": "42819 times\n4 procs"
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
          "id": "235758e9e3d8ab49462d3c6f704eb4d9faf6fadf",
          "message": "pkg/types/evm: create package with ContractReaderConfig (#1076)",
          "timestamp": "2025-03-27T15:28:59-05:00",
          "tree_id": "5f542d92afab27c21ba1295deda1ffc3690e6687",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/235758e9e3d8ab49462d3c6f704eb4d9faf6fadf"
        },
        "date": 1743107420627,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 389.3,
            "unit": "ns/op",
            "extra": "3276052 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.2,
            "unit": "ns/op",
            "extra": "2890777 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28066,
            "unit": "ns/op",
            "extra": "42807 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5460e9530e1e20d128df1e2d5390bdeed62c7dcd",
          "message": "chore(ocr-capability): cleanup feed configuration (#1092)\n\n* wip\n\n* clean up the config structs",
          "timestamp": "2025-03-28T18:12:40-06:00",
          "tree_id": "9d7481970e9727301602dd22547de5907fea863b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5460e9530e1e20d128df1e2d5390bdeed62c7dcd"
        },
        "date": 1743207242672,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362,
            "unit": "ns/op",
            "extra": "2877956 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 421,
            "unit": "ns/op",
            "extra": "2892157 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28012,
            "unit": "ns/op",
            "extra": "42807 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cfal@users.noreply.github.com",
            "name": "cfal",
            "username": "cfal"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "84ec641e075870c9eab73b025282d2d29bba96dd",
          "message": "fix chainwriter LOOP plugin service (#1096)\n\n* pkg/loop/internal/relayer/pluginprovider/contractwriter/contract_writer.go: actually register the contract writer\n\n* pkg/loop/internal/relayer/pluginprovider/contractwriter/contract_writer.go: fix SubmitTransaction, decode params",
          "timestamp": "2025-03-29T08:13:13Z",
          "tree_id": "fa3b8fd0e6119ce70d9c566140311e3e34609d09",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/84ec641e075870c9eab73b025282d2d29bba96dd"
        },
        "date": 1743236072073,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.6,
            "unit": "ns/op",
            "extra": "3284790 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 426.7,
            "unit": "ns/op",
            "extra": "2897170 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28457,
            "unit": "ns/op",
            "extra": "42853 times\n4 procs"
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
          "id": "17bfd8db7e32f8932e61456966fea17be79151e5",
          "message": "Fix wrapper mod for primitves wrapping and extractor modifiers for nested slice of slices (#1090)\n\n* Fix nested slice of slice property extractor mod\n\n* Fix wrapper modifier for primitives\n\n* Fix wrapper modifier tests",
          "timestamp": "2025-03-31T16:11:51+02:00",
          "tree_id": "527a230deb341e0f03715dc8bd8502c5fe405282",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/17bfd8db7e32f8932e61456966fea17be79151e5"
        },
        "date": 1743430395097,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.7,
            "unit": "ns/op",
            "extra": "3221959 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 434.6,
            "unit": "ns/op",
            "extra": "2896154 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28061,
            "unit": "ns/op",
            "extra": "42811 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5c320a9353928934c224f1b662b540bb094a3104",
          "message": "feat(ocr3-capability): llo benchmarks (#1095)\n\n* wip\n\n* clean up the config structs\n\n* cleanup and benchmark ocr3 reporting plugin for llo\n\n* benchmark combination of wf and streams\n\n* fix race; transmitter\n\n* add benchmark for observation phase",
          "timestamp": "2025-03-31T09:22:48-06:00",
          "tree_id": "477b3dce91514f6e96309d14329ddc027c861797",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5c320a9353928934c224f1b662b540bb094a3104"
        },
        "date": 1743434646872,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365.6,
            "unit": "ns/op",
            "extra": "3287060 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 442.6,
            "unit": "ns/op",
            "extra": "2865816 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28229,
            "unit": "ns/op",
            "extra": "42751 times\n4 procs"
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
          "id": "00dd1822393d55515a27a38e0febdf1547040acb",
          "message": "pkg/capabilities/consensus/ocr3: use services.Engine (#1009)",
          "timestamp": "2025-03-31T12:15:10-05:00",
          "tree_id": "ebfc1b39bf7d3127196cbf3bf2e3271fd2e7f7c4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/00dd1822393d55515a27a38e0febdf1547040acb"
        },
        "date": 1743441391105,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365,
            "unit": "ns/op",
            "extra": "3218314 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 418.5,
            "unit": "ns/op",
            "extra": "2865387 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28061,
            "unit": "ns/op",
            "extra": "42199 times\n4 procs"
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
          "id": "46f0d8f85c3cf77ca5b4b6141768059ecf6e2ab3",
          "message": "metering for consensus cap (#1099)\n\n* metering for consensus cap\n\n* Adding MeterableCapability iface\n\n* clean up\n\n* removing MeterableCapability iface in favor of MeteringUnit struct\n\n* fixing tests\n\n* values.ByteSizeOfMap --> private utility\n\n* updating access pattern to metering.unit\n\n* metering.XyzMeteringUnit --> metering.XyzUnit",
          "timestamp": "2025-03-31T17:09:04-04:00",
          "tree_id": "4309cfe33643361d33d76af4e51a8b4564f35dd9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/46f0d8f85c3cf77ca5b4b6141768059ecf6e2ab3"
        },
        "date": 1743455416728,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 360.4,
            "unit": "ns/op",
            "extra": "3319963 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.4,
            "unit": "ns/op",
            "extra": "2851215 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28221,
            "unit": "ns/op",
            "extra": "42859 times\n4 procs"
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
          "id": "1b355c7c8c033b55384de86450d41be1af874a08",
          "message": "Move Billing Service Proto to Common (#1088)\n\n* new billing service proto file\n\n* generate go files\n\n* remove unused import\n\n* move client to common\n\n* using test loggers\n\n---------\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2025-03-31T20:57:18-04:00",
          "tree_id": "8fe27baea04a0db18b7f040a72344e463cd9e276",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1b355c7c8c033b55384de86450d41be1af874a08"
        },
        "date": 1743469111985,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.3,
            "unit": "ns/op",
            "extra": "3311097 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.6,
            "unit": "ns/op",
            "extra": "2793524 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28606,
            "unit": "ns/op",
            "extra": "42189 times\n4 procs"
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
          "id": "d0dccede284bef732af6d7435dd3ef040d48c97f",
          "message": "ocr3 should respect context (#1102)",
          "timestamp": "2025-04-01T08:50:45-07:00",
          "tree_id": "1bc6067f29f66786768681b003ecac025ed24886",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d0dccede284bef732af6d7435dd3ef040d48c97f"
        },
        "date": 1743522773534,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.5,
            "unit": "ns/op",
            "extra": "3305546 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.3,
            "unit": "ns/op",
            "extra": "2750580 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28632,
            "unit": "ns/op",
            "extra": "42184 times\n4 procs"
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
          "id": "d68a079c09d16d0b3391b4906d79c0470341e878",
          "message": "Add features to wrapper and property extractor modifiers (#1098)\n\n* Change hardcoder to support TransformToOffChain for primitive variables\n\n* Fix err handling in transformWithMapsHelper\n\n* Improve extractElement to work on uninitialised slices\n\n* Add a test for hardcoder TransformToOffChain for primitives variables\n\n* Handle uninitialised slices in property extractor and add tests\n\n* Improve transformWithMapsHelper err message\n\n* minor improvements\n\n* Improve err handling in initSliceForFieldPath\n\n* Use derefPtr helper",
          "timestamp": "2025-04-01T20:38:09+02:00",
          "tree_id": "88945f555153f33aa3e550ec180af6e19d428619",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d68a079c09d16d0b3391b4906d79c0470341e878"
        },
        "date": 1743532823472,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.2,
            "unit": "ns/op",
            "extra": "3302316 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 418.3,
            "unit": "ns/op",
            "extra": "2884372 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28436,
            "unit": "ns/op",
            "extra": "42234 times\n4 procs"
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
          "id": "11e6c9259f652d7164daaa249f3770c819e8efd9",
          "message": "Change element extractor to not initialise slice elem for nil values (#1109)",
          "timestamp": "2025-04-04T00:13:00+09:00",
          "tree_id": "ec3a0d6e91e11ba8f48a36e8106803ceceb43c13",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/11e6c9259f652d7164daaa249f3770c819e8efd9"
        },
        "date": 1743693312800,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.1,
            "unit": "ns/op",
            "extra": "3300536 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.5,
            "unit": "ns/op",
            "extra": "2888721 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28888,
            "unit": "ns/op",
            "extra": "42164 times\n4 procs"
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
          "id": "dfdf9600557b8d4b4ec047bab6d1cfbb8375e210",
          "message": "[CAPPL-685] Pass through headers correctly (#1110)",
          "timestamp": "2025-04-07T11:00:46+01:00",
          "tree_id": "045206bedf9b073d7037c83ed1669f24208f38f5",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/dfdf9600557b8d4b4ec047bab6d1cfbb8375e210"
        },
        "date": 1744020106709,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.5,
            "unit": "ns/op",
            "extra": "3379287 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.5,
            "unit": "ns/op",
            "extra": "2827130 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28511,
            "unit": "ns/op",
            "extra": "39883 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "42331373+hendoxc@users.noreply.github.com",
            "name": "Hagen H",
            "username": "hendoxc"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "55789e74a0d558e85fea356875ffb973157118f5",
          "message": "INFOPLAT-2071 Beholder emits to chip ingress (#1106)\n\n* INFOPLAT-2071  Adds `chip-ingress` grpc client\n\n* INFOPLAT-2071 Adds chipingress configuration settings\n\nThis needs to be configurable\n\n* INFOPLAT-2071 Adds new emitter implementation using chipingress client\n\n* INFOPLAT-2071 Adds new emitter dual source emitter\n\n* INFOPLAT-2071 Handles creating emitter based on if chip ingress is configured\n\nINFOPLAT-2071 Adjusts code comment\n\nINFOPLAT-2071 Refactor emitters\n\n* INFOPLAT-2071 Adds testing for chip-ingress emitter\n\nINFOPLAT-2071 Adds tests for creating new client\n\nINFOPLAT-2071 Refactors DualSourceEmitter\n\nINFOPLAT-2071 Fixes test\n\n- runs `go fmt`\n\nINFOPLAT-2071 Runs `go fmt`\n\nINFOPLAT-2071 Adds more testing\n\nINFOPLAT-2071 Adds more testing\n\n* INFOPLAT-2071 Makes chip-ingress emitter log error rather than returning it\n\n- minimize disruptions\n- makes it async too\n\n* INFOPLAT-2071 Removes basic auth config from client\n\n* INFOPLAT-2071 Removes panic on nil check\n\n* INFOPLAT-2071 Adds nil checks and more tests\n\n* INFOPLAT-2071 Refactors ChipIngress Client creation\n\n* INFOPLAT-2071 Fixes test\n\n- `go fmt`\n\nINFOPLAT-2071 Small tidyup\n\n* INFOPLAT-2071 Removes race condition test\n\n* INFOPLAT-2071 Use logger from `pkg`",
          "timestamp": "2025-04-07T12:58:44-04:00",
          "tree_id": "222da7795edb1d0b5b45b1bb82b03f2ca6da6795",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/55789e74a0d558e85fea356875ffb973157118f5"
        },
        "date": 1744045261165,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.3,
            "unit": "ns/op",
            "extra": "3264577 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 430.1,
            "unit": "ns/op",
            "extra": "2881746 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28476,
            "unit": "ns/op",
            "extra": "42663 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "168561091+engnke@users.noreply.github.com",
            "name": "engnke",
            "username": "engnke"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d68eef7f6097bf3e622172cee9c5465fa5cf4991",
          "message": "beholder client - custom msg - add support for source/type attribute (#1111)\n\n* beholder client - custom msg - add support for source/type attribute name\n\n* comments\n\n* refactoring extract funtion",
          "timestamp": "2025-04-07T14:43:00-04:00",
          "tree_id": "da3dfb971a26fefdbc7c95ec3a671ea28bebf94e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d68eef7f6097bf3e622172cee9c5465fa5cf4991"
        },
        "date": 1744051450853,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366.7,
            "unit": "ns/op",
            "extra": "3304754 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.4,
            "unit": "ns/op",
            "extra": "2917315 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28171,
            "unit": "ns/op",
            "extra": "42543 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "cawthornegd@gmail.com",
            "name": "cawthorne",
            "username": "cawthorne"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4581dd3fccdcf58344abb87cd6f8b245a767f296",
          "message": "Fix values Wrap(v any) null Interface{} bug (#1104)\n\n* Fix values Wrap(v any) nil ptr bug\n\n* Add tests\n\n* Remove fixture change\n\n* Update comment\n\n---------\n\nCo-authored-by: Cedric <cedric.cordenier@smartcontract.com>",
          "timestamp": "2025-04-08T11:37:19+01:00",
          "tree_id": "89f77bea8ef587e6d89f99749378fff4965c643e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4581dd3fccdcf58344abb87cd6f8b245a767f296"
        },
        "date": 1744108707445,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.6,
            "unit": "ns/op",
            "extra": "3268617 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.1,
            "unit": "ns/op",
            "extra": "2867019 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28138,
            "unit": "ns/op",
            "extra": "42480 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "david@makewhat.is",
            "name": "David Johansen",
            "username": "makewhatis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "361235460f0a51a6cfc3f89c2b4fda931d8bba39",
          "message": "O11Y-1066 - include Type and Id in fields returned from datasource query (#1114)",
          "timestamp": "2025-04-08T16:53:25+02:00",
          "tree_id": "10cdeb4a61e6dc415b5ba982126a81188d8ce24b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/361235460f0a51a6cfc3f89c2b4fda931d8bba39"
        },
        "date": 1744124146915,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 374.3,
            "unit": "ns/op",
            "extra": "3261441 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.2,
            "unit": "ns/op",
            "extra": "2952630 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28186,
            "unit": "ns/op",
            "extra": "42314 times\n4 procs"
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
          "id": "8b1123f4d37664e49bd6a70484d6065d39b11315",
          "message": "refactoring protos (#1113)\n\n* refactoring protos\n\n* fixing make generate",
          "timestamp": "2025-04-09T10:25:20-04:00",
          "tree_id": "caadeeab0ab03fe7e3ddf89c2c45d6791461b786",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8b1123f4d37664e49bd6a70484d6065d39b11315"
        },
        "date": 1744208798453,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.7,
            "unit": "ns/op",
            "extra": "3297996 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.3,
            "unit": "ns/op",
            "extra": "2960071 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29519,
            "unit": "ns/op",
            "extra": "42650 times\n4 procs"
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
          "id": "aef05efee9859f223d28d3482bd5d194914de762",
          "message": "change billing proto package to billing (#1122)\n\n* change billing proto package to billing\n\n* update tool versions",
          "timestamp": "2025-04-10T12:37:46-05:00",
          "tree_id": "05657fab563d82303d76865f4d0f3eaf7516139b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/aef05efee9859f223d28d3482bd5d194914de762"
        },
        "date": 1744306739624,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366.9,
            "unit": "ns/op",
            "extra": "3237975 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.4,
            "unit": "ns/op",
            "extra": "2859433 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28238,
            "unit": "ns/op",
            "extra": "42603 times\n4 procs"
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
          "id": "317a06a50e203f4fff5007105266ba2059d88805",
          "message": "revert 8b1123f4d37664e49bd6a70484d6065d39b11315 (#1126)",
          "timestamp": "2025-04-11T10:22:04-04:00",
          "tree_id": "64cfe265d67c8b696f3c1ed4923c4850b5b57652",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/317a06a50e203f4fff5007105266ba2059d88805"
        },
        "date": 1744381392416,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.2,
            "unit": "ns/op",
            "extra": "2937240 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.4,
            "unit": "ns/op",
            "extra": "2859358 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28159,
            "unit": "ns/op",
            "extra": "42558 times\n4 procs"
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
          "id": "0f78111abeb25a99f7a5220131fb2615f262875c",
          "message": "feat(utils): adds retry utils to common (#1127)\n\n* feat(utils): adds retry utils to common\n\n* refactor: change name and remove unnecessary function\n\n* remove functional options\n\n* move default instance",
          "timestamp": "2025-04-11T15:09:08-04:00",
          "tree_id": "c8b5e0857e58963a6b101a370f7791c5e36f60e2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0f78111abeb25a99f7a5220131fb2615f262875c"
        },
        "date": 1744398616860,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.4,
            "unit": "ns/op",
            "extra": "3118258 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.6,
            "unit": "ns/op",
            "extra": "2938077 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28172,
            "unit": "ns/op",
            "extra": "42632 times\n4 procs"
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
          "id": "b9b85f941ff7ff15a80a96460718979afe2f61ed",
          "message": "feat(wasm): passes MaxRetries from workflow to host (#1128)\n\n* feat(backoff): adds a backoff library to utils\n\n* adds maxTries param to request\n\n* add license and readme\n\n* removes MaxElapsedTime in favor of context\n\n* simplify timing mechanism\n\n* feat(utils): adds retry utils to common\n\n* refactor: change name and remove unnecessary function\n\n* remove functional options\n\n* move default instance\n\n* remove backoff utils\n\n* docstring",
          "timestamp": "2025-04-11T15:36:56-04:00",
          "tree_id": "e24fb097b3ba58136cf05d38064265bd8f8cfe47",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b9b85f941ff7ff15a80a96460718979afe2f61ed"
        },
        "date": 1744400288615,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.8,
            "unit": "ns/op",
            "extra": "3310918 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 432.8,
            "unit": "ns/op",
            "extra": "2911065 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28164,
            "unit": "ns/op",
            "extra": "42466 times\n4 procs"
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
          "id": "a816cede13685cc1be78125a9a84f9380027568d",
          "message": "(feat): Add mode quorum 'all' to reduce aggregator of OCR3 capability (#1129)",
          "timestamp": "2025-04-14T09:06:46-07:00",
          "tree_id": "ef9b2503b58c6482cb5636822d73335e5c1a23a4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a816cede13685cc1be78125a9a84f9380027568d"
        },
        "date": 1744646879440,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 376,
            "unit": "ns/op",
            "extra": "3286149 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.4,
            "unit": "ns/op",
            "extra": "2907574 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28241,
            "unit": "ns/op",
            "extra": "42409 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "lei.shi@smartcontract.com",
            "name": "Lei",
            "username": "shileiwill"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3a2ccad2faaa95c8f44ac6a96ba977817e8e3689",
          "message": "DEVSVCS-1554 add NewEventEmitter to sdk (#1101)\n\n* remove Logger from interface\n\n* remove error from Emit()\n\n---------\n\nCo-authored-by: Justin Kaseman <justinkaseman@live.com>",
          "timestamp": "2025-04-14T09:54:34-07:00",
          "tree_id": "7461cf2e633519202ccec0ec3f7204a8331c8c99",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3a2ccad2faaa95c8f44ac6a96ba977817e8e3689"
        },
        "date": 1744649743198,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.3,
            "unit": "ns/op",
            "extra": "3309363 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 404.9,
            "unit": "ns/op",
            "extra": "2948216 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28155,
            "unit": "ns/op",
            "extra": "42368 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "secure.michele@smartcontract.com",
            "name": "secure-michelemin",
            "username": "secure-michelemin"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "970c51f3a4781e7ed0e792631b116696ca49c9b2",
          "message": "Update CODEOWNERS (#1124)\n\n* Update CODEOWNERS\n\nUpdating CODEOWNERS as described in the Q1 2025 GitHub UAR.\n\n* Update CODEOWNERS\n\nRemove individual users from code owners.\n\n---------\n\nCo-authored-by: Jordan Krage <jmank88@gmail.com>",
          "timestamp": "2025-04-15T14:17:07-04:00",
          "tree_id": "58ce8f77883f3a6b6d4c6bd11ab9fa3f70829509",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/970c51f3a4781e7ed0e792631b116696ca49c9b2"
        },
        "date": 1744741166214,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.8,
            "unit": "ns/op",
            "extra": "3284539 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.6,
            "unit": "ns/op",
            "extra": "2813292 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28493,
            "unit": "ns/op",
            "extra": "42097 times\n4 procs"
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
          "id": "8703639403c7aadd73412cea0c9ab3bffd0a74a4",
          "message": "pkg/monitoring: add go.mod (#1121)\n\n* pkg/monitoring: add go.mod\n\n* pkg/monitoring: rm local replace\n\n---------\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2025-04-15T18:56:44-05:00",
          "tree_id": "0b15094cdee9634f83e6ff85d340568560a052da",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8703639403c7aadd73412cea0c9ab3bffd0a74a4"
        },
        "date": 1744761543563,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 363.6,
            "unit": "ns/op",
            "extra": "3283263 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.9,
            "unit": "ns/op",
            "extra": "2943532 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28488,
            "unit": "ns/op",
            "extra": "42314 times\n4 procs"
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
          "id": "615547d9128099ec4e9f87b4d91db414dba05bd9",
          "message": "Revert \"revert 8b1123f4d37664e49bd6a70484d6065d39b11315 (#1126)\" (#1136)\n\n* Revert \"revert 8b1123f4d37664e49bd6a70484d6065d39b11315 (#1126)\"\n\nThis reverts commit 317a06a50e203f4fff5007105266ba2059d88805.\n\n* simplify error comparison\n\n---------\n\nCo-authored-by: Awbrey Hughlett <awbrey.hughlett@smartcontract.com>",
          "timestamp": "2025-04-16T16:59:44-04:00",
          "tree_id": "07b8ec89a31400c6fb05936a01ca7becfb8be512",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/615547d9128099ec4e9f87b4d91db414dba05bd9"
        },
        "date": 1744837319293,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366.8,
            "unit": "ns/op",
            "extra": "3276382 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.4,
            "unit": "ns/op",
            "extra": "2888659 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28454,
            "unit": "ns/op",
            "extra": "42436 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "96362174+chainchad@users.noreply.github.com",
            "name": "chainchad",
            "username": "chainchad"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "96abde478d080b087758dde98973f4f3928f8f71",
          "message": "Migrate over loopinstall for installing loop plugins (#1125)",
          "timestamp": "2025-04-17T09:46:54-04:00",
          "tree_id": "9b0c1015c8283b07b031f223e8e031fe803e074d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/96abde478d080b087758dde98973f4f3928f8f71"
        },
        "date": 1744897749722,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 371.6,
            "unit": "ns/op",
            "extra": "3305356 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.5,
            "unit": "ns/op",
            "extra": "2933192 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28167,
            "unit": "ns/op",
            "extra": "41802 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "albertpun8@gmail.com",
            "name": "AP12",
            "username": "albert597"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "95ab818ed750a0dfd1206032c891754f127a5290",
          "message": "Rename codeowner team from realtime->data-tooling (#1141)\n\n* change codeowner\n\n* realtime->data-tooling\n\n* fix typo\n\n* in-place",
          "timestamp": "2025-04-17T12:45:51-04:00",
          "tree_id": "a15976c8783a8e326510f8b9642a3f1c67b903a6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/95ab818ed750a0dfd1206032c891754f127a5290"
        },
        "date": 1744908481925,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.5,
            "unit": "ns/op",
            "extra": "3257150 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.3,
            "unit": "ns/op",
            "extra": "2947413 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28193,
            "unit": "ns/op",
            "extra": "42607 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "albertpun8@gmail.com",
            "name": "AP12",
            "username": "albert597"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ed69fd072ac14fa2a56fe10bcac55cf4a3d20c19",
          "message": "added csa auth to chip ingress (#1132)\n\n* added csa auth to chip ingress\n\n* add header provider\n\n* remove cfg.headers and add static auth\n\n* use imported HeaderProvider from chipingress\n\n* add tests for interceptor\n\n* test auth header\n\n* move unary interceptor into own func with tests\n\n---------\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2025-04-17T13:25:57-04:00",
          "tree_id": "efc9c9114d8ec9f88de97aa590ef0fe9d732ca81",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ed69fd072ac14fa2a56fe10bcac55cf4a3d20c19"
        },
        "date": 1744910826490,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.8,
            "unit": "ns/op",
            "extra": "3313747 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.3,
            "unit": "ns/op",
            "extra": "2883499 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28329,
            "unit": "ns/op",
            "extra": "42465 times\n4 procs"
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
          "id": "d4ab451cef3168aab40053168537883c46bcb405",
          "message": "beholder: config options for chip-ingress (#1123)\n\n* beholder: config options for chip-ingress\n\n- add CHIP-Ingess config options\n- make CL_CHIP_INGRESS_ENABLED env var optional\n\n* Remove ChipIngressEnabled config option\n\n---------\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2025-04-17T13:38:23-04:00",
          "tree_id": "ce3b556d28ee11f4c7d9727b636b4a0e2b74b6eb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d4ab451cef3168aab40053168537883c46bcb405"
        },
        "date": 1744911575690,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.9,
            "unit": "ns/op",
            "extra": "2951289 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.3,
            "unit": "ns/op",
            "extra": "2898458 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28219,
            "unit": "ns/op",
            "extra": "42632 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "96362174+chainchad@users.noreply.github.com",
            "name": "chainchad",
            "username": "chainchad"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5ca460a403430c7039d6fe49fe70e71a316b1ef4",
          "message": "Use multi-line string to clean up help/usage for loopinstall (#1138)",
          "timestamp": "2025-04-17T14:40:25-04:00",
          "tree_id": "a025790ace1f1b9171901b4b85607b180a64562f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5ca460a403430c7039d6fe49fe70e71a316b1ef4"
        },
        "date": 1744915298435,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 384.2,
            "unit": "ns/op",
            "extra": "3247485 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.6,
            "unit": "ns/op",
            "extra": "2829307 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28159,
            "unit": "ns/op",
            "extra": "42668 times\n4 procs"
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
          "id": "160c6083ac30480287cd3edfb7c61108a784691c",
          "message": "[CAPPL-732] Detect NoDAG WASM via specific import (#1143)",
          "timestamp": "2025-04-17T11:56:46-07:00",
          "tree_id": "98eba6e50c09eb03c28723cc9927a9e2ed475b33",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/160c6083ac30480287cd3edfb7c61108a784691c"
        },
        "date": 1744916272797,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.2,
            "unit": "ns/op",
            "extra": "3325017 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.9,
            "unit": "ns/op",
            "extra": "2898274 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28135,
            "unit": "ns/op",
            "extra": "42646 times\n4 procs"
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
          "id": "6b24a042d134e83a45dc1b248b1ed4f91297d696",
          "message": "[CAPPL-733] Make WASM Module mockable (#1147)",
          "timestamp": "2025-04-18T10:24:23-07:00",
          "tree_id": "fcf23221c263a5ae0f1b79ab020b584fa6654ec3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6b24a042d134e83a45dc1b248b1ed4f91297d696"
        },
        "date": 1744997141570,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.7,
            "unit": "ns/op",
            "extra": "3345322 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 428.7,
            "unit": "ns/op",
            "extra": "2920250 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28190,
            "unit": "ns/op",
            "extra": "42691 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "888b361327e8e196f5d2bc63164b8eb857178dcc",
          "message": "Add GettransactionFee to EVM chain service (#1142)\n\n* Add EVM chain\n\n* add GetTransactionFee to EVM chain\n\n\n---------\n\nCo-authored-by: ilija <pavlovicilija42@gmail.com>",
          "timestamp": "2025-04-21T16:19:00-04:00",
          "tree_id": "b0ded8a66277b8e744378c8325d62f9b422c7a75",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/888b361327e8e196f5d2bc63164b8eb857178dcc"
        },
        "date": 1745266821953,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 363.1,
            "unit": "ns/op",
            "extra": "3370342 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.6,
            "unit": "ns/op",
            "extra": "2921358 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28120,
            "unit": "ns/op",
            "extra": "42651 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8b59e5dd60e1d782cad97fc804afa6b0f67194dd",
          "message": "add evm relayer service (#1150)\n\n* add evm relayer service\n\n* clenup\n\n* cleanup\n\n* clean NewEVM\n\n* add relayer client support for evm\n\n* add AsEVM client abstraction\n\n* add AsEVMRelayer call to the Relayer\n\n* change AsEVM signature\n\n* fix tests\n\n* make generate",
          "timestamp": "2025-04-23T14:47:41-04:00",
          "tree_id": "1b3f148699c6cf80292863ad1959794de30078b9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8b59e5dd60e1d782cad97fc804afa6b0f67194dd"
        },
        "date": 1745434132576,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.4,
            "unit": "ns/op",
            "extra": "3362138 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 403.4,
            "unit": "ns/op",
            "extra": "2970597 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28095,
            "unit": "ns/op",
            "extra": "42616 times\n4 procs"
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
          "id": "9c46e9e5a7aa07cf3754b06081c21ce3f042187b",
          "message": "Create the interfaces for the CRE v2 (#1153)",
          "timestamp": "2025-04-24T13:10:23-04:00",
          "tree_id": "ba30766512cd7c671157e63aa318e3c218e32871",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9c46e9e5a7aa07cf3754b06081c21ce3f042187b"
        },
        "date": 1745514703231,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 384.1,
            "unit": "ns/op",
            "extra": "3346766 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.7,
            "unit": "ns/op",
            "extra": "2880108 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28189,
            "unit": "ns/op",
            "extra": "42770 times\n4 procs"
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
          "id": "3df386365d0f58168b076dfc1cac1ec17f65f847",
          "message": "extend beholder tester by providing functions for accessing messages (#1151)\n\n* extend beholder tester by providing functions for accessing messages\n\n* Update pkg/utils/tests/beholder.go\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>\n\n---------\n\nCo-authored-by: Patrick <patrick.huie@smartcontract.com>",
          "timestamp": "2025-04-24T13:32:17-04:00",
          "tree_id": "3710aa3e22df39c8089ff46189f9c04c18fb4846",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3df386365d0f58168b076dfc1cac1ec17f65f847"
        },
        "date": 1745516007412,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 474.7,
            "unit": "ns/op",
            "extra": "2155159 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.4,
            "unit": "ns/op",
            "extra": "2762706 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28184,
            "unit": "ns/op",
            "extra": "42645 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "42331373+hendoxc@users.noreply.github.com",
            "name": "Hagen H",
            "username": "hendoxc"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "0aea2032763545bb54c8f42e95702ae5bad9aa76",
          "message": "INFOPLAT-2216 Updates chipingress protos (#1154)",
          "timestamp": "2025-04-24T14:42:37-04:00",
          "tree_id": "0ebdece40ea229313811370a0eca5800239b7f73",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0aea2032763545bb54c8f42e95702ae5bad9aa76"
        },
        "date": 1745520236441,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 360.3,
            "unit": "ns/op",
            "extra": "3351832 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 403.2,
            "unit": "ns/op",
            "extra": "2804316 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28673,
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
          "id": "d9eabb4a45195fd1458459dd0e7a589215d1acc2",
          "message": "Add a client generator for CRE SDK v2. (#1157)",
          "timestamp": "2025-04-25T15:51:05-04:00",
          "tree_id": "7c5e3b78705f2e680b4a5463452cbe1db2f89f3b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d9eabb4a45195fd1458459dd0e7a589215d1acc2"
        },
        "date": 1745610740640,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 367.8,
            "unit": "ns/op",
            "extra": "3326011 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405,
            "unit": "ns/op",
            "extra": "2980394 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28530,
            "unit": "ns/op",
            "extra": "42075 times\n4 procs"
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
          "id": "d1f468b98b68f8541456020011b8b61dd532dd74",
          "message": "Add the ability to call capabilities with an any instead of a values.Value (#1152)",
          "timestamp": "2025-04-28T10:30:40-04:00",
          "tree_id": "6b7bb8693e97afd2fe1fce1cfcf8cf767d231e92",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d1f468b98b68f8541456020011b8b61dd532dd74"
        },
        "date": 1745850727144,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 386.8,
            "unit": "ns/op",
            "extra": "3224956 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 399,
            "unit": "ns/op",
            "extra": "2993388 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28541,
            "unit": "ns/op",
            "extra": "41952 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7ee5f91ed0658677a644cf9f67e3f265d1213dc1",
          "message": "EVMRelayer -> EVMService, remove embedded Relayer from EVMService interface (#1163)",
          "timestamp": "2025-04-29T16:53:37-04:00",
          "tree_id": "974aa22f18a6d428c32c340cfa5abb4a1f806122",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7ee5f91ed0658677a644cf9f67e3f265d1213dc1"
        },
        "date": 1745960093565,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 371.6,
            "unit": "ns/op",
            "extra": "3184106 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.9,
            "unit": "ns/op",
            "extra": "2885822 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28597,
            "unit": "ns/op",
            "extra": "41365 times\n4 procs"
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
          "id": "f50555291980a4700f0c7de9402118ceafb37b98",
          "message": "LICENSE updates (#1166)",
          "timestamp": "2025-04-29T20:56:38-07:00",
          "tree_id": "9b921b0503ae8830c320a3c2417e26496fa9ebc3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f50555291980a4700f0c7de9402118ceafb37b98"
        },
        "date": 1745985478507,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.9,
            "unit": "ns/op",
            "extra": "3132649 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.9,
            "unit": "ns/op",
            "extra": "2896617 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28466,
            "unit": "ns/op",
            "extra": "42057 times\n4 procs"
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
          "id": "d04a3b64e331c0ca65835a6bd1fdfab16df48876",
          "message": "pkg/config/configtest: move package cfgtest from chainlink/v2 (#1103)",
          "timestamp": "2025-04-30T08:33:40-05:00",
          "tree_id": "0ffb7008e469157912e88bff6bf03dd7a4a38bc2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d04a3b64e331c0ca65835a6bd1fdfab16df48876"
        },
        "date": 1746020098454,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368,
            "unit": "ns/op",
            "extra": "3273645 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 439.8,
            "unit": "ns/op",
            "extra": "2883771 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28496,
            "unit": "ns/op",
            "extra": "41964 times\n4 procs"
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
          "id": "6fc48210a06ee8398e7fa019630c89cb7f46fa54",
          "message": "Update Module v2 protos (#1165)",
          "timestamp": "2025-04-30T11:24:57-07:00",
          "tree_id": "70617ca9d7f36388e78c43229998ee17fa956a3c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6fc48210a06ee8398e7fa019630c89cb7f46fa54"
        },
        "date": 1746037574160,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 390.5,
            "unit": "ns/op",
            "extra": "3326745 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.6,
            "unit": "ns/op",
            "extra": "2904219 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28420,
            "unit": "ns/op",
            "extra": "42159 times\n4 procs"
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
          "id": "de4c1fed34c95ed1ad213a02c21acf0d36129485",
          "message": "swap freeport library (#1169)",
          "timestamp": "2025-04-30T20:30:15-06:00",
          "tree_id": "590c47f4b23d143925e2368555511830b0f403ef",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/de4c1fed34c95ed1ad213a02c21acf0d36129485"
        },
        "date": 1746066752562,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.5,
            "unit": "ns/op",
            "extra": "3326248 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.7,
            "unit": "ns/op",
            "extra": "2506116 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28450,
            "unit": "ns/op",
            "extra": "42136 times\n4 procs"
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
          "id": "d46a3f780fb4ac74fce5892d01086ba8c6e554c4",
          "message": "updates required to report meaningful errors from remote capabilities to workflow/beholder (#1133)",
          "timestamp": "2025-05-01T12:54:14+02:00",
          "tree_id": "f8c2059e1fc7f82409a41ab3c76a138e0fa395fa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d46a3f780fb4ac74fce5892d01086ba8c6e554c4"
        },
        "date": 1746096934315,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365.2,
            "unit": "ns/op",
            "extra": "3308302 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.9,
            "unit": "ns/op",
            "extra": "2897642 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28450,
            "unit": "ns/op",
            "extra": "42166 times\n4 procs"
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
          "id": "27c43a698294845e08b3b60d8315256b90834ac0",
          "message": "Remove unused methods from capabilities regitry, split the interface in two. (#1162)",
          "timestamp": "2025-05-01T09:43:04-04:00",
          "tree_id": "831f4604b0fc873cff773a72727764c67db4441d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/27c43a698294845e08b3b60d8315256b90834ac0"
        },
        "date": 1746107061683,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 369.3,
            "unit": "ns/op",
            "extra": "3348828 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.4,
            "unit": "ns/op",
            "extra": "2855426 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28323,
            "unit": "ns/op",
            "extra": "42627 times\n4 procs"
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
          "id": "2e4acd82d330db8e5e5c634231c84adcbd826eeb",
          "message": "Add mock generation to the capability code generator and test runners and runtimes for workflows (#1158)",
          "timestamp": "2025-05-01T14:32:48-04:00",
          "tree_id": "2f73acbcbf82ebc0aaf47dab48f44bd3d500e2f6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2e4acd82d330db8e5e5c634231c84adcbd826eeb"
        },
        "date": 1746124445757,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.8,
            "unit": "ns/op",
            "extra": "3318780 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.6,
            "unit": "ns/op",
            "extra": "2907932 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28594,
            "unit": "ns/op",
            "extra": "42535 times\n4 procs"
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
          "id": "c607f20507eeffeccde82069e8fc6bd796d5d2bc",
          "message": "Use OnceValues in promise (#1159)",
          "timestamp": "2025-05-01T14:43:48-04:00",
          "tree_id": "6dca03a688d09eb9a008be1fb59c4ff8f4c0b0e8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c607f20507eeffeccde82069e8fc6bd796d5d2bc"
        },
        "date": 1746125100966,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.3,
            "unit": "ns/op",
            "extra": "3309991 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.5,
            "unit": "ns/op",
            "extra": "2929831 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28248,
            "unit": "ns/op",
            "extra": "42714 times\n4 procs"
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
          "id": "6ca21edb8fb498704d08c2df06ab7fdff503c2cf",
          "message": "Add capability server helper generation (#1164)",
          "timestamp": "2025-05-02T09:28:44-04:00",
          "tree_id": "6023e34089ac7bea872c0db42180bdbeb8ff59d2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6ca21edb8fb498704d08c2df06ab7fdff503c2cf"
        },
        "date": 1746192598882,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 367.1,
            "unit": "ns/op",
            "extra": "3336601 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.8,
            "unit": "ns/op",
            "extra": "2923539 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28143,
            "unit": "ns/op",
            "extra": "42603 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "77c113986529188653200de04f4abfeaf05e1acf",
          "message": "make db cleanup best effort (#1171)",
          "timestamp": "2025-05-02T08:45:00-06:00",
          "tree_id": "cc0707f20edcd6fcd5b112ef6487d404c601f901",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/77c113986529188653200de04f4abfeaf05e1acf"
        },
        "date": 1746197171607,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.5,
            "unit": "ns/op",
            "extra": "3342297 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.2,
            "unit": "ns/op",
            "extra": "2797885 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28331,
            "unit": "ns/op",
            "extra": "42639 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "cb65e4bb4e3612b83e8442135225f2d6a2cad337",
          "message": "migrate port allocator (#1176)",
          "timestamp": "2025-05-05T09:55:57-06:00",
          "tree_id": "e5f372bbba00d2cb6fb139a92ac40069d69c1f2c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cb65e4bb4e3612b83e8442135225f2d6a2cad337"
        },
        "date": 1746460644057,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.3,
            "unit": "ns/op",
            "extra": "3335547 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.4,
            "unit": "ns/op",
            "extra": "2671864 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28150,
            "unit": "ns/op",
            "extra": "42614 times\n4 procs"
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
          "id": "995623c69ddc8825e6c55c29c3ae5f0c8053f98d",
          "message": "Add WASI runner and runtime for the CRE v2 SDK (#1175)",
          "timestamp": "2025-05-06T14:03:33-04:00",
          "tree_id": "2d75b8e38130eb742908e5d41d312c0605f9fd6e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/995623c69ddc8825e6c55c29c3ae5f0c8053f98d"
        },
        "date": 1746554688754,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.3,
            "unit": "ns/op",
            "extra": "3362811 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.1,
            "unit": "ns/op",
            "extra": "2861689 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28153,
            "unit": "ns/op",
            "extra": "42628 times\n4 procs"
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
          "id": "ea88ef40551195df6a51def2690dfefd26fb58b3",
          "message": "CapabilityExecutor interface for Module/Host (#1177)\n\n* CapabilityExecutor interface for Module/Host\n\n* Include Capability ID field in CapabilityRequest",
          "timestamp": "2025-05-06T11:50:33-07:00",
          "tree_id": "8fe36b74c22814d70863d5cc605dfdb2aba93530",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ea88ef40551195df6a51def2690dfefd26fb58b3"
        },
        "date": 1746557516269,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.5,
            "unit": "ns/op",
            "extra": "2988212 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415,
            "unit": "ns/op",
            "extra": "2893310 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29408,
            "unit": "ns/op",
            "extra": "41990 times\n4 procs"
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
          "id": "817c93bdec09b1171f8feba834f1e785f0ba9759",
          "message": "Add mismatched package naming handling to protoc template generator (#1173)\n\n* Add mismatched package naming handling to protoc template generator\n\n* Add a test capability that tests sdk gen when dir and package names are mismatched\n\n* run generate",
          "timestamp": "2025-05-06T21:31:11+02:00",
          "tree_id": "932b70d9bd2a8d6d4db092f5d4cf070a39afd6ad",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/817c93bdec09b1171f8feba834f1e785f0ba9759"
        },
        "date": 1746559951813,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 363.5,
            "unit": "ns/op",
            "extra": "3337258 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 421.1,
            "unit": "ns/op",
            "extra": "2908875 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28432,
            "unit": "ns/op",
            "extra": "42144 times\n4 procs"
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
          "id": "d41b1d1357132f850e249e900ba4fb34c0c8786d",
          "message": "Register multiple triggers (#1180)\n\n* chore: adds failing test\n\n* refactor: adds a multi trigger workflow test fix\n\n* test clean up\n\n* refactor: pass an ID to generated triggers\n\n* use a handler name vs Id\n\n* chore: test nodag with mock host\n\n* refactor: removes ID and renames Name\n\n* Update pkg/workflows/wasm/host/wasm_nodag_test.go\n\n* chore: clean up nits",
          "timestamp": "2025-05-07T12:32:07-04:00",
          "tree_id": "8123513828756d8e3bc8c9ca1f6b263fe56e101b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d41b1d1357132f850e249e900ba4fb34c0c8786d"
        },
        "date": 1746635658147,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.3,
            "unit": "ns/op",
            "extra": "3351610 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 402,
            "unit": "ns/op",
            "extra": "2984715 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28135,
            "unit": "ns/op",
            "extra": "42648 times\n4 procs"
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
          "id": "976f49f01c90e6bb6644c206fd7614291dff0037",
          "message": "Alias simple consensus type from pb package in the SDK. This allows workflow authors to not need to import the pb package directly (#1182)\n\nCo-authored-by: Street <5597260+MStreet3@users.noreply.github.com>",
          "timestamp": "2025-05-07T13:21:57-04:00",
          "tree_id": "2ad99f32c3b93295d491e95dac33787b4b5c0649",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/976f49f01c90e6bb6644c206fd7614291dff0037"
        },
        "date": 1746638579745,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.4,
            "unit": "ns/op",
            "extra": "3271292 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 421.6,
            "unit": "ns/op",
            "extra": "2622261 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28192,
            "unit": "ns/op",
            "extra": "42619 times\n4 procs"
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
          "id": "db395570d649b1c19bc63153fb83524fa8afe3f7",
          "message": "refactor: exposes CapabilityWrapper (#1183)",
          "timestamp": "2025-05-07T15:06:01-04:00",
          "tree_id": "8cef86622881170d5776a0025995f57b34a3ad30",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/db395570d649b1c19bc63153fb83524fa8afe3f7"
        },
        "date": 1746644822073,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.3,
            "unit": "ns/op",
            "extra": "3354832 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 403.9,
            "unit": "ns/op",
            "extra": "2874721 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28125,
            "unit": "ns/op",
            "extra": "42460 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2b5a5170a351e63d2b6aeea39e912ffc19f60c50",
          "message": "BCFR-1330 (#1161)\n\nExtend EVMService with evm functionalities\n\n---------\n\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2025-05-09T11:53:41-04:00",
          "tree_id": "44eaae9aa79014e9501bf93dfe6e39dfeda2e4ba",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2b5a5170a351e63d2b6aeea39e912ffc19f60c50"
        },
        "date": 1746806093423,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.7,
            "unit": "ns/op",
            "extra": "3360282 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.3,
            "unit": "ns/op",
            "extra": "2929070 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28123,
            "unit": "ns/op",
            "extra": "42686 times\n4 procs"
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
          "id": "47a2f78860af1240e0148f867ad71fbb64f03255",
          "message": "Allow unstubed methods for non-strict triggers (#1178)",
          "timestamp": "2025-05-09T15:11:54-04:00",
          "tree_id": "bdb8f22c4e6b445722f4bc63249ef9b8375d2c11",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/47a2f78860af1240e0148f867ad71fbb64f03255"
        },
        "date": 1746817989212,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.8,
            "unit": "ns/op",
            "extra": "3346030 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.7,
            "unit": "ns/op",
            "extra": "2721775 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28134,
            "unit": "ns/op",
            "extra": "42572 times\n4 procs"
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
          "id": "338652aa53ff6805a321469d29f80e92633b2448",
          "message": "feat: module executes CallCapability provided by workflow engine (#1184)\n\n* wip set capability executor\n\n* feat: implements SetCapabilityExecutor\n\n* feat: tests SetCapabilityExecutor\n\n* feat: set callcapability handler in module\n\n* fix: handling of callId",
          "timestamp": "2025-05-12T10:56:01-04:00",
          "tree_id": "d7a533e193007fcc2063d17ce7fc745af1b14fc3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/338652aa53ff6805a321469d29f80e92633b2448"
        },
        "date": 1747061835213,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 399.4,
            "unit": "ns/op",
            "extra": "3348652 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.1,
            "unit": "ns/op",
            "extra": "2930205 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28142,
            "unit": "ns/op",
            "extra": "42666 times\n4 procs"
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
          "id": "72874b2434b345e776d32ebe651d222974102261",
          "message": "Capability Interface Test Framework (#1083)\n\nThis commit introduces a testing framework for asserting essential tests that all capabilities should conform to. As a\nstarting point, both trigger and executable capabilities must conform to the `BaseCapability` interface and produce\nproperly structured information. Additionally, executable capabilities should return metering information in\na `CapabilityResponse`.\n\nTests for the `BaseCapability` cannot be disabled but those for `ExecutableCapability` can be per capability\nimplementation as each implemtation will be different.",
          "timestamp": "2025-05-13T12:22:55-05:00",
          "tree_id": "6b7175fae8c077fcf3d74c4185ebc91cb0eaefbd",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/72874b2434b345e776d32ebe651d222974102261"
        },
        "date": 1747157050806,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364,
            "unit": "ns/op",
            "extra": "3389190 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.5,
            "unit": "ns/op",
            "extra": "2952498 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28321,
            "unit": "ns/op",
            "extra": "42644 times\n4 procs"
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
          "id": "cb5474e426a230ddd9b7b8a4dc5225b03dd38453",
          "message": "Consensus capability proto (#1186)",
          "timestamp": "2025-05-13T13:55:34-04:00",
          "tree_id": "1c675a7248a8f4a4a12697e174606a74106c15a5",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cb5474e426a230ddd9b7b8a4dc5225b03dd38453"
        },
        "date": 1747159005879,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.1,
            "unit": "ns/op",
            "extra": "2980464 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.9,
            "unit": "ns/op",
            "extra": "2924230 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28166,
            "unit": "ns/op",
            "extra": "42314 times\n4 procs"
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
          "id": "113c305fde94c94d3881b058123f2ce62eb8fcd4",
          "message": "chore: expose test workflow (#1187)\n\n* chore: exposes v2 test workflow\n\n* fix: workflow imports",
          "timestamp": "2025-05-13T14:07:46-04:00",
          "tree_id": "75a53562cc4971cdfd23fff18b117e7f5ccb7e6e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/113c305fde94c94d3881b058123f2ce62eb8fcd4"
        },
        "date": 1747159743043,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.3,
            "unit": "ns/op",
            "extra": "3364388 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.5,
            "unit": "ns/op",
            "extra": "2750623 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28098,
            "unit": "ns/op",
            "extra": "42648 times\n4 procs"
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
          "id": "07ee1579d0e9ffdc5133d92149f4304bf70b47ac",
          "message": "migrate cron capability to v2 api (#1181)\n\n* migrate cron capability to v2 api\n\n* regenerate",
          "timestamp": "2025-05-13T19:53:18+01:00",
          "tree_id": "cd9ec43a6a574ac9b48d5b1409d75906fbe18557",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/07ee1579d0e9ffdc5133d92149f4304bf70b47ac"
        },
        "date": 1747162488248,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.8,
            "unit": "ns/op",
            "extra": "3339858 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 421.9,
            "unit": "ns/op",
            "extra": "2935840 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28130,
            "unit": "ns/op",
            "extra": "42676 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "kiryll.kuzniecow@gmail.com",
            "name": "Kiryll Kuzniecow",
            "username": "kirqz23"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "be3024969afc90aec185e909ecfe0df42bba8fa4",
          "message": "Use beholder newAttributes function to process attributes provided to Emit function (#1188)",
          "timestamp": "2025-05-14T11:50:49-05:00",
          "tree_id": "b758e49fa8a529d61116b81b9e2ad3bd50fb2757",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/be3024969afc90aec185e909ecfe0df42bba8fa4"
        },
        "date": 1747241525295,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.8,
            "unit": "ns/op",
            "extra": "3381271 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.1,
            "unit": "ns/op",
            "extra": "2978599 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28285,
            "unit": "ns/op",
            "extra": "42637 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "90b1d1b66ce4f62f585610f4b9b71b1713629f5e",
          "message": "PLEX-500 (#1179)\n\n* wire QueryTrackedLogs\n\n* evm.EVMVisitor -> evm.Visitor\n\n\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2025-05-15T06:10:02-04:00",
          "tree_id": "52999bf7b03ee37a75fe188046744c0e22f4563d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/90b1d1b66ce4f62f585610f4b9b71b1713629f5e"
        },
        "date": 1747303888842,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 377.3,
            "unit": "ns/op",
            "extra": "3313524 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.3,
            "unit": "ns/op",
            "extra": "2920978 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28911,
            "unit": "ns/op",
            "extra": "41887 times\n4 procs"
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
          "id": "44dbf2720df72e12890d877506dac8c2fd140cd8",
          "message": "refactor: export concrete server type on generate (#1191)\n\n* refactor: export concrete server type on generate\n\n* refactor: make server public",
          "timestamp": "2025-05-15T12:09:21-04:00",
          "tree_id": "6304852a6ecfa6087ef9f2ccde2452261eb357de",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/44dbf2720df72e12890d877506dac8c2fd140cd8"
        },
        "date": 1747325442783,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.1,
            "unit": "ns/op",
            "extra": "3350358 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.2,
            "unit": "ns/op",
            "extra": "2937096 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28471,
            "unit": "ns/op",
            "extra": "38868 times\n4 procs"
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
          "id": "966be0b926dbe7cf475d2d00db724f49112e341a",
          "message": "fix: add nil check for server that is not initialized (#1194)",
          "timestamp": "2025-05-15T14:04:21-04:00",
          "tree_id": "e5867b4dea36acfb2ea6c098e38c4febe434fdef",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/966be0b926dbe7cf475d2d00db724f49112e341a"
        },
        "date": 1747332343623,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.9,
            "unit": "ns/op",
            "extra": "3280738 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 424.9,
            "unit": "ns/op",
            "extra": "2899665 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28543,
            "unit": "ns/op",
            "extra": "42244 times\n4 procs"
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
          "id": "0cdd13a7cb01bf7885e06e8460462954b42be884",
          "message": "chore: adds unit tests and clearer error message on failure (#1193)\n\n* chore: adds unit tests and clearer error message on failure\n\n* fix: properly use errors",
          "timestamp": "2025-05-15T15:51:51-04:00",
          "tree_id": "7c3b32260cc5269f1b7ebb162fbe78684de20abb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0cdd13a7cb01bf7885e06e8460462954b42be884"
        },
        "date": 1747338786021,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.6,
            "unit": "ns/op",
            "extra": "3340802 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.2,
            "unit": "ns/op",
            "extra": "2915422 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28785,
            "unit": "ns/op",
            "extra": "41907 times\n4 procs"
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
          "id": "8262bc72993a9f0951a552836d2ddc0e63bff60b",
          "message": "fix: ensures every call can be found by an await (#1197)",
          "timestamp": "2025-05-15T19:40:41-04:00",
          "tree_id": "32ceeea9e4f4d04cabbb4551544ed20e56f15122",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8262bc72993a9f0951a552836d2ddc0e63bff60b"
        },
        "date": 1747352516477,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.6,
            "unit": "ns/op",
            "extra": "3388064 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.8,
            "unit": "ns/op",
            "extra": "2952675 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28564,
            "unit": "ns/op",
            "extra": "42166 times\n4 procs"
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
          "id": "5d5e3379929b90e15cf0c242a27d5068b008c5d6",
          "message": "fix(capabilities): enable long running trigger response transform (#1198)\n\n* fix(capabilities): separate transform from register request\n\n* fix(capabilities): pass a stop channel to RegisterTrigger in servers",
          "timestamp": "2025-05-16T12:35:04-04:00",
          "tree_id": "d11f0c48bd4cc2b62e2a3d0a67ace27c73b3e751",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5d5e3379929b90e15cf0c242a27d5068b008c5d6"
        },
        "date": 1747413383472,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.3,
            "unit": "ns/op",
            "extra": "3355836 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.4,
            "unit": "ns/op",
            "extra": "2945244 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28635,
            "unit": "ns/op",
            "extra": "41726 times\n4 procs"
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
          "id": "7b64923532730b8227aca43a632fde2ce05fcf87",
          "message": "Fix a bug where generated code cannot import the same package name twice (#1195)",
          "timestamp": "2025-05-16T14:43:34-04:00",
          "tree_id": "7f93a77f416e015bfc5cafda395f5d3fcd36c722",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7b64923532730b8227aca43a632fde2ce05fcf87"
        },
        "date": 1747421092653,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.2,
            "unit": "ns/op",
            "extra": "3336435 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.1,
            "unit": "ns/op",
            "extra": "2913996 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28719,
            "unit": "ns/op",
            "extra": "42133 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "80bc8b13c0e7839849182ef9e55b8ac0063145c9",
          "message": "add GetEstimateFee (#1196)\n\n* add GetEstimateFee\n\n\n\n---------\n\nCo-authored-by: ilija42 <57732589+ilija42@users.noreply.github.com>",
          "timestamp": "2025-05-19T12:12:08-04:00",
          "tree_id": "5bfb469f3ba3e47f221aafe71507efa2cdd88725",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/80bc8b13c0e7839849182ef9e55b8ac0063145c9"
        },
        "date": 1747671204397,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 364.2,
            "unit": "ns/op",
            "extra": "3297458 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 428.5,
            "unit": "ns/op",
            "extra": "2767701 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28164,
            "unit": "ns/op",
            "extra": "42610 times\n4 procs"
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
          "id": "1ae98580fe03ade3297cdc94fe859ab5f4889eec",
          "message": "Part of CAPPL-816: Use the consensus capability SDK in the runtime. A fake still needs to be used. (#1192)",
          "timestamp": "2025-05-21T10:28:53-04:00",
          "tree_id": "1df642979f3466aa3af1da6341829d80e2e77910",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1ae98580fe03ade3297cdc94fe859ab5f4889eec"
        },
        "date": 1747837811130,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366.3,
            "unit": "ns/op",
            "extra": "3280584 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.8,
            "unit": "ns/op",
            "extra": "2863972 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28118,
            "unit": "ns/op",
            "extra": "42604 times\n4 procs"
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
          "id": "0d4dd3e1d000ee5050e20c33d1a88cada73fadf1",
          "message": "billing protos references protos repo (#1199)",
          "timestamp": "2025-05-21T10:01:42-05:00",
          "tree_id": "d10c765452ea5c5afa87b7eda9a8161c5940e1d2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0d4dd3e1d000ee5050e20c33d1a88cada73fadf1"
        },
        "date": 1747839841437,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 371.7,
            "unit": "ns/op",
            "extra": "3300871 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 426.2,
            "unit": "ns/op",
            "extra": "2797413 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28965,
            "unit": "ns/op",
            "extra": "42021 times\n4 procs"
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
          "id": "eba13189be0f015ff10fcb6abca04bc91f72b98a",
          "message": "Plex 131 evm capability Part 1 (#1174)\n\n* Change proto import paths where needed to be relative to root pkg dir\n\n* Add evm chain capability\n\n* Fix evm chain cap protos\n\n* Add evm chain cap sdk grpc gen\n\n* Fix evm cap protos\n\n* Add EVM() to Relayer in relayer set\n\n* Fix evm Relayer grpc bigint conversions\n\n* Fix proto gen\n\n* rm unused constants\n\n* fix errors\n\n* fixing imports and code-gen\n\n* Add query tracked logs grpc converter\n\n* Update proto convertors\n\n* Fix query.proto gen\n\n* Fix proto conversions in evm.go\n\n* package ref fix\n\n* Implement EVM service on the relayer set\n\n* Improve evm relayerset:\n\n- Detach EVM grpc service from RelayerSet proto\n- Refactor code and files to be more intuitive\n\n* Cleanup evm service proto conversion helpers\n\n* Resolve merge pb gen conflicts\n\n* minor fix\n\n* Restructure evm chain cap and service file structure and minor improvements\n\n* Consolidate GetTransactionByHash and GetTransactionReceipt naming\n\n* cleanup unused code\n\n* Expand and cleanup relayerset test\n\n* Move evm service to chain-cap folder so that the cap impl. can import it\n\n* Move proto helpers to chain-capabilities/evm/chain-service so that the cap impl. can import them\n\n* Reorganise query and codec proto to fix cyclical and internal package imports from the evm chain cap\n\n* Fix expressions proto convertor\n\n* Run make generate\n\n* Change evm client and evm client relayerset naming capitalisation and rm leftover print\n\n* Make linter happy\n\n* rerun CI\n\n* Move codec to internal/codec from loop/chain-common\n\n* run generate\n\n* extract changes to split PR up\n\n---------\n\nCo-authored-by: Lautaro Fernandez <juan.lautarofernandez@smartcontract.com>",
          "timestamp": "2025-05-21T17:57:59+02:00",
          "tree_id": "c7e6e25085a97bdfe0c4ca86761fc43df60b3117",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eba13189be0f015ff10fcb6abca04bc91f72b98a"
        },
        "date": 1747843159437,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.8,
            "unit": "ns/op",
            "extra": "3364533 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 419.7,
            "unit": "ns/op",
            "extra": "2859006 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28508,
            "unit": "ns/op",
            "extra": "42028 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vladimiramnell@gmail.com",
            "name": "Vladimir",
            "username": "Unheilbar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "723cad356d858717ad07962a28c2c2561a45fd4e",
          "message": "update description for GetTransactionFee (#1201)",
          "timestamp": "2025-05-21T12:37:34-04:00",
          "tree_id": "a41ca28ae56dd6b922c6b8a96e52a3bb85d60461",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/723cad356d858717ad07962a28c2c2561a45fd4e"
        },
        "date": 1747845524570,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.5,
            "unit": "ns/op",
            "extra": "3353262 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.7,
            "unit": "ns/op",
            "extra": "2568116 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28562,
            "unit": "ns/op",
            "extra": "42156 times\n4 procs"
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
          "id": "65a9b738252b2d2d15f0c1c83bd878124c1ec74b",
          "message": "Plex 131 evm capability part 2 (#1202)\n\n* Change proto import paths where needed to be relative to root pkg dir\n\n* Add evm chain capability\n\n* Fix evm chain cap protos\n\n* Add evm chain cap sdk grpc gen\n\n* Fix evm cap protos\n\n* Add EVM() to Relayer in relayer set\n\n* Fix evm Relayer grpc bigint conversions\n\n* Fix proto gen\n\n* rm unused constants\n\n* fix errors\n\n* fixing imports and code-gen\n\n* Add query tracked logs grpc converter\n\n* Update proto convertors\n\n* Fix query.proto gen\n\n* Fix proto conversions in evm.go\n\n* package ref fix\n\n* Implement EVM service on the relayer set\n\n* Improve evm relayerset:\n\n- Detach EVM grpc service from RelayerSet proto\n- Refactor code and files to be more intuitive\n\n* Cleanup evm service proto conversion helpers\n\n* Resolve merge pb gen conflicts\n\n* minor fix\n\n* Restructure evm chain cap and service file structure and minor improvements\n\n* Consolidate GetTransactionByHash and GetTransactionReceipt naming\n\n* cleanup unused code\n\n* Expand and cleanup relayerset test\n\n* Move evm service to chain-cap folder so that the cap impl. can import it\n\n* Move proto helpers to chain-capabilities/evm/chain-service so that the cap impl. can import them\n\n* Reorganise query and codec proto to fix cyclical and internal package imports from the evm chain cap\n\n* Fix expressions proto convertor\n\n* Run make generate\n\n* Change evm client and evm client relayerset naming capitalisation and rm leftover print\n\n* Make linter happy\n\n* rerun CI\n\n* Move codec to internal/codec from loop/chain-common\n\n* run generate\n\n* extract changes to split PR up\n\n* Revert \"extract changes to split PR up\"\n\nThis reverts commit 94834cdd0cb5e4833c40bae7dfbd59b18b37b90d.\n\n---------\n\nCo-authored-by: Lautaro Fernandez <juan.lautarofernandez@smartcontract.com>",
          "timestamp": "2025-05-21T21:02:41+02:00",
          "tree_id": "cb673f1406c016d5d9e3578a5a40a5b04cc97d1d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/65a9b738252b2d2d15f0c1c83bd878124c1ec74b"
        },
        "date": 1747854249016,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.7,
            "unit": "ns/op",
            "extra": "3265058 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 418.2,
            "unit": "ns/op",
            "extra": "2931909 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28253,
            "unit": "ns/op",
            "extra": "42661 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "kiryll.kuzniecow@gmail.com",
            "name": "Kiryll Kuzniecow",
            "username": "kirqz23"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6309c8950e05a023899a39bdd3381f1cab089286",
          "message": "Add ShowAllValues option to Gauge Panels (#1206)",
          "timestamp": "2025-05-22T07:24:33-05:00",
          "tree_id": "f65d904355d8726631086d34fb48238d832b75f7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6309c8950e05a023899a39bdd3381f1cab089286"
        },
        "date": 1747916755219,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 351.4,
            "unit": "ns/op",
            "extra": "3401296 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.3,
            "unit": "ns/op",
            "extra": "2943999 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28149,
            "unit": "ns/op",
            "extra": "42670 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "108959691+amit-momin@users.noreply.github.com",
            "name": "amit-momin",
            "username": "amit-momin"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "da3dec84d2cd2cafd36f9253a43f702f37efdf27",
          "message": "Fixed typo (#1203)",
          "timestamp": "2025-05-22T10:41:27-05:00",
          "tree_id": "7aa1b256c7cbb584b69f408d4a322c6ba522bbac",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/da3dec84d2cd2cafd36f9253a43f702f37efdf27"
        },
        "date": 1747928571685,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 352.5,
            "unit": "ns/op",
            "extra": "3408226 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 406.3,
            "unit": "ns/op",
            "extra": "2941088 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28255,
            "unit": "ns/op",
            "extra": "42450 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "kiryll.kuzniecow@gmail.com",
            "name": "Kiryll Kuzniecow",
            "username": "kirqz23"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "bb4679b11a8008b137753cc03a7fbacde624792b",
          "message": "Add NoValue option to Gauge panel builders (#1207)",
          "timestamp": "2025-05-22T18:31:01+02:00",
          "tree_id": "639c7ab03e8435880ac9ff34930f4aeb51b0febb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bb4679b11a8008b137753cc03a7fbacde624792b"
        },
        "date": 1747931546003,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 352.4,
            "unit": "ns/op",
            "extra": "3403509 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.4,
            "unit": "ns/op",
            "extra": "2930274 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28534,
            "unit": "ns/op",
            "extra": "42712 times\n4 procs"
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
          "id": "9ac137199b9949f60c9707fab9fc6c1e01423d27",
          "message": "Billing: Use GRPC NewClient (#1204)\n\n* Billing: Use GRPC NewClient\n\nTo avoid a blocking GRPC billing client constructor, avoid using the deprecated `Dial` function\nin favor of the `NewClient` constructor. This bakes-in reconnects.\n\n* fix tests",
          "timestamp": "2025-05-22T12:28:03-05:00",
          "tree_id": "5116c54ccb7b27b17b44b4f76db74513962b31a3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9ac137199b9949f60c9707fab9fc6c1e01423d27"
        },
        "date": 1747934962452,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355,
            "unit": "ns/op",
            "extra": "3385381 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.6,
            "unit": "ns/op",
            "extra": "2844160 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28187,
            "unit": "ns/op",
            "extra": "42632 times\n4 procs"
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
          "id": "44d96a7ad0e552fa85bf28dfc4c49587cbcb672b",
          "message": "CAPPL-881: Make capability call IDs deterministic (#1200)",
          "timestamp": "2025-05-22T14:39:27-04:00",
          "tree_id": "38c07850b136a47541d28414c0026fbb23e0fe6b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/44d96a7ad0e552fa85bf28dfc4c49587cbcb672b"
        },
        "date": 1747939253498,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.9,
            "unit": "ns/op",
            "extra": "3367711 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.2,
            "unit": "ns/op",
            "extra": "2934360 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28205,
            "unit": "ns/op",
            "extra": "42492 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "96362174+chainchad@users.noreply.github.com",
            "name": "chainchad",
            "username": "chainchad"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "191131ef4d110a96a01e07d8cadb63fa1a4cd8f5",
          "message": "Fix loopinstall to install via local relative path (#1210)\n\n* Fix loopinstall to install via local relative path\n\n* Avoid filepath.Clean() and remove env var expansion in yaml inputs to simplify\n\n* Remove outdated docs",
          "timestamp": "2025-05-27T10:01:10-04:00",
          "tree_id": "72477b60320b0a4e963f32044b3ce1e3360c6541",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/191131ef4d110a96a01e07d8cadb63fa1a4cd8f5"
        },
        "date": 1748354551168,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.7,
            "unit": "ns/op",
            "extra": "3416635 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 424.9,
            "unit": "ns/op",
            "extra": "2932110 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28205,
            "unit": "ns/op",
            "extra": "42619 times\n4 procs"
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
          "id": "5a4fb8e255ffd85c9d1da659aa0b0cc453862424",
          "message": "use insecure credentials as the default fallback for local testing (#1212)",
          "timestamp": "2025-05-27T11:40:52-05:00",
          "tree_id": "bbc0c2ecf8805e8132e5ce104f31dfea44a0e6b1",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5a4fb8e255ffd85c9d1da659aa0b0cc453862424"
        },
        "date": 1748364134762,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 351.7,
            "unit": "ns/op",
            "extra": "3166165 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.8,
            "unit": "ns/op",
            "extra": "2860609 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28115,
            "unit": "ns/op",
            "extra": "42651 times\n4 procs"
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
          "id": "89eb1ce3d76f150c9d7f7cf80d8e3dd63ac5f90d",
          "message": "pkg/beholder: move loop.OtelAttributes here (#1208)",
          "timestamp": "2025-05-28T08:36:21-05:00",
          "tree_id": "19410d114318a221f6513459902dc4dfb98124b2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/89eb1ce3d76f150c9d7f7cf80d8e3dd63ac5f90d"
        },
        "date": 1748439460982,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.5,
            "unit": "ns/op",
            "extra": "3321808 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 404.8,
            "unit": "ns/op",
            "extra": "2952025 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28149,
            "unit": "ns/op",
            "extra": "42624 times\n4 procs"
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
          "id": "54820a58edd1e6a58918179c4213e9999c23a3bf",
          "message": "chip-ingress: set Authority header for gRPC connection (#1215)\n\n* chip-ingress: require port in address, set Authority header\n\n* chip-ingress: fix test\n\n* fix(lint): correctly set LOCAL_VERSION\n\n* beholder: fix test",
          "timestamp": "2025-05-28T13:30:47-05:00",
          "tree_id": "3ca534576ec7ace434daa0d41e04cec6a793fb51",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/54820a58edd1e6a58918179c4213e9999c23a3bf"
        },
        "date": 1748457124755,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 352.4,
            "unit": "ns/op",
            "extra": "3316093 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.5,
            "unit": "ns/op",
            "extra": "2749249 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28140,
            "unit": "ns/op",
            "extra": "41700 times\n4 procs"
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
          "id": "446fdc904af9c2f6909e023a546d0afb123df1f0",
          "message": "Fix a bug in logging, we expect raw bytes but try to be too fancy in the log function. (#1219)",
          "timestamp": "2025-05-29T12:13:39-04:00",
          "tree_id": "7f53253e0f2ae1b150ff0968b0a1ddb42471a369",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/446fdc904af9c2f6909e023a546d0afb123df1f0"
        },
        "date": 1748535303051,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 350.8,
            "unit": "ns/op",
            "extra": "3391430 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.2,
            "unit": "ns/op",
            "extra": "2953242 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28245,
            "unit": "ns/op",
            "extra": "42230 times\n4 procs"
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
          "id": "be9c134785d7ffd052bfa90eaf916db73fb04a70",
          "message": "pkg/utils: add CL_RUN_FLAKEY to run skipped flakey tests (#1137)",
          "timestamp": "2025-05-29T13:40:17-05:00",
          "tree_id": "cc9c6502b9cfa276849c91bd49bcacca02c695f8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/be9c134785d7ffd052bfa90eaf916db73fb04a70"
        },
        "date": 1748544093877,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.8,
            "unit": "ns/op",
            "extra": "3398186 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 402.8,
            "unit": "ns/op",
            "extra": "2800786 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28756,
            "unit": "ns/op",
            "extra": "41924 times\n4 procs"
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
          "id": "fc2f55458d617222b348d0fb86e432fcc486727b",
          "message": "Remove default logger in wasm to keep testing the same as normal execution. (#1222)",
          "timestamp": "2025-05-29T17:07:44-04:00",
          "tree_id": "0d1d72cc7d777cfb227864469fe88efe7c37cb8a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/fc2f55458d617222b348d0fb86e432fcc486727b"
        },
        "date": 1748552948482,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.2,
            "unit": "ns/op",
            "extra": "3358417 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 404.5,
            "unit": "ns/op",
            "extra": "2989242 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28480,
            "unit": "ns/op",
            "extra": "42111 times\n4 procs"
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
          "id": "b40f30ddfef17db95b39bcca445be76df9eab2b8",
          "message": "fix(capabilities): internally stop forwarding responses on unregister (#1221)",
          "timestamp": "2025-05-29T18:51:14-04:00",
          "tree_id": "c77b2ec177f828b91f66777b163f93f6e6713906",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/b40f30ddfef17db95b39bcca445be76df9eab2b8"
        },
        "date": 1748559158874,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.2,
            "unit": "ns/op",
            "extra": "3394626 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 401.7,
            "unit": "ns/op",
            "extra": "2967541 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28488,
            "unit": "ns/op",
            "extra": "42141 times\n4 procs"
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
          "id": "be37bd03a5670652b047fd668d41dc3bd7003384",
          "message": "added expected resources to info and requested limits to request (#1217)\n\n* added expected resources to info and requested limits to request\n\n* add resources to proto\n\n* fix request from proto test\n\n* add resource test to interface tests\n\n* rename properties to spend types\n\n* stronger types for spend types",
          "timestamp": "2025-05-30T13:29:28-05:00",
          "tree_id": "4bcc65014d49361d7cc37c4ad854ef42036e9860",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/be37bd03a5670652b047fd668d41dc3bd7003384"
        },
        "date": 1748629846227,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 353.7,
            "unit": "ns/op",
            "extra": "3420469 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.8,
            "unit": "ns/op",
            "extra": "2949956 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28455,
            "unit": "ns/op",
            "extra": "41486 times\n4 procs"
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
          "id": "7f68c6e25151377cac469917612d549a5e0da74f",
          "message": "pkg/http: move package http from core (#1227)",
          "timestamp": "2025-06-02T07:48:02-05:00",
          "tree_id": "4047f25d7476f355ca5bf32caa573072606facbb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7f68c6e25151377cac469917612d549a5e0da74f"
        },
        "date": 1748868561497,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.8,
            "unit": "ns/op",
            "extra": "3374414 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.7,
            "unit": "ns/op",
            "extra": "2917938 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28523,
            "unit": "ns/op",
            "extra": "42056 times\n4 procs"
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
          "id": "3c15a42d826633473c0ca0bbaf60e869e9f34c70",
          "message": "Plex 1458 Rearrange evm relayer service and chain capability packages and add a comment to callContract (#1218)\n\n* Update evm service CallContract comment\n\n* Fix evm capability service and package name\n\n* Change evm capability and service naming and rearrange packages\n\n* lint",
          "timestamp": "2025-06-02T16:19:24+02:00",
          "tree_id": "bc652bc3617147b431acb63a2e1ee168fe5a4dd5",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/3c15a42d826633473c0ca0bbaf60e869e9f34c70"
        },
        "date": 1748874057065,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354,
            "unit": "ns/op",
            "extra": "3400707 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.1,
            "unit": "ns/op",
            "extra": "2916122 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28500,
            "unit": "ns/op",
            "extra": "42076 times\n4 procs"
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
          "id": "8aa3f7dc56eaf05a74000464f73a38f82e9fa8cd",
          "message": "Plex 1458 part-3 - Remove wrappers in EVM service because of  evm capability sdk UX (#1223)\n\n* Update evm service CallContract comment\n\n* Fix evm capability service and package name\n\n* Change evm capability and service naming and rearrange packages\n\n* lint\n\n* Update evm service to not use proto wrappers for evm types\n\n* run generate",
          "timestamp": "2025-06-02T17:44:15+02:00",
          "tree_id": "3c6b4ccd3de035db37cc3051e16f2f40f15f7cd9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8aa3f7dc56eaf05a74000464f73a38f82e9fa8cd"
        },
        "date": 1748879135407,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.5,
            "unit": "ns/op",
            "extra": "3355270 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 430.4,
            "unit": "ns/op",
            "extra": "2939703 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28152,
            "unit": "ns/op",
            "extra": "42625 times\n4 procs"
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
          "id": "0a15f29b33ebdbe554ed6259ae9ae05e88055544",
          "message": "[CAPPL-736] Beholder Logger (#1229)\n\nA new logger object that combines logging and sending events to Beholder Client, following the existing custmsg methods.",
          "timestamp": "2025-06-03T05:39:58-05:00",
          "tree_id": "fa063913991c8a8bf03915913dbe87d5f12e8c91",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0a15f29b33ebdbe554ed6259ae9ae05e88055544"
        },
        "date": 1748947278847,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365.2,
            "unit": "ns/op",
            "extra": "3113136 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.5,
            "unit": "ns/op",
            "extra": "2842263 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28375,
            "unit": "ns/op",
            "extra": "42537 times\n4 procs"
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
          "id": "0dcaa6d2cc86ceec1ee38a681a31d41572a1bc04",
          "message": "Fix labels in Beholder Logger (#1232)",
          "timestamp": "2025-06-03T12:06:44-07:00",
          "tree_id": "3386dbd6b087cbb3f00e2dd983e20ec8662baf85",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/0dcaa6d2cc86ceec1ee38a681a31d41572a1bc04"
        },
        "date": 1748977686084,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.9,
            "unit": "ns/op",
            "extra": "3382779 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410,
            "unit": "ns/op",
            "extra": "2910066 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28148,
            "unit": "ns/op",
            "extra": "42675 times\n4 procs"
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
          "id": "12ebb766d8165eb642cec51a079ec4c2de2fbb77",
          "message": "[CRE-457] Add vault plugin type (#1228)",
          "timestamp": "2025-06-04T12:13:05+01:00",
          "tree_id": "9160eb98453afaa69be56653b86163042081b3d2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/12ebb766d8165eb642cec51a079ec4c2de2fbb77"
        },
        "date": 1749035666569,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 360.1,
            "unit": "ns/op",
            "extra": "3250615 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 424.7,
            "unit": "ns/op",
            "extra": "2624953 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28166,
            "unit": "ns/op",
            "extra": "42448 times\n4 procs"
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
          "id": "8a0b0d1e6a78d0a8d0ea17fcdb7b5924bebf04d6",
          "message": "fix(capabilities): pass separate context to stream (#1231)\n\n* fix(capabilities): pass separate context to stream\n\n* fix(capabilities): remove race on logger\n\n* fix: cleanup any existing calls for trigger ID\n\n* fix(capability): resolve race in test struct\n\n* chore(capability): update docstrings",
          "timestamp": "2025-06-04T12:17:24-04:00",
          "tree_id": "df1cd27849ab6ffa2ba61e797781d9c6fc2b5254",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8a0b0d1e6a78d0a8d0ea17fcdb7b5924bebf04d6"
        },
        "date": 1749053928472,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.9,
            "unit": "ns/op",
            "extra": "3403021 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.1,
            "unit": "ns/op",
            "extra": "2928416 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28280,
            "unit": "ns/op",
            "extra": "42688 times\n4 procs"
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
          "id": "d97e5f19c487a57e8e4d83f8bfddf6c90a7e7c0f",
          "message": "fix(capabilities): require got less than sent (#1235)",
          "timestamp": "2025-06-04T11:19:25-07:00",
          "tree_id": "1d0f1a3037102e906d7d8af6841d2b4af6e0a08b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d97e5f19c487a57e8e4d83f8bfddf6c90a7e7c0f"
        },
        "date": 1749061305380,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365.6,
            "unit": "ns/op",
            "extra": "3341049 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.4,
            "unit": "ns/op",
            "extra": "2898032 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28211,
            "unit": "ns/op",
            "extra": "42688 times\n4 procs"
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
          "id": "c4fb36f5716e26a7ff98f2224d2a6cff1bb05fc9",
          "message": "pkg/loop: expand EnvConfig and make available from Server (#1149)",
          "timestamp": "2025-06-04T21:27:36-05:00",
          "tree_id": "c6d9cd852c692868e21a0f9e386f49bcb4ab87bd",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c4fb36f5716e26a7ff98f2224d2a6cff1bb05fc9"
        },
        "date": 1749090540689,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 360.8,
            "unit": "ns/op",
            "extra": "3268242 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.4,
            "unit": "ns/op",
            "extra": "2899380 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28438,
            "unit": "ns/op",
            "extra": "42181 times\n4 procs"
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
          "id": "c062537be72caa2f892764b692095c58c743a674",
          "message": "check for nil map (#1239)",
          "timestamp": "2025-06-05T07:04:06-05:00",
          "tree_id": "0ef9657dab8ffe554791dd702db5c7549f31c20b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/c062537be72caa2f892764b692095c58c743a674"
        },
        "date": 1749125115542,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.6,
            "unit": "ns/op",
            "extra": "3382233 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.7,
            "unit": "ns/op",
            "extra": "2734956 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28415,
            "unit": "ns/op",
            "extra": "42157 times\n4 procs"
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
          "id": "1c982c45a39bdec733146d6b0c0f494ca9959a7a",
          "message": "use latest billing proto (#1237)",
          "timestamp": "2025-06-05T09:15:39-05:00",
          "tree_id": "56d0e2962cbb0dfc41dff49c6c64bd60af2dd86c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1c982c45a39bdec733146d6b0c0f494ca9959a7a"
        },
        "date": 1749133073203,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.1,
            "unit": "ns/op",
            "extra": "3398001 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.5,
            "unit": "ns/op",
            "extra": "2898825 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28707,
            "unit": "ns/op",
            "extra": "42105 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "12178754+anirudhwarrier@users.noreply.github.com",
            "name": "Anirudh Warrier",
            "username": "anirudhwarrier"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1974ede5e920bd7b7de433e43e635a43a70d9482",
          "message": "chore: bump freeport- support windows (#1241)",
          "timestamp": "2025-06-05T20:09:21+04:00",
          "tree_id": "0d6923d46e051c0918146110904c29c6b125fc02",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1974ede5e920bd7b7de433e43e635a43a70d9482"
        },
        "date": 1749139899924,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 382.1,
            "unit": "ns/op",
            "extra": "3214053 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 434.4,
            "unit": "ns/op",
            "extra": "2660202 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28487,
            "unit": "ns/op",
            "extra": "41540 times\n4 procs"
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
          "id": "95ba916687f9ca77228ef8259e87697bee938e24",
          "message": "adding timestamp to BaseMessage (#1238)\n\n* adding timestamp to BaseMessage\n\n* make generate",
          "timestamp": "2025-06-06T12:27:14-04:00",
          "tree_id": "8ab76198fb7073d9ef934d125b26ac370946c83b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/95ba916687f9ca77228ef8259e87697bee938e24"
        },
        "date": 1749227375375,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 381.4,
            "unit": "ns/op",
            "extra": "3308138 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.9,
            "unit": "ns/op",
            "extra": "2874422 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28933,
            "unit": "ns/op",
            "extra": "42158 times\n4 procs"
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
          "id": "bbf13d4e5c0428ed03830c37e5e9cc39e2484602",
          "message": "Seed random for setup and modes (#1236)",
          "timestamp": "2025-06-06T14:57:10-04:00",
          "tree_id": "e6b89b826d4e34831df20c61d2b128090cba5f5e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bbf13d4e5c0428ed03830c37e5e9cc39e2484602"
        },
        "date": 1749236293522,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357,
            "unit": "ns/op",
            "extra": "3378025 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 425.6,
            "unit": "ns/op",
            "extra": "2824537 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28633,
            "unit": "ns/op",
            "extra": "41594 times\n4 procs"
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
          "id": "a1830e9317e9616f8e8fa93f1e971d8408cb9034",
          "message": "Add nil underlying map check (#1243)",
          "timestamp": "2025-06-10T13:47:24+03:00",
          "tree_id": "11c9849e87e29ddfaabb9dfefe02860676398f17",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a1830e9317e9616f8e8fa93f1e971d8408cb9034"
        },
        "date": 1749552522083,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358,
            "unit": "ns/op",
            "extra": "3391441 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 435.2,
            "unit": "ns/op",
            "extra": "2911024 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28531,
            "unit": "ns/op",
            "extra": "42076 times\n4 procs"
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
          "id": "f0d618f73b03d75fead86953527525bdbc86eb2f",
          "message": "Add relayerset mock and extra validation to limitAndSort proto helpers (#1245)",
          "timestamp": "2025-06-10T14:13:03+02:00",
          "tree_id": "a317fbe75b70820e3d5e6f2d63457d6ff76898a8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f0d618f73b03d75fead86953527525bdbc86eb2f"
        },
        "date": 1749557661390,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.8,
            "unit": "ns/op",
            "extra": "3361134 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.5,
            "unit": "ns/op",
            "extra": "2879912 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28539,
            "unit": "ns/op",
            "extra": "42132 times\n4 procs"
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
          "id": "f86489c37b7a036a47575f42267406ce2c3ad91a",
          "message": "pkg/loop/internal/relayer: include external job ID for unique names (#1246)",
          "timestamp": "2025-06-10T07:35:30-05:00",
          "tree_id": "d7d5450cbec6ac2b1f9ae45a33b3e9f48b5f7dd3",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f86489c37b7a036a47575f42267406ce2c3ad91a"
        },
        "date": 1749559018506,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359,
            "unit": "ns/op",
            "extra": "3344528 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 422.7,
            "unit": "ns/op",
            "extra": "2786971 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 29260,
            "unit": "ns/op",
            "extra": "42126 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "96362174+chainchad@users.noreply.github.com",
            "name": "chainchad",
            "username": "chainchad"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "86a4a7db83ee2e9a3ec216bcadc99ed5c4f3f92c",
          "message": "Create stale PR workflow (#1242)",
          "timestamp": "2025-06-10T11:29:55-04:00",
          "tree_id": "20ac6d153890f37570749d0674757446d87bd6b7",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/86a4a7db83ee2e9a3ec216bcadc99ed5c4f3f92c"
        },
        "date": 1749569492912,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356,
            "unit": "ns/op",
            "extra": "3381406 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 436.4,
            "unit": "ns/op",
            "extra": "2945020 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28429,
            "unit": "ns/op",
            "extra": "42128 times\n4 procs"
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
          "id": "67d52ef3a68393299043d029e81e8f03f51b7da7",
          "message": "Update GetId to GeID in module.go, core fails lint otherwise (#1248)",
          "timestamp": "2025-06-10T13:11:47-04:00",
          "tree_id": "b3fa1047c964888264c42a66a2d413f936db8c4b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/67d52ef3a68393299043d029e81e8f03f51b7da7"
        },
        "date": 1749575582968,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 365.9,
            "unit": "ns/op",
            "extra": "3268468 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.8,
            "unit": "ns/op",
            "extra": "2913840 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28495,
            "unit": "ns/op",
            "extra": "42234 times\n4 procs"
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
          "id": "26e78071ce46e347cf2fc8256b707b856279877e",
          "message": "requests handling (#1247)\n\nresponse handling\n\nmoved consensus requests handling out of ocr3 package",
          "timestamp": "2025-06-11T11:47:23+01:00",
          "tree_id": "e07cffa630ee5ad374d11373b864e9cec6f57363",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/26e78071ce46e347cf2fc8256b707b856279877e"
        },
        "date": 1749638918551,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.3,
            "unit": "ns/op",
            "extra": "3308883 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 439.2,
            "unit": "ns/op",
            "extra": "2913926 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28479,
            "unit": "ns/op",
            "extra": "42080 times\n4 procs"
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
          "id": "eb34d3bf1a6492a7cfc1f51465e34e38f8942105",
          "message": "Allow consensus on time type, fix identical consensus big int, and rename the tag (#1244)\n\n* Allow consensus on time type, fix identical consensus big int, and rename the tag\n\n* change value's time test to use nanos ensuring correct encoding",
          "timestamp": "2025-06-11T10:22:43-04:00",
          "tree_id": "2c51e6128820013cc8380b34e3a73e73ad0e043a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/eb34d3bf1a6492a7cfc1f51465e34e38f8942105"
        },
        "date": 1749651838558,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.3,
            "unit": "ns/op",
            "extra": "3392124 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.2,
            "unit": "ns/op",
            "extra": "2928292 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28492,
            "unit": "ns/op",
            "extra": "42141 times\n4 procs"
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
          "id": "a4eee159cd08fbdf69082e8ad4dfacdf73064e7a",
          "message": "pkg/beholder/beholdertest: move beholder test utils to separate package (#1250)",
          "timestamp": "2025-06-12T08:57:02-05:00",
          "tree_id": "d7930daa7df51ba9ea6e513a80ccfe621a89f6d2",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a4eee159cd08fbdf69082e8ad4dfacdf73064e7a"
        },
        "date": 1749736694015,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.4,
            "unit": "ns/op",
            "extra": "3394702 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.8,
            "unit": "ns/op",
            "extra": "2910624 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28432,
            "unit": "ns/op",
            "extra": "42097 times\n4 procs"
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
          "id": "676553ada522d44501139177f707fa7a0d4a2324",
          "message": "Bump chainlink-protos/billing to 1c32d2efe48faca8f2158117fdde57e137cacb13 (#1260)",
          "timestamp": "2025-06-12T17:42:55-07:00",
          "tree_id": "78df906a8708b6fd831f2f39991803434ca8ce75",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/676553ada522d44501139177f707fa7a0d4a2324"
        },
        "date": 1749775504462,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358,
            "unit": "ns/op",
            "extra": "3336242 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.1,
            "unit": "ns/op",
            "extra": "2875906 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28498,
            "unit": "ns/op",
            "extra": "42112 times\n4 procs"
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
          "id": "bd3977b4bb4982f042b61fb2e6a70b5b42ef93f7",
          "message": "chip-ingress: support TLS with HTTP/2 for gRPC client (#1261)\n\nCo-authored-by: Jordan Krage <jmank88@gmail.com>",
          "timestamp": "2025-06-13T16:27:55-04:00",
          "tree_id": "46effed2fb118e16c5bb8247384e0dc85e409e95",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/bd3977b4bb4982f042b61fb2e6a70b5b42ef93f7"
        },
        "date": 1749846541331,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.4,
            "unit": "ns/op",
            "extra": "3377876 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.8,
            "unit": "ns/op",
            "extra": "2937744 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28146,
            "unit": "ns/op",
            "extra": "42652 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "kiryll.kuzniecow@gmail.com",
            "name": "Kiryll Kuzniecow",
            "username": "kirqz23"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8fe61292a55018bd8e2bdc27135170da8b6097fb",
          "message": "INFOPLAT-2282: Add beholderNoopLogerProvider and LogStreamingEnabled flag to control emitting logs (#1263)\n\n* Add beholderNoopLogerProvider and LogStreamingEnabled flag to control emitting logs\n\n* Fix tests\n\n* Fix config test\n\n* Flatten if conditions\n\n* Fix tests",
          "timestamp": "2025-06-16T11:46:53+02:00",
          "tree_id": "f1d22d434fdaa1aa2cc4febb7392c7881703568d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/8fe61292a55018bd8e2bdc27135170da8b6097fb"
        },
        "date": 1750067277304,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 352.9,
            "unit": "ns/op",
            "extra": "3355990 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.2,
            "unit": "ns/op",
            "extra": "2935824 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28583,
            "unit": "ns/op",
            "extra": "42142 times\n4 procs"
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
          "id": "a62474121a947956c421dd029059050d6398a394",
          "message": "Add txhash to sequences (#1254)\n\n* Add txhash to sequences\n\n* Update contract reader sequence proto",
          "timestamp": "2025-06-16T15:14:19+02:00",
          "tree_id": "b263ac597a57b5049aa4ae749b70a5693074ea0c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a62474121a947956c421dd029059050d6398a394"
        },
        "date": 1750079740405,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.3,
            "unit": "ns/op",
            "extra": "3336087 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.9,
            "unit": "ns/op",
            "extra": "2860257 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28814,
            "unit": "ns/op",
            "extra": "42159 times\n4 procs"
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
          "id": "1b60a6dfce5e90e9234ec6d68344aec317900971",
          "message": "beholder: relax validation for beholder_data_schema required attribute (#1267)",
          "timestamp": "2025-06-16T11:22:52-04:00",
          "tree_id": "49dbfcd87f7cbee35379c492160ced8f6f5be61d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1b60a6dfce5e90e9234ec6d68344aec317900971"
        },
        "date": 1750087444123,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.3,
            "unit": "ns/op",
            "extra": "3352372 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.7,
            "unit": "ns/op",
            "extra": "2885444 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28523,
            "unit": "ns/op",
            "extra": "42145 times\n4 procs"
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
          "id": "db6559760098595e7d52692f9a41a725dd69727c",
          "message": "Update SDK per the mental model (#1266)",
          "timestamp": "2025-06-16T12:52:19-04:00",
          "tree_id": "9d810f058ba3f646a896869ebffb312b8e3b6758",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/db6559760098595e7d52692f9a41a725dd69727c"
        },
        "date": 1750092809343,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 360.4,
            "unit": "ns/op",
            "extra": "3335973 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.9,
            "unit": "ns/op",
            "extra": "2582714 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28510,
            "unit": "ns/op",
            "extra": "42097 times\n4 procs"
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
          "id": "9839ff5867ae4d11f0cec44805b0691ee6365ce7",
          "message": "generate SDK for HTTP action capability (#1234)",
          "timestamp": "2025-06-16T15:27:19-04:00",
          "tree_id": "b958e6a3108403205c7186a68d4d982478f10a4a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9839ff5867ae4d11f0cec44805b0691ee6365ce7"
        },
        "date": 1750102113062,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 368.2,
            "unit": "ns/op",
            "extra": "3244684 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 418.4,
            "unit": "ns/op",
            "extra": "2867030 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28498,
            "unit": "ns/op",
            "extra": "42126 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "165708424+pavel-raykov@users.noreply.github.com",
            "name": "pavel-raykov",
            "username": "pavel-raykov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9749819c00719d565c4219166d1e54fdbc97d546",
          "message": "Minor. (#1272)",
          "timestamp": "2025-06-17T20:16:50+02:00",
          "tree_id": "eac7bb76380433843e01e1627300358ae966d275",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9749819c00719d565c4219166d1e54fdbc97d546"
        },
        "date": 1750184292175,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.5,
            "unit": "ns/op",
            "extra": "3369385 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.8,
            "unit": "ns/op",
            "extra": "2920632 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28462,
            "unit": "ns/op",
            "extra": "42199 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "juan.lautarofernandez@smartcontract.com",
            "name": "Juan Lautaro Fernandez",
            "username": "fernandezlautaro"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6a1496bbe01100e89d49f0e2de4775d76d5bc69b",
          "message": "PLEX-1436: LogTrigger implementation (#1190)\n\n* PLEX-1436: LogTrigger POC implementation\n\n* PLEX-1461: trigger sends one log at a time to distinguish their IDs\n\n* rebasing to master take 1\n\n* adjusting code to latest API changes\n\n* updating API to have confidence of SAFE, LATEST, FINALIZED\n\n* rebasing w main",
          "timestamp": "2025-06-17T22:40:59-03:00",
          "tree_id": "fa97521e80fc2deb7f200b63d7e8e9d7eeac6fb6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/6a1496bbe01100e89d49f0e2de4775d76d5bc69b"
        },
        "date": 1750210993663,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 366.8,
            "unit": "ns/op",
            "extra": "3267399 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 432.4,
            "unit": "ns/op",
            "extra": "2892073 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28507,
            "unit": "ns/op",
            "extra": "42166 times\n4 procs"
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
          "id": "03012f597e11d4a4901df3b437ce869ac10c30f0",
          "message": "introduce gatewayConnector and gatewayConnectorHandler as gRPC services (#1256)\n\n* remove start and close methods. use bytes instead of gateway message type over grpc\n\n* Update pkg/types/core/gateway_connector.go\n\n* rename sign to sign message",
          "timestamp": "2025-06-18T11:44:20+09:00",
          "tree_id": "2b10b49f7882649280053f00b64d8e6967de591e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/03012f597e11d4a4901df3b437ce869ac10c30f0"
        },
        "date": 1750214736225,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.3,
            "unit": "ns/op",
            "extra": "3378132 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.2,
            "unit": "ns/op",
            "extra": "2882562 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28523,
            "unit": "ns/op",
            "extra": "41012 times\n4 procs"
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
          "id": "a5a42ee8701be96574c4c416f6c2e2e3993558f4",
          "message": "Updated ocr3 Metadata type to include Encoding and Decoding and updated common beholder code  (#1160)\n\n* Updated ocr3 Metadata type to include Encoding and Decoding\n\n* Revert \"add GetEstimateFee (#1196)\"\n\nThis reverts commit 80bc8b13c0e7839849182ef9e55b8ac0063145c9.\n\n* Reapply \"add GetEstimateFee (#1196)\"\n\nThis reverts commit f3ed96e428724590d600ab73b9c6be5fce146d10.\n\n* addressed feedback\n\n* Moved proto emitter and helpers to common\n\n* Published common attritutes\n\n* addressed feedback\n\n* added beholder attribute data_type\n\n* removed behodler for attr keys'\n\n* Revert \"pkg/loop: expand EnvConfig and make available from Server (#1149)\"\n\nThis reverts commit c4fb36f5716e26a7ff98f2224d2a6cff1bb05fc9.\n\n* Reapply \"pkg/loop: expand EnvConfig and make available from Server (#1149)\"\n\nThis reverts commit 1ff9a4634a1f7bbf2ffcaaec8d038587689e8857.\n\n* Revert \"Seed random for setup and modes (#1236)\"\n\nThis reverts commit bbf13d4e5c0428ed03830c37e5e9cc39e2484602.\n\n* Reapply \"Seed random for setup and modes (#1236)\"\n\nThis reverts commit 54755c8caea7637e627f489b1603c74b1d80cd1d.\n\n* Revert \"requests handling (#1247)\"\n\nThis reverts commit 26e78071ce46e347cf2fc8256b707b856279877e.\n\n* Reapply \"requests handling (#1247)\"\n\nThis reverts commit 00a1fdb4172af042269eefc66880b0b900ab89d4.\n\n* Made ToSchemaFullName public",
          "timestamp": "2025-06-18T12:28:08-04:00",
          "tree_id": "f233f8abaede6dfae23cf0da119120222eba5e2d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a5a42ee8701be96574c4c416f6c2e2e3993558f4"
        },
        "date": 1750264156287,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.1,
            "unit": "ns/op",
            "extra": "3399758 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.2,
            "unit": "ns/op",
            "extra": "2908112 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28480,
            "unit": "ns/op",
            "extra": "42189 times\n4 procs"
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
          "id": "534622ec8df55de35cac25eb10724aa5ba7a51cd",
          "message": "define SDK for HTTP trigger (#1240)\n\n* define SDK for HTTP trigger\n\n---------\n\nCo-authored-by: Akhil Chainani <akhil.chainani1@gmail.com>",
          "timestamp": "2025-06-19T02:39:55+09:00",
          "tree_id": "060f0168c4d95d071f29df4cc05ae5d1bb53ca5c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/534622ec8df55de35cac25eb10724aa5ba7a51cd"
        },
        "date": 1750268537990,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.8,
            "unit": "ns/op",
            "extra": "3362640 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 422.3,
            "unit": "ns/op",
            "extra": "2885719 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28625,
            "unit": "ns/op",
            "extra": "42123 times\n4 procs"
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
          "id": "1d4be6761facb516d2227dbf91b586624a2e6ec5",
          "message": "move common gateway code from core (#1257)\n\n* remove start and close methods. use bytes instead of gateway message type over grpc\n\n* move common gateway code from core for reuse in capabilities repo\n\n* Revert \"remove start and close methods. use bytes instead of gateway message type over grpc\"\n\nThis reverts commit db3ab142e33132a61ae772c9ffd4b64088958077.\n\n* remove gateway message from here\n\n* move rate limiter to a top level pkg\n\n* Update pkg/types/gateway/round_robin_selector.go\n\nCo-authored-by: Jordan Krage <jmank88@gmail.com>\n\n* gomodtidy\n\n---------\n\nCo-authored-by: Jordan Krage <jmank88@gmail.com>\nCo-authored-by: Bolek <1416262+bolekk@users.noreply.github.com>",
          "timestamp": "2025-06-18T11:19:27-07:00",
          "tree_id": "1eb858ccdb8bfbc88f0278791e02184a64548259",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/1d4be6761facb516d2227dbf91b586624a2e6ec5"
        },
        "date": 1750270835183,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.6,
            "unit": "ns/op",
            "extra": "3392058 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412,
            "unit": "ns/op",
            "extra": "2924869 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28458,
            "unit": "ns/op",
            "extra": "42118 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "34992934+prashantkumar1982@users.noreply.github.com",
            "name": "Prashant Yadav",
            "username": "prashantkumar1982"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "66edcc4a4449aa41692cbff2275ea1d7bb91c5ed",
          "message": "Add a generic jsonrpc2 library (#1279)",
          "timestamp": "2025-06-19T18:01:34+03:00",
          "tree_id": "144de09b8eb259ad9e66b004501a1c817df41e5f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/66edcc4a4449aa41692cbff2275ea1d7bb91c5ed"
        },
        "date": 1750345366735,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.8,
            "unit": "ns/op",
            "extra": "3345994 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.4,
            "unit": "ns/op",
            "extra": "2868384 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28540,
            "unit": "ns/op",
            "extra": "41173 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "16602512+krehermann@users.noreply.github.com",
            "name": "krehermann",
            "username": "krehermann"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a209f3051bd4703bd157e2a842d4a241b6c06d89",
          "message": "add configtest helpers (#1274)\n\n* add configtest helpers\n\n* address feedback",
          "timestamp": "2025-06-19T17:43:27Z",
          "tree_id": "2cb0c92fc5d7925085af82222d7577248f248ec4",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a209f3051bd4703bd157e2a842d4a241b6c06d89"
        },
        "date": 1750355081228,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 362.6,
            "unit": "ns/op",
            "extra": "3333747 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.3,
            "unit": "ns/op",
            "extra": "2902093 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28524,
            "unit": "ns/op",
            "extra": "42034 times\n4 procs"
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
          "id": "cc40cf268dfdbfffaeeeecea8c0903673b013f50",
          "message": "Have consensus tags respect mapstructure use in values.Value (#1252)\n\nCo-authored-by: Justin Kaseman <justinkaseman@live.com>",
          "timestamp": "2025-06-19T23:24:42Z",
          "tree_id": "3a214bff157253010a59bea3007e1a300ce6f691",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/cc40cf268dfdbfffaeeeecea8c0903673b013f50"
        },
        "date": 1750375546103,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.2,
            "unit": "ns/op",
            "extra": "3372525 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.9,
            "unit": "ns/op",
            "extra": "2893765 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28489,
            "unit": "ns/op",
            "extra": "42009 times\n4 procs"
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
          "id": "13f3ead5fedbb4d531f240eac9b85f289b1f4bc2",
          "message": "[CAPPL-782] Handle user logs independently of system logs (#1278)",
          "timestamp": "2025-06-20T08:07:18-07:00",
          "tree_id": "ab558c522d758d28ea8f54454066aadfc5672422",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/13f3ead5fedbb4d531f240eac9b85f289b1f4bc2"
        },
        "date": 1750432115612,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357,
            "unit": "ns/op",
            "extra": "3359937 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 420.8,
            "unit": "ns/op",
            "extra": "2888158 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28491,
            "unit": "ns/op",
            "extra": "42187 times\n4 procs"
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
          "id": "ef933e8e86fede6855f1cb142c2aea998464aac5",
          "message": "Use timestamp in cron, seperate legacy call from original (#1268)",
          "timestamp": "2025-06-20T08:31:51-07:00",
          "tree_id": "30107ae04421bb85501b62cbd43b7494bab5718c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ef933e8e86fede6855f1cb142c2aea998464aac5"
        },
        "date": 1750433586104,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.4,
            "unit": "ns/op",
            "extra": "3360174 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.1,
            "unit": "ns/op",
            "extra": "2887294 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28552,
            "unit": "ns/op",
            "extra": "41066 times\n4 procs"
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
          "id": "9469184bca40b195b5ef05754c30779cee1c1b3f",
          "message": "[CRE-446] Implement SDK support for GetSecret (#1270)\n\n* [CRE-446] Implement SDK support for GetSecret\n\n* Use timestamp in cron, seperate legacy call from original (#1268)\n\n* [fix] Address review comments:\n- wcx > env\n- remove TODO\n- disallow '/' in SetSecret\n\n---------\n\nCo-authored-by: Ryan Tinianov <tinianov@live.com>",
          "timestamp": "2025-06-23T09:22:51Z",
          "tree_id": "e36849b90ce54f657e79e9ffe4c3b6f490f1a98d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9469184bca40b195b5ef05754c30779cee1c1b3f"
        },
        "date": 1750670701665,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.3,
            "unit": "ns/op",
            "extra": "3382768 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 420.7,
            "unit": "ns/op",
            "extra": "2871627 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28562,
            "unit": "ns/op",
            "extra": "42214 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "brunotm@gmail.com",
            "name": "Bruno Moura",
            "username": "brunotm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9cb7ec4a4def677580561176cf6e6609ca13aea2",
          "message": "pkg/types/llo: add evm_abi_encode_unpacked_expr report format and calculated aggregator (#1259)",
          "timestamp": "2025-06-23T14:01:16Z",
          "tree_id": "87110fe353510307695506f8c16b202f34f85a9d",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9cb7ec4a4def677580561176cf6e6609ca13aea2"
        },
        "date": 1750687341926,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.5,
            "unit": "ns/op",
            "extra": "3345838 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 415.4,
            "unit": "ns/op",
            "extra": "2890928 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28764,
            "unit": "ns/op",
            "extra": "42026 times\n4 procs"
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
          "id": "d03456a6d046a6e01f6a51561ad12613981f76c2",
          "message": "Lower cron capability version to 1.0.0 (#1286)",
          "timestamp": "2025-06-23T10:57:59-06:00",
          "tree_id": "179d680a54b5091ca35502b43dda142bc6ad432c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d03456a6d046a6e01f6a51561ad12613981f76c2"
        },
        "date": 1750697951565,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.7,
            "unit": "ns/op",
            "extra": "3339492 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413,
            "unit": "ns/op",
            "extra": "2899928 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28930,
            "unit": "ns/op",
            "extra": "42116 times\n4 procs"
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
          "id": "a17cdfe27dfd738c6c900a28243b6332c8f0d24c",
          "message": "add gateway connector to standard capabilities interface (#1258)\n\n* add gateway connector to standard capabilities\n\n* gateway connector and handler uses jsonrpc structs instead of []byte",
          "timestamp": "2025-06-23T17:58:19Z",
          "tree_id": "e56f3a43f0649c6000be330fb16b5b76d69cfb25",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/a17cdfe27dfd738c6c900a28243b6332c8f0d24c"
        },
        "date": 1750701592121,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 353.1,
            "unit": "ns/op",
            "extra": "3393423 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 419.4,
            "unit": "ns/op",
            "extra": "2883906 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28476,
            "unit": "ns/op",
            "extra": "40442 times\n4 procs"
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
          "id": "93f383781b0aba88ded6a303e7744af02dcd213e",
          "message": "Update to use chainlink protos for CRE sdk, the CRE generator, and values.Value. (#1281)",
          "timestamp": "2025-06-24T18:10:23+02:00",
          "tree_id": "ce5a7e51257176db4db36e5fd88bf729f7d3aa5e",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/93f383781b0aba88ded6a303e7744af02dcd213e"
        },
        "date": 1750781562713,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 358.8,
            "unit": "ns/op",
            "extra": "3338826 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 413.7,
            "unit": "ns/op",
            "extra": "2917545 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28242,
            "unit": "ns/op",
            "extra": "42625 times\n4 procs"
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
          "id": "f5267b8c1e7c32f1b7c2e180085520f4880cb21f",
          "message": "Remove replace for pkg/values and use the last commit (#1292)",
          "timestamp": "2025-06-24T12:29:59-05:00",
          "tree_id": "e8c459e8e86fd538ca6c4828248d1bdca588ce73",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f5267b8c1e7c32f1b7c2e180085520f4880cb21f"
        },
        "date": 1750786334242,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357,
            "unit": "ns/op",
            "extra": "3359463 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.6,
            "unit": "ns/op",
            "extra": "2937870 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28156,
            "unit": "ns/op",
            "extra": "42526 times\n4 procs"
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
          "id": "7566a2b110f13c6598ececbf1e16176b9ccb46b3",
          "message": "Create a standard test suite that all languages can run against to verify host-guest interactions (#1284)",
          "timestamp": "2025-06-24T12:33:10-07:00",
          "tree_id": "36041a52599b3b159f94be2272eabfa4ca72aa2c",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/7566a2b110f13c6598ececbf1e16176b9ccb46b3"
        },
        "date": 1750793664004,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.2,
            "unit": "ns/op",
            "extra": "3198686 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.1,
            "unit": "ns/op",
            "extra": "2911370 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28140,
            "unit": "ns/op",
            "extra": "42685 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "juan.lautarofernandez@smartcontract.com",
            "name": "Juan Lautaro Fernandez",
            "username": "fernandezlautaro"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "204c365e1e93f1cd9c9da49da2e00c06b8a59433",
          "message": "PLEX-1513: update LogTrigger topics API (#1287)\n\n* PLEX-1513: update LogTrigger topics API\n\n* PLEX-1513: update LogTrigger topics API\n\n* generate protos",
          "timestamp": "2025-06-25T11:45:58+01:00",
          "tree_id": "06846a276d3a5787cdcd2a52e61c48c59bd7700f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/204c365e1e93f1cd9c9da49da2e00c06b8a59433"
        },
        "date": 1750848419336,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 353.5,
            "unit": "ns/op",
            "extra": "3357742 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 404.3,
            "unit": "ns/op",
            "extra": "2948400 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28226,
            "unit": "ns/op",
            "extra": "42696 times\n4 procs"
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
          "id": "9ba03877f50b39d96a281606b8fc81d5e39ffa5d",
          "message": "Rename On to Handler in CRE SDK (#1296)",
          "timestamp": "2025-06-25T10:04:35-04:00",
          "tree_id": "2f956476f3472497c332d90f03058ba24e1ff16b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/9ba03877f50b39d96a281606b8fc81d5e39ffa5d"
        },
        "date": 1750860337882,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.1,
            "unit": "ns/op",
            "extra": "3316201 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 416.3,
            "unit": "ns/op",
            "extra": "2859110 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28367,
            "unit": "ns/op",
            "extra": "42562 times\n4 procs"
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
          "id": "5648b4dd059b693d790a8db3f24df228241eab7c",
          "message": "More verbose error message for incompatible trigger event protos (#1294)",
          "timestamp": "2025-06-25T15:20:10Z",
          "tree_id": "13558e85b4fe084ba0aa5cdc06653c93507fc0ee",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5648b4dd059b693d790a8db3f24df228241eab7c"
        },
        "date": 1750864870777,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 367.1,
            "unit": "ns/op",
            "extra": "3289532 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 405.7,
            "unit": "ns/op",
            "extra": "2966493 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28300,
            "unit": "ns/op",
            "extra": "42404 times\n4 procs"
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
          "id": "5613187324ad638bb25fb5ce65ddc515bc454c25",
          "message": "Improve Beholder Logger (#1280)\n\n1. Don't drop errors\n2. Protect against blocking calls",
          "timestamp": "2025-06-25T17:05:38Z",
          "tree_id": "250680b1bd2d98d4bdac7ce147492ec080410be6",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/5613187324ad638bb25fb5ce65ddc515bc454c25"
        },
        "date": 1750871198732,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.4,
            "unit": "ns/op",
            "extra": "2860522 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.6,
            "unit": "ns/op",
            "extra": "2920566 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28417,
            "unit": "ns/op",
            "extra": "41425 times\n4 procs"
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
          "id": "e50b2e7ffe2da8e1e6e010947313801e363343a1",
          "message": "[chore] Bump protos (#1300)",
          "timestamp": "2025-06-26T14:12:12Z",
          "tree_id": "ce8a0d3702d1472c1a484149e55b0e0d0f59a2fa",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e50b2e7ffe2da8e1e6e010947313801e363343a1"
        },
        "date": 1750947265966,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 357.1,
            "unit": "ns/op",
            "extra": "3365604 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.4,
            "unit": "ns/op",
            "extra": "2849208 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28142,
            "unit": "ns/op",
            "extra": "42199 times\n4 procs"
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
          "id": "4322119b019c718e2b11f3acf8821da5d66c61d9",
          "message": "[PRIV-83] Add vault and SDK protos (#1290)\n\n- Add vault protos\n- Update chainlink-protos to incorporate a change where we add ID\ninformation to errors.\n- Fix bug where the maxResponseSize wasn't passed into the runner during\nthe subscription phase.",
          "timestamp": "2025-06-26T18:31:45+02:00",
          "tree_id": "b8608a0d978735fe4cc1135c6b2bb4ddff12f814",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/4322119b019c718e2b11f3acf8821da5d66c61d9"
        },
        "date": 1750955641194,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.9,
            "unit": "ns/op",
            "extra": "3384406 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.1,
            "unit": "ns/op",
            "extra": "2930146 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28242,
            "unit": "ns/op",
            "extra": "42674 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "pablolagreca@hotmail.com",
            "name": "pablolagreca",
            "username": "pablolagreca"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d07b85c7c0872e21c1e904050c05014c0eb03583",
          "message": "PLEX-250 - WriteReport initial implementation (#1264)\n\n* PLEX-250 - WriteReport initial implementation\n\n* all tests passing\n\n* adding back get tx result\n\n* applying merge to main\n\n* updaing gen code\n\n* adding back commented code and adding TODO item\n\n* updating obs go mods\n\n* rebase\n\n* reverting addition of confirmed state\n\n* changing TxStatus.SUCCESS so it is not the default int value\n\nchanging TxStatus enums so success is not 0\n\nimproving docs\n\n* running latest version of cre-protogen\n\nfixing cap code\n\n* pr feedback",
          "timestamp": "2025-06-26T16:49:37-07:00",
          "tree_id": "1f96530a1f7ca6f15615ac4964b510f1d2afb977",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/d07b85c7c0872e21c1e904050c05014c0eb03583"
        },
        "date": 1750981858315,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 352.9,
            "unit": "ns/op",
            "extra": "3394848 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.4,
            "unit": "ns/op",
            "extra": "2921973 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28136,
            "unit": "ns/op",
            "extra": "42675 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "34992934+prashantkumar1982@users.noreply.github.com",
            "name": "Prashant Yadav",
            "username": "prashantkumar1982"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2cbb7418aaa59b90512dd88ddb4a882743fc20eb",
          "message": "jsonrpc2 library simplification (#1293)\n\n* Make auth optional\n\n* refactoring\n\n* json rpc library refactoring",
          "timestamp": "2025-06-27T00:29:29Z",
          "tree_id": "4bde46fcee16b3382a3859e83b03162031328cb9",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/2cbb7418aaa59b90512dd88ddb4a882743fc20eb"
        },
        "date": 1750984254134,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.1,
            "unit": "ns/op",
            "extra": "3399536 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410,
            "unit": "ns/op",
            "extra": "2935530 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28180,
            "unit": "ns/op",
            "extra": "42651 times\n4 procs"
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
          "id": "95f07da6c2df3b27060878f2e8df4ac6431bad00",
          "message": "[PRIV-77] Mock implementation of the Vault (#1304)",
          "timestamp": "2025-06-27T12:32:42Z",
          "tree_id": "496465ce7b814cf8dc2206454248076d62cf691b",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/95f07da6c2df3b27060878f2e8df4ac6431bad00"
        },
        "date": 1751027644337,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 359.1,
            "unit": "ns/op",
            "extra": "3362709 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 417.3,
            "unit": "ns/op",
            "extra": "2859624 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28233,
            "unit": "ns/op",
            "extra": "42694 times\n4 procs"
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
          "id": "ed6ed7b7fcd76d8e81925ad0a9e0ed895629ec12",
          "message": "Copy EVM protos to CRE so it can stand alone, remove QueryTrackedLogsRequest as it was going to be removed soon anyway (#1291)",
          "timestamp": "2025-06-27T16:34:34+01:00",
          "tree_id": "b88de1452a6b27a4a40ea43e6f129a5fff1de69a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/ed6ed7b7fcd76d8e81925ad0a9e0ed895629ec12"
        },
        "date": 1751038548827,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.9,
            "unit": "ns/op",
            "extra": "3335142 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 411.6,
            "unit": "ns/op",
            "extra": "2913228 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28174,
            "unit": "ns/op",
            "extra": "42646 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "165708424+pavel-raykov@users.noreply.github.com",
            "name": "pavel-raykov",
            "username": "pavel-raykov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "731d426aa846bc036d3fe9e923dbdad247b90d50",
          "message": "Fix chainlink-common version. (#1305)",
          "timestamp": "2025-06-27T16:55:44Z",
          "tree_id": "d6cc124390a4a10fe0bdf82cfc1f6bdd76b553ac",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/731d426aa846bc036d3fe9e923dbdad247b90d50"
        },
        "date": 1751043474960,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 363.6,
            "unit": "ns/op",
            "extra": "3382960 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 414.2,
            "unit": "ns/op",
            "extra": "2872814 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28436,
            "unit": "ns/op",
            "extra": "42637 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "165708424+pavel-raykov@users.noreply.github.com",
            "name": "pavel-raykov",
            "username": "pavel-raykov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "836c9fc4f3925e171c605323b33d1a99d494950f",
          "message": "[CRE-494] Use same API in RelayerSet and Relayer (#1295)",
          "timestamp": "2025-06-30T15:41:37+02:00",
          "tree_id": "b3e92b8ba45762274fa30b0a68ba711bc0da16a8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/836c9fc4f3925e171c605323b33d1a99d494950f"
        },
        "date": 1751290970509,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 367.2,
            "unit": "ns/op",
            "extra": "3267613 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 423.5,
            "unit": "ns/op",
            "extra": "2846156 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28560,
            "unit": "ns/op",
            "extra": "41900 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "42331373+hendoxc@users.noreply.github.com",
            "name": "Hagen H",
            "username": "hendoxc"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "49f54f1c379d380b0279045068efb9a8f2b0a132",
          "message": "Update chip ingress client (#1306)\n\n* Adds `NewEventWithAttributes`\n\n- marks `NewEvent` as deprecated\n\n* Removes default cfg func\n\nRemoves default cfg func\n\n* Upgrades cloud events version\n\n* Regenerate chip protos\n\n- adds generate.go\n\n* Makes removes grpc client wrapper indirection\n\n* Updates chip-ingress in beholder client\n\nRemoves mockery entry\n\nFixes test\n\nRuns `fmt`\n\n* Removes generate.go",
          "timestamp": "2025-06-30T13:08:04-04:00",
          "tree_id": "1a8c8bed8d3091f7d393c3b7691aeeb943214820",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/49f54f1c379d380b0279045068efb9a8f2b0a132"
        },
        "date": 1751303412074,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 361.6,
            "unit": "ns/op",
            "extra": "3383908 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 404.8,
            "unit": "ns/op",
            "extra": "2957082 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28412,
            "unit": "ns/op",
            "extra": "42104 times\n4 procs"
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
          "id": "f216eaa9aa54fc23a6b1f90586bb5941803baba3",
          "message": "Add beholder metrics helper struct and GetChainInfo to relayer and relayerset (#1276)\n\n* Add beholder metrics helper struct\n\n* Add NewFloat64Histogram to MetricInfo\n\n* Add NewInt64Histogram\n\n* Add GetChainInfo to relayer and relayerset",
          "timestamp": "2025-06-30T20:00:21+02:00",
          "tree_id": "d92bfb884155e0545f89d75d2c310d9498dd98bb",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/f216eaa9aa54fc23a6b1f90586bb5941803baba3"
        },
        "date": 1751306497364,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.5,
            "unit": "ns/op",
            "extra": "3396824 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 409.2,
            "unit": "ns/op",
            "extra": "2888931 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28127,
            "unit": "ns/op",
            "extra": "42693 times\n4 procs"
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
          "id": "e903795cfa47ba37b5114564fad91ef488c29563",
          "message": "add cache settings to http action. introduce gateway types for http trigger (#1301)",
          "timestamp": "2025-06-30T22:01:42Z",
          "tree_id": "734ca4b1e438194b0af748977473e2208a67f8dc",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/e903795cfa47ba37b5114564fad91ef488c29563"
        },
        "date": 1751320977046,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 355.7,
            "unit": "ns/op",
            "extra": "3417962 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.7,
            "unit": "ns/op",
            "extra": "2919402 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28180,
            "unit": "ns/op",
            "extra": "42549 times\n4 procs"
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
          "id": "503bf64f672d4e55a343c4484400ba68f28878a5",
          "message": "Bump chainlink-protos/billing/go to 48e1e6e1717c (#1310)\n\n* Bump chainlink-protos/billing/go to 48e1e6e1717c\n\n* Rework billing workflow client",
          "timestamp": "2025-07-01T16:36:07Z",
          "tree_id": "4c2445c2d7388b51f1bc19195bdf9bf08ff76d6a",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/503bf64f672d4e55a343c4484400ba68f28878a5"
        },
        "date": 1751387894490,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 353.2,
            "unit": "ns/op",
            "extra": "3399997 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 410.6,
            "unit": "ns/op",
            "extra": "2909397 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28159,
            "unit": "ns/op",
            "extra": "42715 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "168561091+engnke@users.noreply.github.com",
            "name": "engnke",
            "username": "engnke"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "36c47e10fbe9af88f2db1cd21cd141d8f31cd12e",
          "message": "Update chip-ingress client to accept and extract additional attributes (#1283)\n\n* Update chip-ingress client to accept and extract additional attributes\n\n* update test cases to include attributes\n\n* minor refactoring\n\n* update to accept recordedtime attribute\n\n* time attributes in utc\n\n---------\n\nCo-authored-by: Kiryll Kuzniecow <kiryll.kuzniecow@gmail.com>",
          "timestamp": "2025-07-01T14:28:40-04:00",
          "tree_id": "d84d5a52b9812d3e619ee15ea697e7c7ec010929",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/36c47e10fbe9af88f2db1cd21cd141d8f31cd12e"
        },
        "date": 1751394590912,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.3,
            "unit": "ns/op",
            "extra": "3363525 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412.1,
            "unit": "ns/op",
            "extra": "2934543 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28353,
            "unit": "ns/op",
            "extra": "42730 times\n4 procs"
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
          "id": "94ad8428563ac696dfe1e14dc20beb400d421850",
          "message": "Bump chainlink protos to 37bd0d618b58951f039b509dc49c92da68bd0808 for billing client protos (#1319)",
          "timestamp": "2025-07-01T19:48:00Z",
          "tree_id": "c9395e5d06c2f9af77e6c5d9dc051064c0cca400",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/94ad8428563ac696dfe1e14dc20beb400d421850"
        },
        "date": 1751399415762,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 351.7,
            "unit": "ns/op",
            "extra": "3427426 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 408.9,
            "unit": "ns/op",
            "extra": "2875963 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28132,
            "unit": "ns/op",
            "extra": "42723 times\n4 procs"
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
          "id": "31e037fcea7449b4761488db44b3edf40d3b5709",
          "message": "pkg/types: add UnimplementedRelayer (#1316)",
          "timestamp": "2025-07-01T16:02:01-05:00",
          "tree_id": "aca8501057b24c99a3350a2313d931cb0d782bb8",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/31e037fcea7449b4761488db44b3edf40d3b5709"
        },
        "date": 1751403799206,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 356.6,
            "unit": "ns/op",
            "extra": "3380480 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 412,
            "unit": "ns/op",
            "extra": "2926221 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28131,
            "unit": "ns/op",
            "extra": "42709 times\n4 procs"
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
          "id": "455da707ca1b946fe4b4f6e1c4d6f3355eff3df8",
          "message": "Add messages for CRUDL operations (#1314)\n\n* Add messages for CRUDL operations\n\n* Add messages for CRUDL operations",
          "timestamp": "2025-07-02T10:08:59Z",
          "tree_id": "cef58b7ea5b06a91c848eb03cbdf51593e482b5f",
          "url": "https://github.com/smartcontractkit/chainlink-common/commit/455da707ca1b946fe4b4f6e1c4d6f3355eff3df8"
        },
        "date": 1751451010621,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkKeystore_Sign/nop/in-process",
            "value": 354.5,
            "unit": "ns/op",
            "extra": "3405841 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/hex/in-process",
            "value": 407.8,
            "unit": "ns/op",
            "extra": "2936966 times\n4 procs"
          },
          {
            "name": "BenchmarkKeystore_Sign/ed25519/in-process",
            "value": 28096,
            "unit": "ns/op",
            "extra": "42484 times\n4 procs"
          }
        ]
      }
    ]
  }
}