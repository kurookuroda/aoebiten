# Pyxel → Ebiten 移植計画書

## 1. 目的
Pyxel で実装された「青空文庫リーダー（Aozora Reader）」を Go 言語 + Ebiten へ移植し、以下を再現する。

- タイピング風の1文字ずつ表示（テキスト送り）
- フォントサイズ切り替え（10px / 12px）
- キーボード・ゲームパッド入力対応
- ページ送り・ページ戻り・スキップ・高速送り
- 効果音（ビープ音）
- 読了画面・ページ番号表示

---

## 2. 技術スタック・依存関係

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.22+ |
| ゲームエンジン | `github.com/hajimehoshi/ebiten/v2` |
| フォント解析 | `golang.org/x/image/font/opentype` |
| 画像/描画 | `github.com/hajimehoshi/ebiten/v2` 標準パッケージ |
| 音声 | `github.com/hajimehoshi/ebiten/v2/audio` |
| 入力補助 | `github.com/hajimehoshi/ebiten/v2/inpututil` |

---

## 3. 移植方針（Pyxel → Ebiten 概念対応）

| Pyxel | Ebiten | 備考 |
|-------|--------|------|
| `pyxel.init()` | `ebiten.RunGame()` + `Layout()` | 画面サイズは `Layout` で指定 |
| `pyxel.Font` | `font.Face` (`opentype.NewFace`) | TTF をパースして `font.Face` を生成 |
| `pyxel.text()` | `text.Draw()` | 座標・文字列・色・Face を指定 |
| `pyxel.btn()` | `ebiten.IsKeyPressed()` | 押下中判定 |
| `pyxel.btnp()` | `inpututil.IsKeyJustPressed()` | 1回だけ押判定 |
| `pyxel.frame_count` | 手動カウンター `frameCount` | `Update()` 内でインクリメント |
| `pyxel.Sound` / `pyxel.play()` | `audio.NewContext` + `audio.Player` | PCM（矩形波）を生成して再生 |
| `pyxel.FONT_WIDTH` | `font.MeasureString()` | プロポーショナルフォント対応で必須 |
| `pyxel.rect()` / `rectB()` | `ebiten.NewImage()` + `DrawImage()` / `vector` | 矩形描画は `image/draw` または `vector` パッケージ |

---

## 4. 実装ステップ

### Step 1: プロジェクト構成とメインエントリ

1. `go mod init aozora-reader`
2. `main.go` を作成し、`Game` 構造体に `ebiten.Game` インターフェースを実装
   - `Update() error`
   - `Draw(screen *ebiten.Image)`
   - `Layout(outsideWidth, outsideHeight int) (int, int)`
3. `ebiten.SetWindowSize(SCREEN_W*2, SCREEN_H*2)` などでウィンドウサイズを設定（論理解像度は 256x256 を維持）

### Step 2: 定数・設定の移植

Pyxel 版の定数をそのまま Go の `const` / `var` へ移行。

- `SCREEN_W`, `SCREEN_H`, `BOX_X`, `BOX_Y`, `BOX_W`, `BOX_H`
- `PADDING`, `FOOTER_H`, `MAX_TEXT_W`
- `CHAR_INTERVAL`, `FAST_INTERVAL`
- `FONT_CONFIG` → `map[int]string`
- `FILE_PATH`

### Step 3: フォント管理（TTF読み込み・Face生成）

1. TTF ファイルを `embed` または `os.ReadFile` で読み込み
2. `opentype.Parse()` でフォントデータを解析
3. `opentype.NewFace(font, &opentype.FaceOptions{Size: float64(size), DPI: 72})` で `font.Face` を生成
4. サイズごとに `map[int]font.Face` にキャッシュ
5. フォント切り替え時は `currentFace` をし替え、ページを再構築

### Step 4: テキスト処理（段落読み込み・折り返し・ページ分割）

1. **段落読み込み** (`loadParagraphs`)
   - `os.ReadFile` + `strings.Split(string(data), "\n")`
2. **折り返し** (`wrapParagraphs`)
   - `font.MeasureString(face, text).Ceil()` で幅を計測
   - 1文字ずつ（または1ルーンずつ）追加して `maxW` を超えたら改行
   - 日本語は単語境界を考慮せず文字単位でよい
3. **ページ分割** (`paginate`)
   - `rowsPerPage = (BOX_H - PADDING*2 - FOOTER_H) / lineHeight`
   - `[][]string` 型でページを保持

### Step 5: 入力抽象化（キーボード＋ゲームパッド）

1. **キーボード**
   - `ebiten.KeyZ`, `ebiten.KeyX`, `ebiten.KeyF`, `ebiten.KeyEnter` などを直接使用
   - `inpututil.IsKeyJustPressed` で `btnp` 相当
   - `ebiten.IsKeyPressed` で `btn` 相当
2. **ゲームパッド**
   - Ebiten の `ebiten.StandardGamepadButton` を使用
   - `ebiten.GamepadID` を列挙し、接続中のパッドを確認
   - 対応表:
     - A → `StandardGamepadButtonRightBottom`
     - B → `StandardGamepadButtonRightRight`
     - X → `StandardGamepadButtonRightLeft`
     - UP → `StandardGamepadButtonLeftTop`
     - DOWN → `StandardGamepadButtonLeftBottom`
     - SELECT → `StandardGamepadButtonCenterLeft`
     - START → `StandardGamepadButtonCenterRight`
3. **入力ヘルパー**
   - `_btn(key)`, `_btnp(key)` に相当するヘルパーを実装し、キーボードとゲームパッドの OR 条件を統合
   - キーリピート（長押し）が必要な場合は `inpututil` + 独自カウンターで実装（Pyxel の `btnp` は自動リピートを持つため）

### Step 6: 効果音（Pyxel風ビープ音の生成）

Pyxel の `snd.set(note, tone, volume, effect, speed)` に相当する機能は Ebiten にないため、PCM データを自前で生成する。

1. `audio.NewContext(44100)` でオーディオコンテキストを初期化
2. 矩形波（Square Wave）を生成する関数を用意
   - 周波数（c3 = 130.81Hz など）
   - デューティ比（Pyxel の "t" tone は矩形波と推定）
   - 再生時間（Pyxel の speed に相当）
3. 4種類の SFX を `[]byte`（WAV または raw PCM）として生成し、`audio.NewPlayerFromBytes` で `audio.Player` を作成
   - `sndTalk`: 通常タイピング音
   - `sndTalkSpace`: スペース高速時
   - `sndTalkFast`: DOWN+A 高速時
   - `sndTalkFaster`: DOWN+B 超高速時
4. 再生時は `Player.Play()` を呼び出し、再生終了を待たずに次のフレームへ

### Step 7: ゲーム構造体と Update ロジック

`App` クラス → `Game` 構造体へ移植。

1. **フィールド**
   - `fonts map[int]font.Face`
   - `paragraphs []string`
   - `pages [][]string`
   - `pageIndex`, `revealed`, `timer`, `frameCount`
   - `pageDone`, `skipCooldown`
   - `currentSize int`
   - `sndTalk`, `sndTalkSpace`, `sndTalkFast`, `sndTalkFaster *audio.Player`
2. **メソッド**
   - `rebuildPages()`: フォント変更時にページを再構築
   - `reset()`: 読み位置を先頭へ
   - `toggleFontSize()`: 10 ↔ 12 切り替え
   - `getTypingSpeed()`: 入力状態から interval と charsPerTick を返す
   - `playTalkSound()`: 速度に応じた効果音を再生
3. **Update フロー**
   - Q / Escape → `return ebiten.Termination`（または `os.Exit`）
   - R / SELECT / START → `reset()`
   - F / X → `toggleFontSize()`
   - UP → 前のページ（または読了位置へ戻る）
   - skip / next 判定（DOWN単体、ENTER、マウス左など）
   - `pageDone == false` → `revealed` を進める（`
` はスキップ）
   - `pageDone == true` → 次のページへ

### Step 8: Draw ロジック（描画）

1. **画面クリア**
   - `screen.Fill(color.RGBA{0, 0, 0, 255})`
2. **テキストボックス**
   - 背景矩形: `vector.DrawFilledRect()` または `ebiten.NewImage()` で事前生成
   - 枠線: `vector.StrokeRect()`
3. **テキスト描画**
   - `text.Draw(screen, line, face, x, y, color.White)`
   - ベースライン補正に注意（Ebiten の `text.Draw` は左上ではなくベースライン基準ではないが、Y座標は文字の左上ではなくベースラインに近い位置になるため、フォントサイズ分のオフセットを確認）
   - 各行の Y 座標: `BOX_Y + PADDING + i*lineHeight`
4. **「▼」カーソル**
   - `pageDone && (frameCount/30)%2 == 0` で点滅表示
5. **読了画面**
   - `pageIndex >= len(pages)` のとき `-- 読了 --` を表示
6. **UI フッター**
   - 左下: フォントサイズ（例: `12px`）
   - 右下: ページ番号（例: `3/15`）
   - 右寄せは `MeasureString` で幅を計算して X 座標を決定

### Step 9: ビルドと実行確認

1. `go build -o aozora-reader`
2. 以下を確認
   - テキストファイルが存在しない場合のフォールバックメッセージ
   - フォントサズ切り替え時のページ再構築
   - ゲームパッド接続・未接続の両方で動作
   - 音声が再生されること（Ebiten のオーディオは一部環境で遅延あり）

---

## 5. ファイル構成案

```
aozora-reader/
├── main.go              # エントリポイント、Game 構造体、Update/Draw/Layout
├── config.go            # 定数、設定値
├── text.go              # 段落読み込み、折り返し、ページ分割
├── input.go             # キーボード・ゲームパッド入力ヘルパー
├── audio.go             # PCM 生成、Player 作成、効果音再生
├── go.mod
├── go.sum
├── PixelMplus10-Regular.ttf
├── PixelMplus12-Regular.ttf
└── aozora_416.txt
```

※ 単一ファイル (`main.go` のみ) にまとめてもよいが、責務を分離すると移植作業が並行しやすい。

---

## 6. 注意点・既知の課題

| 項目 | 詳細 |
|------|------|
| **ベースライン** | Ebiten の `text.Draw` は Y 座標をベースラインとして解釈する場合があるため、Pyxel とピクセル完全一致しない可能性あり。描画後に微調整が必要。 |
| **キーリピート** | Pyxel の `btnp` は自動リピートを持つが、`inpututil.IsKeyJustPressed` は持たない。長押しリピートが必要な場合は独自タイマーを実装する。 |
| **プロポーショナルフォント** | PixelMplus は等幅ではない可能性があるため、`MeasureString` による幅計測は必須。Pyxel 版の `font.text_width` に相当。 |
| **ゲームパッド互換** | Ebiten の `StandardGamepadButton` は Xbox レイアウト準拠。他機種では物理ボタンと表示が異なる場合がある。 |
| **音声フォーマット** | Ebiten の `audio` は PCM（unsigned 8bit または signed 16bit, little endian）を期待する。WAV ヘッダを自前で付与するか、無音埋め込みで調整する必要がある。 |
| **マウス入力** | `ebiten.IsMouseButtonPressed` / `inpututil.IsMouseButtonJustPressed` を使用。Pyxel と同様に左クリックで進める。 |
| **可変フォントサイズ時のページング** | フォント切り替え時に `rebuildPages()` を呼び出し、`pageIndex` が範囲外にならないようにクリップする。 |

---

## 7. 優先順位（実装順）

1. **基盤**: `main.go` + `Game` 構造体 + 画面表示まで
2. **テキスト**: ファイル読み込み + 折り返し + ページ分割 + 静的表示
3. **入力**: キーボード入力でページ送り・スキップ
4. **タイピング効果**: 1文字ずつ表示 + タイマー制御
5. **効果音**: ビープ音生成 + 再生
6. **ゲームパッド**: パッド入力対応
7. **UI Polish**: フッター、読了画面、フォント切り替え、リセット
