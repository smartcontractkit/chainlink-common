package actionandtrigger

//go:generate protoc -I. -I../../../../pb --go_out=. --go_opt=paths=source_relative "--cre_out=mode=don,id=basic-test-action-trigger@1.0.0:." action_and_trigger.proto
