package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/intelfike/nestmap"
	"github.com/spf13/cobra"
)

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

// ==================== other ====================

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン番号を表示する",
	Long:  "バージョン番号を表示する",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tager v1.0")
	},
}

// ==================== tag ====================

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグ関連のコマンド",
	Long:  "タグを管理するためのサブコマンド",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "タグを一覧する",
	Long:  "タグを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		cur := moveTag(args).Child("tags")
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.IsMap() {
			fmt.Println(strings.Join(cur.Keys(), "\n"))
		}
		// else if cur.IsArray() {
		// 	for _, v := range cur.ToArray() {
		// 		fmt.Println(v.ToString())
		// 	}
		// }
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "新しいタグを作成する",
	Long:  "新しいタグを作成する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
		root := config.Child("root", "tags")
		for _, v := range args {
			if root.HasChild(v) {
				fmt.Println(v, "は既に存在しています。")
			}
			// タグの初期化
			tagini := getInitedTag()
			root.Child(v).Set(tagini.Interface())
			if err := save(); err != nil {
				fmt.Println("何故かセーブできませんでした")
				return
			}
		}
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

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "タグにタグを登録する",
	Long:  "タグにタグを登録する。\n登録先のタグ、登録するタグの両方が create されている必要があります。",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
                        cmd.SetUsageTemplate("tager tag add tags [tag] [tags]...")
                        cmd.Help()
                        return
                }
		cur := moveTag(args[0:1]).Child("tags")
		if !cur.Exists() {
                        fmt.Println(args[0], "そのようなタグは存在しません")
                        return
                }
		for _, v := range args[1:] {
			if !config.Child("root", "tags").HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			initag := getInitedTag()
			cur.Child(v).Set(initag.Interface())
		}
	},
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "タグにファイルを登録する",
	Long:  "タグにファイルを登録する。\n登録先のタグが create されている必要があります。",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.SetUsageTemplate("tager tag add files [tag] [file]...")
			cmd.Help()
			return
		}
		cur := moveTag(args[0:1]).Child("files")
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません。")
			return
		}
		for _, v := range args[1:] {
			_, err := os.Stat(v)
			if err != nil {
				fmt.Println(v, "そのようなファイルは存在しません。")
				continue
			}
			full, err := filepath.Abs(v)
			if err != nil {
				fmt.Println(v, "ファイル名の指定が正しくありません。")
				continue
			}
			cur.Child(full).Set(v)
			err = save()
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	},
}

// ==================== file ====================
var fileCmd = &cobra.Command{
        Use:   "file",
        Short: "ファイル関連のコマンド",
        Long:  "ファイルを管理するためのサブコマンド",
        Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var flsCmd = &cobra.Command{
        Use:   "ls",
        Short: "ファイルを一覧する",
        Long:  "ファイルを一覧する",
        Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.SetUsageTemplate("tager file ls [tag]...")
			cmd.Help()
			return
		}
		cur := moveTag(args).Child("files")
		if !cur.Exists() {
                        fmt.Println("そのようなタグは存在しません")
                        return
                }
                if cur.IsMap() {
                        fmt.Println(strings.Join(cur.Keys(), "\n"))
                }
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

func initFile(){
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
	tagCmd.AddCommand(lsCmd, createCmd, addCmd)
	addCmd.AddCommand(tagsCmd, filesCmd)
	fileCmd.AddCommand(flsCmd)
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
