package cmd

import (
	daemongo_tcp "Filehub/daemon_tcp"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "filehub",
		Short: "FileHub CLI for file transfer",
		Long:  `FileHub CLI allows sending and receiving files over TCP using peer-to-peer connections.`,
	}

	// Subcommand for receiving files
	var receiveCmd = &cobra.Command{
		Use:   "receive",
		Short: "Start the receiver thread to accept incoming files",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting receiver thread...")
			daemongo_tcp.ReceiverThread()
		},
	}

	// Subcommand for sending files
	var sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send a file to a specified peer",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			filePath, _ := cmd.Flags().GetString("file")
			peerName, _ := cmd.Flags().GetString("peer")
			fmt.Printf("Sending file: %s to peer: %s\n", filePath, peerName)
			daemongo_tcp.SenderThread(filePath, peerName)
		},
	}

	var disocverCmd = &cobra.Command{
		Use:   "send",
		Short: "Send a file to a specified peer",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			filePath, _ := cmd.Flags().GetString("file")
			peerName, _ := cmd.Flags().GetString("peer")
			fmt.Printf("Sending file: %s to peer: %s\n", filePath, peerName)
			daemongo_tcp.SenderThread(filePath, peerName)
		},
	}

	// Add flags to the send command
	sendCmd.Flags().StringP("file", "f", "", "Path to the file to send")
	sendCmd.Flags().StringP("peer", "p", "", "Name of the peer to send the file to")

	// Make file and peer flags required
	sendCmd.MarkFlagRequired("file")
	sendCmd.MarkFlagRequired("peer")

	// Register subcommands
	rootCmd.AddCommand(receiveCmd)
	rootCmd.AddCommand(sendCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
