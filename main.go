/*
# タグ型ファイル管理システム(ラッパー)
## 気をつけること
- ドキュメントを内包しているかのような使いやすさ(サブコマンドの利用)
- タイプミスなどを防ぐフールプルーフ設計(タグを定義すること)
- 環境設定などが必要ない設計(シングルバイナリでの提供、設定ファイルはHOMEディレクトリに設置など)


## 目的1
以下のように、バッククォートで挟んで列挙することで、まとめてファイルを操作することを目的とする。
これは個人的に便利なだけ。

chmod 777 `tager files golang`
rm -f `tager file ls golang`

## 目的2
GUIからコマンドを呼び出すことも視野に入れている。
electronやWebアプリなど

## 目的3
サブコマンドの方式の開発に慣れる(Dockerやgitなどに倣う)


## その他詳細
- タグ名もファイルパスも一意なため、key-value型のデータ管理を利用する。(jsonの利用)
高速さ、処理の簡単さが魅力的。

## 作成予定のサブコマンド
tager
	version
	tag
		ls
		add
			tags
			files
		remove
			tags
			files

	tags [tags]...
	file
		ls
		add tags
		remove tags
	files [tags]...
---
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/intelfike/nestmap"
	"github.com/spf13/cobra"
)

// ==================== 定義 ====================
var (
	config     *nestmap.Nestmap
	configFile string
)

var RootCmd = &cobra.Command{
	Use:   "tager",
	Short: "Semantic File System",
	Long:  "[Semantic File System]",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン番号を表示する",
	Long:  "バージョン番号を表示する",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tager v1.0")
	},
}

// ==================== func ====================
func moveTag(tags []string) *nestmap.Nestmap {
	tags = strings.Split(strings.Join(tags, "/"), "/")

	root := config.Child("root", "tags")
	cur := root
	for _, v := range tags {
		if v == "" {
			continue
		}
		cur = cur.Child(v, "tags")
	}
	return cur.Parent()
}

func getInitedTag() *nestmap.Nestmap {
	tagini := nestmap.New()
	tagini.Child("tags").MakeMap()
	tagini.Child("files").MakeMap()
	return tagini
}

func initFile() {
	os.Create(configFile)
	config.Child("root", "tags").MakeMap()
	save()
}

func save() error {
	b, err := config.BytesIndent()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, b, 0766)
}

func init() {
	config = nestmap.New()
	config.Indent = "\t"

	cobra.OnInitialize()
	RootCmd.AddCommand(versionCmd, tagCmd, fileCmd)
	tagCmd.AddCommand(taglsCmd, createCmd, deleteCmd, addCmd, removeCmd, autoremoveCmd)
	addCmd.AddCommand(addTagsCmd, addFilesCmd)
	removeCmd.AddCommand(removeTagsCmd, removeFilesCmd)
	autoremoveCmd.AddCommand(autoremoveAllCmd, autoremoveTagsCmd, autoremoveFilesCmd)
	fileCmd.AddCommand(filelsCmd)
	// taglsCmd.Use = "tags"
	// filelsCmd.Use = "files"
	// RootCmd.AddCommand(taglsCmd, filelsCmd)
}

func main() {
	// 設定ファイルの読み込み
	dir := os.Getenv("HOME") + "/.tager"
	configFile = dir + "/tag.json"
	if _, err := os.Stat(dir); err != nil {
		os.Mkdir(dir, 0777)
		initFile()
	}

	confb, err := ioutil.ReadFile(configFile)
	if err != nil {
		initFile()
		confb, _ = ioutil.ReadFile(configFile)
	}

	m := new(interface{})

	err = json.Unmarshal(confb, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config.Set(*m)

	// コマンド実行
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
