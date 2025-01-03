# go-echo-restapi-boilerplate

Go Echo の REST API のボイラープレート

## 技術構成

- go
- sqlboiler
- sql-migrate
- echo
- oapi-codegen
- air
- ozzo-validation
- godotenv
- go-txdb
- stretchr/testify
- go-randomdata
- factory-go/factory

## 機能

### 認証

|                        | HTTP メソッド | URI                  | 権限     |
| ---------------------- | ------------- | -------------------- | -------- |
| 会員登録バリデーション | POST          | /auth/validateSignUp | 権限なし |
| 会員登録               | POST          | /auth/signUp         | 権限なし |
| ログイン               | POST          | /auth/signIn        | 権限なし |

### TODO リスト

|          | HTTP メソッド | URI        | 権限     |
| -------- | ------------- | ---------- | -------- |
| 作成     | POST          | /todos     | 認証済み |
| 一覧取得 | GET           | /todos     | 認証済み |
| 詳細     | GET           | /todos/:id | 認証済み |
| 更新     | PUT           | /todos/:id | 認証済み |
| 削除     | DELETE        | /todos/:id | 認証済み |

## 環境構築

### 1. 環境変数のファイルの作成

- root ディレクトリ配下の.env.sample をコピーし、.env とする
- .env.sample は開発環境をサンプルとしているため、設定値の調整は不要

```
cp .env.sample .env
```

### 2. Docker の build・立ち上げ

- ビルド

```
docker-compose build
```

- 起動

```
docker-compose up -d
```

- コンテナに入る

```
docker-compose exec api_server sh
```

### 3. 開発環境 DB のマイグレーション

- 2 で起動・コンテナに入った上で、以下のコマンドを実行

マイグレーションファイルの作成

```
make create-migration FILENAME=<ファイル名>
```

マイグレーション実行

```
make migrate
```

### 4. Web サーバの起動

air によりホットリロード形式で自動起動

postman 等を使用し、API アクセスができていることを確認

## sqlboiler によるコードの自動生成

コンテナに入った上で、以下のコマンドを実行

```
make prepare-sqlboiler

make gen-models
```

## openapi からスキーマの型生成

openapi 配下にドメインに応じたパスでスキーマを作成する

```
例) todos

- api_server/app/openapi
  - todos
		- todo.yaml
```

スキーマを生成したら Go の API のスキーマと型ファイルを生成する

```
make gen-schema DOMAIN=todos
```

## 設計方針

- Controller - Service のレイヤードアーキテクチャ
  - ロジックは Service に寄せる
  - Controller はリクエスト・レスポンスのハンドリング
    - リクエスト・レスポンスそれぞれのデータの加工含む
  - 将来的に複数モデルの保存などを行うケースが出てきたら、Service 層から Transactions クラスに切り出したりすると良さそう

## テスト方針

- それぞれの層でテストを書く

  - DB 接続があるところはモックを使わずに行う(実環境に近い形でテストする方が不具合検知に役立つため)
    - サービスが大きくなってくると、モックを使用せず結合テストした方がリグレッションテストにも寄与するため

- 正常系/異常系ともに書く
  - 可能な限り C1 カバレッジで書きたいところ
  - 事故があるとまずい機能については、C2 カバレッジで書いても良さそう

## テスト実行

api_server コンテナに入った上で、以下のコマンドを実行

```
make test-local
```
