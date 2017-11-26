package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var showCommentCmd = &cobra.Command{
	Use:   "comment [flags] TAG",
	Short: "タグのコメントを表示する",
	Long:  "タグにコメントを表示する\n登録先のタグが create されている必要があります",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		cur := rootTags.Child(args[0])
		if !cur.Exists() {
			fmt.Println(args[0], "そのようなタグは存在しません")
			return
		}
		if !cur.HasChild("comment") {
			return
		}
		comment := cur.Child("comment").ToString()
		fmt.Println(comment)

		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

var addCommentCmd = &cobra.Command{
	Use:   "comment [flags] TAG COMMENT",
	Short: "タグにコメントを登録する",
	Long:  "タグにコメントを登録する\n登録先のタグが create されている必要があります",
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
		arg := strings.Join(args[1:], " ")
		cur.Child("comment").Set(arg)
		if err := save(); err != nil {
			fmt.Println(err)
			return
		}
	},
}
