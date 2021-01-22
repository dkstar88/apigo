/*
Copyright © 2021 daochun.zhao <daochun.zhao@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	apigoGrpc "apigo/grpc"
	"apigo/grpc/httprunner"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"s"},
	Short:   "Starts a HTTP runner server",
	Long:    `Starts a HTTP runner server`,
	Run:     server,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().String("Host", "127.0.0.1", "gRPC Host address")
	serverCmd.PersistentFlags().UintP("Port", "p", 7000, "gRPC Host Port")
}

func server(cmd *cobra.Command, _ []string) {
	address, _ := cmd.PersistentFlags().GetString("Host")
	port, _ := cmd.PersistentFlags().GetUint("Port")
	startGrpcServer(address, port)
	println("Server started.")
}

func startGrpcServer(address string, port uint) {
	netListener := getNetListener(address, port)
	gRPCServer := grpc.NewServer()
	grpcRunnerImpl := apigoGrpc.NewHttpRunnerServer()
	//gRPCServer.RegisterService(&grpcRunnerImpl, &apigoGrpc.HttpRunnerServer{})
	httprunner.RegisterHttpRunnerServer(gRPCServer, grpcRunnerImpl)

	println(fmt.Sprintf("Starting gRPC server at %s:%d", address, port))
	// start the server
	if err := gRPCServer.Serve(netListener); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}

func getNetListener(address string, port uint) net.Listener {
	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", address, port),
	)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return lis
}
