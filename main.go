/*
# タグ型ファイル管理システム(ラッパー)
## 気をつけること
- ドキュメントを内包しているかのような使いやすさ(サブコマンドの利用)
- タイプミスなどを防ぐフールプルーフ設計(タグを定義すること)
- 環境設定などが必要ない設計(シングルバイナリでの提供、設定ファイルはHOMEディレクトリに設置など)

## 対象
複数人かつ大規模なプロジェクトに参加している人
ディレクトリ管理を柔軟にしたい人


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
- タグからシンボリックリンク集を自動生成(mount機能)
- tag1 AND tag2 / tag1 OR tag2 のような計算機能(-cオプションで実装)
- タグからタグへコピーする機能
- 複数のタグを統合する機能
-
- autoremoveの削除メッセージ

## 作成予定のサブコマンド
tager
	version
	mount [-r] [tag]
	copy [tag] [tag]
	show
		-c コメントの表示
		ANDの計算
---
*/
package main

import (
	"encoding/json"
	"errors"
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
	rootTags   *nestmap.Nestmap
	showFlagR  *bool
	mountFlagR *bool
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

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "シンボリックリンク集を作成する",
	Long:  "シンボリックリンク集を作成する\nカレントディレクトリに、指定されたタグ名と同じディレクトリ名で作成されます",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		if len(args) != 1 {
			addUsage(cmd, " TAG")
			cmd.Help()
			return
		}
		dir := "tager-" + args[0]

		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}

		if err := os.Mkdir(dir, 0777); err != nil {
			fmt.Println("tag:", args[0])
			fmt.Println(err)
			return
		}
		if cur.HasChild("files") {
			for _, v := range cur.Child("files").Keys() {
				newname := strings.Replace(v, "/", "-", -1)
				if err := os.Symlink(v, dir+"/"+newname); err != nil {
					fmt.Println(err)
					continue
				}
			}
		}
		if *mountFlagR {
			recNestTag(cur, dir, func(nm *nestmap.Nestmap, path string) {
				path = path + "/" + nm.BottomPath().(string)
				if err := os.Mkdir(path, 0777); err != nil {
					fmt.Println("tag:", path)
					fmt.Println(err)
					return
				}
				if !nm.HasChild("files") {
					return
				}
				for _, v := range nm.Child("files").Keys() {
					newname := strings.Replace(v, "/", "-", -1)
					if err := os.Symlink(v, path+"/"+newname); err != nil {
						fmt.Println("tag:", v)
						fmt.Println(err)
						continue
					}
				}
			})
		}
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "データを一覧する",
	Long:  "データを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "タグにデータを登録する",
	Long:  "タグにデータを登録する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "タグからデータの登録を解除する",
	Long:  "タグからデータの登録を解除する\n削除するデータが存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var autoremoveCmd = &cobra.Command{
	Use:   "autoremove",
	Short: "タグから存在しないデータを自動削除する",
	Long:  "タグから存在しないデータを自動削除する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// ==================== func ====================
func andStrings(a, b []string) []string {
	c := make([]string, 0)
	for _, v := range a {
		index := -1
		for n2, v2 := range b {
			if v == v2 {
				index = n2
			}
		}
		if index == -1 {
			continue
		}
		c = append(c, b[index])
	}
	return c
}

func showTags(tagNames []string) {
	for _, v := range tagNames {
		if !rootTags.Child(v).HasChild("comment") {
			fmt.Println(v)
			continue
		}
		comment := rootTags.Child(v, "comment").String()
		fmt.Println(v, ":", comment)
	}
}

func addUsage(cmd *cobra.Command, s string) {
	ut := cmd.UsageTemplate()
	ut = strings.Replace(ut, "{{.UseLine}}", "{{.UseLine}}"+s, -1)
	cmd.SetUsageTemplate(ut)
}

func nestTag(tags ...string) (*nestmap.Nestmap, error) {
	tags = strings.Split(strings.Join(tags, "/"), "/")

	cur := rootTags
	for _, v := range tags {
		if v == "" {
			continue
		}
		if !cur.HasChild(v) {
			return nil, errors.New("tag not exists")
		}
		cur = rootTags.Child(v, "tags")
	}
	return cur.Parent(), nil
}
func recNestTag(nm *nestmap.Nestmap, path string, cb func(*nestmap.Nestmap, string)) {
	if !nm.HasChild("tags") {
		return
	}
	for _, v := range nm.Child("tags").Keys() {
		cb(rootTags.Child(v), path)
		recNestTag(rootTags.Child(v), path+"/"+v, cb)
	}
}

func initFile() {
	os.Create(configFile)
	rootTags.MakeMap()
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
	rootTags = config.Child("root", "tags")

	cobra.OnInitialize()
	RootCmd.AddCommand(versionCmd, mountCmd)
	RootCmd.AddCommand(showCmd, createCmd, deleteCmd, addCmd, removeCmd, autoremoveCmd)
	showCmd.AddCommand(showTagsCmd, showFilesCmd, showAllCmd, showCommentCmd)
	addCmd.AddCommand(addTagsCmd, addFilesCmd, addCommentCmd)
	removeCmd.AddCommand(removeTagsCmd, removeFilesCmd)
	autoremoveCmd.AddCommand(autoremoveAllCmd, autoremoveTagsCmd, autoremoveFilesCmd)

	showFlagR = showCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にデータを表示する")
	mountFlagR = mountCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にファイルをマウントする")

	// fileCmd.AddCommand(filelsCmd)
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
