/*
# タグ型ファイル管理システム(ラッパー)
## 気をつけること
- ドキュメントを内包しているかのような使いやすさ(サブコマンドの利用)
- タイプミスなどを防ぐフールプルーフ設計(タグを定義すること)
	findコマンド連携などをさせない
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
	config          *nestmap.Nestmap
	configFile      string
	rootTags        *nestmap.Nestmap
	showFlagR       *bool
	mountFlagR      *bool
	addFileFlagR    *bool
	removeFileFlagR *bool
	tager           = new(Tager)
)

var RootCmd = &cobra.Command{
	Use:   "tager",
	Short: "Semantic File System",
	Long:  "[Semantic File System]",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !tager.isInited() {
			fmt.Println("初期設定がされていません\ntager init を実行してください")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初期設定",
	Long: `初期設定
最初に実行してください

~/.tager/config.json が生成されます
apt install sshfs が実行されます
`,
	Run: func(cmd *cobra.Command, args []string) {
		if !tager.isInited() {
			tager.init()
		} else {
			fmt.Println("初期設定済みです")
		}
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

var infoCmd = &cobra.Command{
	Use:   "info [TAG]",
	Short: "現在のツール情報を表示する",
	Long:  "現在のツール情報を表示する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			// リンク切れの詳細表示
			tags, err := tager.autoremovableTags(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, tag := range tags {
				fmt.Println(tag, "というタグのリンクが切れています")
			}
			files, _ := tager.autoremovableFiles(args[0])
			for _, file := range files {
				fmt.Println(file, "というファイルのリンクが切れています")
			}

			fmt.Println()
			fmt.Println("tager autoremove [TAGS...]")
			return
		}
		// 「現在」の情報のため、カレントタグの情報表示
		fmt.Println("current tag:", config.Child("root", "current"))
		fmt.Println()
		// autoremoveでのリンク切れ削除のチェック用
		for _, v := range tager.rootTags.Keys() {
			tags, err := tager.autoremovableTags(v)
			if len(tags) != 0 && err == nil {
				fmt.Println(v, "タグに", len(tags), "個のタグのリンク切れが見つかりました")
			}
			files, _ := tager.autoremovableFiles(v)
			if len(files) != 0 {
				fmt.Println(v, "タグに", len(files), "個のファイルのリンク切れが見つかりました")
			}
		}
		fmt.Println()
		fmt.Println("tager info TAG で詳細を確認することができます")
	},
}

var chCmd = &cobra.Command{
	Use:   "ch TAG",
	Short: "カレントタグを変更する",
	Long:  "カレントタグを変更する",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}
		if err := execValis(cmd, args, tagExists); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		config.Child("root", "current").Set(args[0])
	},
	PersistentPostRun: savePost,
}

var mountCmd = &cobra.Command{
	Use:   "mount [flags] TAG",
	Short: "シンボリックリンク集を作成する",
	Long:  "シンボリックリンク集を作成する\nカレントディレクトリに、指定されたタグ名と同じディレクトリ名で作成されます",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		if len(args) != 1 {
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

var createCmd = &cobra.Command{
	Use:   "create [flags] TAG",
	Short: "新しいタグを作成する",
	Long:  "新しいタグを作成する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		root := rootTags
		for _, v := range args {
			if root.HasChild(v) {
				fmt.Println(v, "というタグは既に存在しています")
				continue
			}
			if strings.ContainsAny(v, "/") {
				fmt.Println("/", "タグ名にこれらの文字は利用できません")
			}
			if v == "." {
				fmt.Println(".", "タグ名は名前は予約されています")
			}
			// タグの初期化
			root.Child(v).MakeMap()
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [flags] TAG",
	Short: "タグを完全に削除する",
	Long:  "タグを完全に削除する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		root := rootTags
		for _, v := range args {
			if !root.HasChild(v) {
				fmt.Println(v, "というタグは存在しません")
				continue
			}
			// タグの初期化
			root.Child(v).Remove()
		}
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "データを一覧する",
	Long:  "データを一覧する",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		cmd.SetArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add COMMAND TAG DATA...",
	Short: "タグにデータを登録する",
	Long:  "タグにデータを登録する",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.Help()
			os.Exit(0)
		}
		if err := execValis(cmd, args, tagExists); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cmd.SetArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPostRun: savePost,
}

var removeCmd = &cobra.Command{
	Use:   "remove COMMAND TAG DATA...",
	Short: "タグからデータの登録を解除する",
	Long:  "タグからデータの登録を解除する\n削除するデータが存在していない場合は無視されます",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.Help()
			os.Exit(0)
		}
		if err := execValis(cmd, args, tagExists); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cmd.SetArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPostRun: savePost,
}

var autoremoveCmd = &cobra.Command{
	Use:   "autoremove COMMAND [TAG...]",
	Short: "タグから存在しないデータを自動削除する",
	Long:  "タグから存在しないデータを自動削除する",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPostRun: savePost,
}

// ==================== validate func ====================

// execute validations
func execValis(cmd *cobra.Command, args []string, funcs ...func(cmd *cobra.Command, args []string) error) error {
	for _, f := range funcs {
		if err := f(cmd, args); err != nil {
			return err
		}
	}
	return nil
}

func tagExists(cmd *cobra.Command, args []string) error {
	cur := rootTags.Child(args[0])
	if !cur.Exists() {
		return errors.New(args[0] + " そのようなタグは存在しません")
	}
	return nil
}

// ==================== func ====================
func init() {
	config = nestmap.New()
	config.Indent = "\t"
	rootTags = config.Child("root", "tags")
	// 設定ファイルの読み込み
	dir := os.Getenv("HOME") + "/.tager"
	configFile = dir + "/config.json"

	cobra.OnInitialize()
	RootCmd.AddCommand(initCmd, versionCmd, infoCmd, mountCmd, chCmd)
	RootCmd.AddCommand(showCmd, createCmd, deleteCmd, addCmd, removeCmd, autoremoveCmd)
	showCmd.AddCommand(showTagsCmd, showFilesCmd, showAllCmd, showCommentCmd)
	addCmd.AddCommand(addTagsCmd, addFilesCmd, addCommentCmd)
	removeCmd.AddCommand(removeTagsCmd, removeFilesCmd)
	autoremoveCmd.AddCommand(autoremoveAllCmd, autoremoveTagsCmd, autoremoveFilesCmd)

	showFlagR = showCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にタグを辿ってデータを表示する")
	mountFlagR = mountCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にファイルをマウントする")
	addFileFlagR = addFilesCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にファイルを探索してタグに登録する")
	removeFileFlagR = removeFilesCmd.PersistentFlags().BoolP("recursive", "r", false, "再帰的にファイルを探索してタグから登録を解除する")

	// fileCmd.AddCommand(filelsCmd)
	// taglsCmd.Use = "tags"
	// filelsCmd.Use = "files"
	// RootCmd.AddCommand(taglsCmd, filelsCmd)
}

func main() {
	confb, err := ioutil.ReadFile(configFile)

	m := new(interface{})

	err = json.Unmarshal(confb, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config.Set(*m)

	tager.readConfig(configFile)

	// コマンド実行
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !f.IsDir()
}

func parseTagName(s string) string {
	if s == "." {
		current := config.Child("root", "current")
		if !current.Exists() {
			fmt.Println(". を利用しましたが、カレントタグが未登録です\ntager ch -h を参照してください")
			os.Exit(1)
		}
		s = current.ToString()
	}
	return s
}
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
func uniqueStrings(ss ...string) []string {
	result := make([]string, 0)
	unique := map[string]bool{}
	for _, v := range ss {
		_, ok := unique[v]
		if ok {
			continue
		}
		unique[v] = true
		result = append(result, v)
	}
	return result
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

func savePost(cmd *cobra.Command, args []string) {
	if err := tager.saveConfig(); err != nil {
		fmt.Println(err)
		return
	}
}
func save() error {
	b, err := config.BytesIndent()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, b, 0766)
}
