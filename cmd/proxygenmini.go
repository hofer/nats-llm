package cmd

import (
	"github.com/hofer/nats-llm/internal/proxy"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var apiKey string

var proxyGenminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Connecting to the Nats.io Server: %s", proxyNatsUrl)
		nc, err := nats.Connect(proxyNatsUrl)
		if err != nil {
			log.Fatal(err)
		}

		err = proxy.StartNatsGeminiProxy(nc, apiKey)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	proxyCmd.AddCommand(proxyGenminiCmd)
	proxyGenminiCmd.PersistentFlags().StringVarP(&apiKey, "apiKey", "k", os.Getenv("GEMINI_API_KEY"), "Gemini API key")
}
