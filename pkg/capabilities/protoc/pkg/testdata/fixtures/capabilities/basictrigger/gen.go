package basicaction

//go:generate protoc --go_out=. --go_opt=paths=source_relative "--cre_out=mode=don,id=basic-test-trigger@1.0.0,trigger=true:." basic_trigger.proto
