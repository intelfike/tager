package main

import (
	"fmt"
	"os"

	"github.com/intelfike/nestmap"
	"github.com/spf13/cobra"
)

var showTagsCmd = &cobra.Command{
	Use:   "tag",
	Short: "タグを一覧する",
	Long:  "タグを一覧する",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if *showFlagR {
				fmt.Println("-r --recursive を指定した場合には表示するタグ名も入力してください")
				return
			}
			showTags(rootTags.Keys())
			return
		}

		// AND計算のため
		tags := make([]string, 0)

		for _, v := range args {
			cur := rootTags.Child(v)
			if !cur.Exists() {
				fmt.Println("そのようなタグは存在しません")
				continue
			}
			tags = append(tags, v)
		}

		// AND計算

		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println("そのようなタグは存在しません")
			return
		}
		if cur.Child("tags").IsMap() {
			if *showFlagR {
				recNestTag(cur, args[0], func(nm *nestmap.Nestmap, path string) {
					fmt.Println(path + "/" + nm.BottomPath().(string))
				})
			} else {
				showTags(cur.Child("tags").Keys())
			}
		}
	},
}

// ==================== add ====================

// 循環参照のチェックをだね
var addTagsCmd = &cobra.Command{
	Use:   "tag [flags] TAG TAGS...",
	Short: "タグにタグを登録する",
	Long:  "タグにタグを登録する\n登録先のタグ、登録するタグの両方が create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		cur := rootTags.Child(args[0])
		for _, v := range args[1:] {
			if !rootTags.HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			if args[0] == v {
				fmt.Println(v, "登録元と登録先のタグが同じです")
				continue
			}
			if cur.Child("tags").HasChild(v) {
				fmt.Println(v, "というタグは既に", args[0], "に登録されています")
				continue
			}
			cur.Child("tags", v).Set(v)
		}
		// 循環参照をチェックして拒否するため
		recNestTag(cur, "", func(nm *nestmap.Nestmap, path string) {
			s := nm.BottomPath().(string)
			unique := args[0] != s
			if !unique {
				fmt.Println(s, ":循環参照です\n登録に失敗しました")
				os.Exit(1)
			}
		})
	},
}

// ==================== remove ====================

var removeTagsCmd = &cobra.Command{
	Use:   "tag [flags] TAG TAGS...",
	Short: "タグからタグを削除する",
	Long:  "タグからタグを削除する\n削除するタグが存在していない場合は無視されます",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		for _, v := range args[1:] {
			if !rootTags.Child("tags").HasChild(v) {
				fmt.Println(v, "そのようなタグは存在しません")
				continue
			}
			cur.Child("tags", v).Remove()
		}
	},
}

// ==================== autoremove ====================
var autoremoveTagsCmd = &cobra.Command{
	Use:   "tag [TAG....]",
	Short: "タグから存在しないタグを自動削除する",
	Long:  "タグから存在しないタグを自動削除する\nタグ名が未指定の場合はすべてのタグが対象です",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 引数がなければすべてが対象
			args = rootTags.Keys()
		}
		for _, v := range args {
			cur := rootTags.Child(v)
			if !cur.Exists() {
				fmt.Println(v, "そのようなタグはありません")
				continue
			}
			if !cur.HasChild("tags") {
				continue
			}
			for _, tag := range cur.Child("tags").Keys() {
				if rootTags.HasChild(tag) {
					continue
				}
				cur.Child("tags", tag).Remove()
				fmt.Println(v, "から", tag, "というタグを削除しました")
			}
		}
	},
}
