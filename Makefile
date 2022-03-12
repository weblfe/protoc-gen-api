GO111MODULE=on
github_proxy=https://ghproxy.com/

install:
	@go get

gen_examples: install
	@protoc --go_out=./_examples/ --api_out=./_examples/ --go_opt=paths=source_relative -I_examples ./_examples/*.proto

gen_pb:
	@protoc --go_out=./testdata/ --api_out=./testdata/ --go_opt=paths=source_relative -I testdata ./testdata/**/*.proto

test: test_examples
	@go test ./...

test_examples:
	@cd _examples && go test

run_examples:
	@cd _examples && go run main.go greeter.pb.go greeter.http.go

curl_google_option_proto:
	@curl ${github_proxy}/https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > _examples/google/api/annotations.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > testdata/google/api/annotations.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > _examples/google/api/http.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > testdata/google/api/http.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto> _examples/google/protobuf/any.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto> testdata/google/protobuf/any.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/api.proto> _examples/google/protobuf/api.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/api.proto> testdata/google/protobuf/api.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/duration.proto> _examples/google/protobuf/duration.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/duration.proto> testdata/google/protobuf/duration.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/empty.proto> _examples/google/protobuf/empty.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/empty.proto> testdata/google/protobuf/empty.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/field_mask.proto> _examples/google/protobuf/field_mask.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/field_mask.proto> testdata/google/protobuf/field_mask.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/source_context.proto> _examples/google/protobuf/source_context.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/source_context.proto> testdata/google/protobuf/source_context.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/struct.proto> _examples/google/protobuf/struct.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/struct.proto> testdata/google/protobuf/struct.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/timestamp.proto> _examples/google/protobuf/timestamp.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/timestamp.proto> testdata/google/protobuf/timestamp.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/type.proto> _examples/google/protobuf/type.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/type.proto> testdata/google/protobuf/type.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/wrappers.proto> _examples/google/protobuf/wrappers.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/wrappers.proto> testdata/google/protobuf/wrappers.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/descriptor.proto> _examples/google/protobuf/descriptor.proto
	@curl ${github_proxy}/https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/descriptor.proto> testdata/google/protobuf/descriptor.proto
