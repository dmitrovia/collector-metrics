build:
	go build cmd/grpc/main.go

install-bin:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Скачивание proto google
google-proto:
	curl https://raw.githubusercontent.com/googleapis/googleapis/974ad5bdfc9ba768db16b3eda2850aadd8c10a2c/google/api/annotations.proto --create-dirs -o google/api/annotations.proto
	curl https://raw.githubusercontent.com/googleapis/googleapis/974ad5bdfc9ba768db16b3eda2850aadd8c10a2c/google/api/http.proto --create-dirs -o google/api/http.proto

# Скачивание proto validate
proto-validate:
	curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/validate.proto --create-dirs -o buf/validate/validate.proto
#curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/expression.proto --create-dirs -o buf/validate/expression.proto
#curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/priv/private.proto --create-dirs -o buf/validate/priv/private.proto

# Скачивание proto openapiv2
proto-openapiv2:
	curl https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto --create-dirs -o protoc-gen-openapiv2/options/annotations.proto
	curl https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto --create-dirs -o protoc-gen-openapiv2/options/openapiv2.proto