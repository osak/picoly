goexperiment := "jsonv2"

# デフォルト: ビルド
default: build

# ビルド
build:
    GOEXPERIMENT={{goexperiment}} go build -o bin/picoly ./cmd/picoly

# テスト（全パッケージ）
test:
    GOEXPERIMENT={{goexperiment}} go test ./...

# テスト（詳細出力）
test-verbose:
    GOEXPERIMENT={{goexperiment}} go test -v ./...

# lint
lint:
    GOEXPERIMENT={{goexperiment}} go vet ./...

# E2Eインテグレーションテスト（ビルド後に実行）
e2e: build
    GOEXPERIMENT={{goexperiment}} go run ./cmd/integration-test

# クリーン
clean:
    rm -f bin/picoly

# 依存関係の更新
deps:
    GOEXPERIMENT={{goexperiment}} go mod tidy
