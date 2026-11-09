# gpdf JSON — 文書定義の仕様（正典）

> gpdf の `template.FromJSON` に渡す JSON を書くために必要なものを1ファイルにまとめたもの。本文・JSON Schema・検証済みの実例が入っている。AI に JSON を書かせるときは、このファイルをそのまま渡せばよい。

- 公開場所: `https://gpdf.dev/spec/gpdf-json-spec.ja.md`（英語版 `gpdf-json-spec.en.md`）
- JSON Schema: `https://gpdf.dev/schema/gpdf.schema.json`（付録A と同じもの）
- 生成物。原本は gpdf リポジトリの `schema/` にある

🔴 **ここに載っている JSON はすべて実装に通したもの。** 付録A の Schema は gpdf が読み込む Go の型とテストで突き合わせてあり、付録B の実例はすべて Schema の検証と PDF の生成をテストで通している。

---

## 1. 書いて、確かめて、直す

文書は **JSON ファイル1つ**。手順は4つ。

1. **書く。** 付録B のいちばん近い実例から始める。
2. **Schema で検証する。** 付録A（または上の URL）に対して JSON Schema バリデータを掛ける。知らない欄・不正な色・不正な寸法・範囲外の `span` はここで落ちる。
3. **生成する。** 「7. 生成のしかた」のプログラムで PDF にする。
4. **確かめる。** 「8. 出来上がりの確かめ方」の手順で、文字が化けていないか・要素が消えていないかを見る。

- 🔴 **gpdf 自身は多くの間違いを黙って飲み込む。** 不正な値を持つ要素は捨てられ、エラーにならずに PDF ができる（「6. 黙って飲み込まれるもの」）。だから **手順2 と手順4 を省かないこと。** 生成が成功したことは、意図どおりにできたことを意味しない。
- **欄の名前・型・列挙値・書式の正は付録A。** この本文には写していない。

## 2. 文書の形

最上位は `page`（必須）と、中身を置く欄からなる。

| 欄 | 役割 |
|---|---|
| `page` | 用紙サイズ・向き・余白・全ページ共通の背景 |
| `metadata` | PDF に埋め込むタイトル・作成者など |
| `header` / `footer` | 全ページの上端・下端に繰り返す行 |
| `body` | 本文の行。収まらなければ**自動で改ページ**する |
| `pages` | 明示的に分けたページの配列。各ページが自分の `body` を持つ |
| `absolute` | 座標を指定して置く要素（ハンコ・透かし・欄外の注記など） |

- `body` と `pages` は両方書ける。`body` が1ページ目として先に置かれ、`pages` が後に続く。
- **改ページを自分で決めたいときだけ `pages` を使う。** 長い表や文章は `body` に置けば勝手に次ページへ流れる。

### 行と列（12カラムのグリッド）

本文は **行（`row`）の並び** で、各行は **列（`cols`）** を左から並べる。列の幅は `span` で、**1行の `span` の合計を 12 にする。**

- 🔴 **合計が 12 かどうかは誰も検査しない**（Schema は各 `span` が 1〜12 であることしか見ない）。12 を超えると右端からはみ出し、足りないと右側が空く。
- 空きを作りたいときは、中身の無い列（`{"span": 4, "text": ""}`）を置く。
- 行の高さは既定で中身に合わせて伸びる。`height` を寸法で書くと固定になる。

### 列の中身：省略形と `elements`

列に要素を1つだけ置くなら、列に直接書く（**省略形**）。

```json
{ "span": 12, "text": "見出し", "style": { "size": 18, "bold": true } }
```

複数を縦に積むなら `elements` 配列に `type` 付きで並べる。

```json
{ "span": 12, "elements": [
  { "type": "text", "content": "見出し", "style": { "size": 18, "bold": true } },
  { "type": "spacer", "height": "4mm" },
  { "type": "text", "content": "本文" }
]}
```

- 🔴 **`elements` があると、同じ列の省略形の欄（`text` / `table` など）はすべて無視される。** 混ぜないこと。
- 省略形の `text` に対応する文字列の欄は、`elements` の中では `content` になる。

## 3. 要素

`type` は次の13種。各要素の欄は付録A を見ること。

| `type` | 何を置くか | 省略形 |
|---|---|---|
| `text` | 文章。`style` で大きさ・太字・色・揃え・背景・余白など | `text` + `style` |
| `link` | リンク付きの文章。URL は `style.link` に書く（青・下線が既定） | なし |
| `richText` | 1段落の中で部分ごとに書式を変える（太字の一部・文中リンクなど） | `richText` |
| `image` | PNG / JPEG / SVG。`src` は data URI・ファイルパス・base64・SVG 文字列 | `image` |
| `table` | 見出し行とデータ行の表。長ければページをまたぐ | `table` |
| `list` | 箇条書き（`"ordered"` で番号付き） | `list` |
| `line` | 横線 | `line` |
| `spacer` | 縦の空き（`height` に寸法） | `spacer`（寸法の文字列） |
| `qrcode` | QR コード | `qrcode` |
| `barcode` | バーコード（Code 128） | `barcode` |
| `box` | 背景・枠・角丸・余白を持つ入れ物。中に `elements` を積む。入れ子にできる | `box` |
| `pageNumber` | 現在のページ番号（`header` / `footer` で使う） | なし |
| `totalPages` | 総ページ数 | なし |

- **`header` / `footer` も本文と同じ行と列で書く。** ページ番号は `footer` の列の `elements` に `pageNumber` を置く。
- `pageNumber` と `totalPages` は数字だけを出す。「1 / 3」のような組み立ては、列を分けて隣に並べる。

## 4. 寸法と色

- **寸法は単位付きの文字列。** `mm` / `cm` / `in` / `pt` / `em` / `%`。単位の無い数字は `pt`。正確な書式は付録A の `dimension`。
- **文字の大きさ（`size`）と字間（`letterSpacing`）だけは数値で、単位は pt。**
- **色は文字列。** `#RRGGBB` / `rgb(r, g, b)`（各 0.0〜1.0）/ `gray(v)`（0.0〜1.0）/ 8色の名前。`rgb()` は **0〜255 ではない**ことに注意。
- **`opacity` の 0 は「透明」ではなく「指定なし＝不透明」。** 透かしを薄くしたいなら 0.1〜0.3 のように 0 より大きい値を書く。
- 4辺を個別に書く欄（`paddings` / `margins` / `widths`）は CSS と同じ **上・右・下・左** の順。

## 5. フォント

- **フォントを登録しなければ、PDF 標準のフォント（Helvetica 系）が使われる。** 書けるのは実質 **ASCII だけ**。
- 🔴 **日本語・中国語・韓国語は、フォントを登録しないと `?` になる。** エラーにはならない。`—`（全角ダッシュ）のような ASCII 外の記号も、標準フォントでは消えたり別の字に化けたりする。
- **フォントの登録は JSON ではなく生成側（Go）の仕事。** JSON の `fontFamily` は「登録済みの名前を選ぶ」ことしかできず、登録されていない名前を書くと**黙って既定のフォントに戻る**。
- 🔴 **表のセルには `fontFamily` を指定する欄が無く、既定のフォントで描かれる。** だから CJK を含む文書は、**フォントを既定のフォントとして登録する**（「7. 生成のしかた」）。そうすれば JSON に `fontFamily` を1つも書かなくてよい。
- `bold` / `italic` は、その書体のファイルが登録されていなければ通常の書体で描かれる。

## 6. 黙って飲み込まれるもの

gpdf は次のものを **エラーにせず、捨てるか既定値に戻して PDF を作る。** 手順2（Schema）で落ちるものと、手順4（確かめ）でしか分からないものがある。

| 書いたもの | gpdf の振る舞い | どこで気づけるか |
|---|---|---|
| 知らない欄（綴り間違い・存在しない機能） | 無視 | Schema |
| 要素の不正な色（`"#ggg"`、`"rgb(255,0,0)"` など） | その色指定だけ無視 | Schema |
| 不正な寸法の `row.height` | **その行ごと消える** | Schema |
| 不正な寸法の `spacer` | 空きが消える | Schema |
| 知らない `type` | その要素が消える | Schema |
| `elements` と省略形の併記 | 省略形が消える | 生成後に目で見る |
| 読めない画像（パスの誤り・壊れた base64） | **その画像が消える** | 生成後に目で見る |
| `span` の合計が 12 でない | はみ出す・空く | 生成後に目で見る |
| 登録していない `fontFamily` | 既定のフォントに戻る | 生成後に目で見る |
| 標準フォントで ASCII 以外の文字 | `?` になる・消える | `pdftotext` |
| 表の `headerStyle` の `bold` / `size` / `fontFamily` | **無視**（効くのは `color` と `background` だけ） | 生成後に目で見る |
| `type: "link"` で `style.link` が無い | ただの文章になる | 生成後に目で見る |

**エラーになるのは** JSON の構文エラー、未知の用紙サイズ、`page.margins` の不正な寸法、全ページ背景（`page.background`）の不正な色・画像くらい。

## 7. 生成のしかた

`template.FromJSON` に JSON を渡して `Generate` する。次のプログラムはそのまま使える。第3引数にフォントファイルを渡すと、それを既定のフォントとして登録する。

```go
// render: go run . <in.json> <out.pdf> [font.ttf]
package main

import (
	"log"
	"os"

	"github.com/gpdf-dev/gpdf/template"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: render <in.json> <out.pdf> [font.ttf]")
	}
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var opts []template.Option
	if len(os.Args) > 3 { // CJK etc.: register the font and make it the default
		font, err := os.ReadFile(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts,
			template.WithFont("Main", font),
			template.WithDefaultFont("Main", 10))
	}
	doc, err := template.FromJSON(src, nil, opts...)
	if err != nil {
		log.Fatal(err)
	}
	pdf, err := doc.Generate()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(os.Args[2], pdf, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go mod init render && go get github.com/gpdf-dev/gpdf
go run . invoice.json invoice.pdf                       # 英数字だけの文書
go run . quote.json quote.pdf NotoSansJP-Regular.ttf    # 日本語を含む文書
```

- **画像のファイルパスは、プログラムを実行した場所からの相対パス。** 迷うなら data URI で JSON に埋め込む。
- `FromJSON` の第2引数にデータを渡すと、JSON は **Go テンプレート** として先に展開される（`{{.Name}}` など）。AI が書いた JSON を直接生成するときは `nil` を渡す。

## 8. 出来上がりの確かめ方

**生成が成功しても、意図どおりとは限らない**（「6. 黙って飲み込まれるもの」）。PDF ができたら次を見る。

1. **文字が出ているか。** `pdftotext out.pdf -` で文字を抜き出し、**`?` が並んでいないか**、書いた文章が全部あるかを見る。フォントの登録漏れはここで分かる。
2. **ページ数が想定どおりか。** `pdfinfo out.pdf` の `Pages`。1ページのつもりが2ページなら、何かがはみ出して改ページしている。
3. **見た目。** PDF を読めるなら PDF を、読めないなら `pdftoppm -png -r 60 out.pdf page` で画像にして見る。消えた画像・はみ出した列・効いていない見出しの書式はここでしか分からない。

AI に書かせる場合、**AI 自身がこの3つを実行して直せる。** 生成と確認を AI の手に渡しておけば、人が見るのは最後の1回で済む。
