package basictrigger

//go:generate protoc -I. -I../../../../pb --go_out=. --go_opt=paths=source_relative "--cre_out=mode=don,id=basic-test-trigger@1.0.0:." basic_trigger.proto
