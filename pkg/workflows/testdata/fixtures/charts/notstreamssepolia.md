```mermaid
flowchart

	trigger[\"<b>trigger</b><br>trigger<br><i>(notstreams[at]1.0.0)</i>"/]
	
	data-feeds-report[["<b>data-feeds-report</b><br>consensus<br><i>(offchain_reporting[at]1.0.0)</i>"]]
	trigger --> data-feeds-report
		trigger --> data-feeds-report
		
	unnamed2[/"target<br><i>(write_ethereum-testnet-sepolia[at]1.0.0)</i>"\]
	trigger --> unnamed2
		
```