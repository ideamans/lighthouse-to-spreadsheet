# Lighthouse / Google Spreadsheet ユーティリティ

このツールはLighthouseの実行結果のJSONを解析し、パフォーマンススコアおよび重要な指標をGoogle Spreadsheetに追記するユーティティです。

Lighthouseを繰り返し実行する場面で、その推移を簡単に記録するために作成しました。

## 認証

GCPのプロジェクトで`Sheets API`を有効にし、サービスアカウントを作成します。サービスアカウントにロールの付与は不要です。

Google Spreadsheetでスプレッドシートを作成し、サービスアカウントのメールアドレスを`編集者`として共有します。

サービスアカウントのキー(JSON形式)をダウンロードし、`$HOME/.lighthouse-to-spreadsheet/service-account.json`として保存します。

## 設定

カレントディレクトリに`.env`ファイルを作成し、次の値を入力します。

- `LIGHTHOUSE_RESULT_PATH` Lighthouseの結果JSONファイルのパス。
- `SPREADSHEET_ID` スプレッドシートID。
- `SHEET_NAME` スプレッドシート上のシート名。

## 実行

次のコマンドを実行します。

```bash
lighthouse-to-spreadsheet
```

スプレッドシートにLighthouse結果ファイルのダイジェストが追記されます。

## コマンドラインからの設定

コマンド引数でも設定できます。

```bash
lighthouse-to-spreadsheet -lighthouse-result "Lighthouseの結果JSONファイルパス" -spreadsheet-id "スプレッドシートID" -sheet-name "シート名"
```
