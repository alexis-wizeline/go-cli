package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var youtubeCmd = &cobra.Command{
	Use:     "youtube [search]",
	Short:   "commands to open youtube or youtube music in the default browser",
	Aliases: []string{"y"},
	Run: func(cmd *cobra.Command, args []string) {
		url := "https://www.youtube.com/"
		openMusic, err := cmd.Flags().GetBool("music")
		if err != nil {
			fmt.Println("error with music:", err)
		}
		if openMusic {
			url = "https://music.youtube.com/"
		}

		if len(args) > 0 {
			url = parseSearch(url, args, openMusic)
		}
		c := exec.Command("open", url)

		err = c.Run()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func parseSearch(url string, args []string, music bool) string {
	sep := " "
	template := "%s/results?search_query=%s"
	if music {
		sep = "+"
		template = "%s/search?q=%s"
	}

	return fmt.Sprintf(template, url, strings.Join(args, sep))
}

func init() {
	rootCmd.AddCommand(youtubeCmd)

	youtubeCmd.Flags().BoolP("music", "m", false, "will open youtube music instead")
}
