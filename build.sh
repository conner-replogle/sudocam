protoc -I=./protos/ --go_out=./ ./protos/msgs.proto
#protoc --python_out=python/ --proto_path ./protos  ./protos/msgs.proto
cd server/ui
pnpm dlx pbjs -t static-module -w commonjs --ts src/types/binding.ts ../../protos/msgs.proto
