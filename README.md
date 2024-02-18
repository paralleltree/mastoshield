# mastoshield

条件に一致するリクエストをフィルタするリバースプロキシです。

## Environment Variables

|Name|Required|Default value|Description|
|:--|:--:|:--|:--|
|`UPSTREAM_ENDPOINT`|Yes||プロキシ先のエンドポイント|
|`DENY_RESPONSE_CODE`|No|404|アクセスを拒否する場合のステータスコード|
|`PORT`|No|3000|プロキシがListenするポート番号|
|`EXIT_TIMEOUT`|No|10|プロキシが終了する際に待機するタイムアウト秒数|

## Command-line Arguments

|Argument|Description|
|:--|:--|
|`--rule-file`|ルール定義ファイルを指定します|
|`--test-rule`|ルール定義を検証し終了します|

## Ruleset Definition

リクエストの検証ルールはYAMLファイルに記述します。

```yaml
rulesets:
  - action: deny
    rules:
      - source: note_body
        contains: blocked_text
```

各rulesetは記述された順に検証されます。
rulesに含まれる条件に全て一致した場合に、そのrulesetのactionを適用します。

|Matcher|Description|
|:--|:--|
|`note_body`|投稿に`contains`で指定された文字列が含まれるか判定します。|
|`mention_count`|投稿のメンション数が`more_than`で指定した数より多いか判定します。|
|`remote_ip`|リクエスト元のIPアドレスが`contains`で指定されたアドレスに含まれるか判定します。|
|`user_agent`|リクエストのUserAgentに`contains`で指定された文字列が含まれるか判定します。|

## Logging

ログはLTSV形式で標準出力へ出力されます。
Infoレベルで出力されるログの種類は以下の通りです。

|event|Description|
|:--|:--|
|`event:start`|サーバーが起動する際に発生します。設定された内容が追加で出力されます。|
|`event:requestHandled`|サーバーがリクエストを処理した際に発生します。リクエストの内容、処理結果が出力されます。|
|`event:shutdown`|サーバーが終了する際に発生します。|

```
time:2024-02-17T14:57:10.916731Z        level:Info      event:start     port:2900       upstream:http://localhost:3333
time:2024-02-17T14:58:07.166895Z        level:Info      event:requestHandled    xid:cn8civjrf0ev2cl84qq0        action:allow    method:GET      path:/api/v1/timelines/home     url:/api/v1/timelines/home      remote:192.168.0.16       useragent:Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36
time:2024-02-17T14:58:33.175855Z        level:Info      event:shutdown
```
