/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iptecharch/cache/client"
	"github.com/iptecharch/cache/proto/cachepb"
	sdcpb "github.com/iptecharch/sdc-protos/sdcpb"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var readPath []string
var storeName string
var format string

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read value from a cache instance",

	RunE: func(cmd *cobra.Command, _ []string) error {
		var store cachepb.Store
		switch storeName {
		case "config":
			store = cachepb.Store_CONFIG
		case "state":
			store = cachepb.Store_STATE
		default:
			return fmt.Errorf("unknown store name: %s", storeName)
		}

		c, err := client.New(cmd.Context(), &client.ClientConfig{
			Address:       address,
			MaxReadStream: 1,
			Timeout:       timeout,
		})
		if err != nil {
			return err
		}

		paths := make([][]string, 0, len(readPath))
		for _, p := range readPath {
			paths = append(paths, strings.Split(p, ","))
		}
		for rs := range c.Read(cmd.Context(), cacheName, store, paths) {
			switch format {
			case "":
				fmt.Println(prototext.Format(rs))
			case "json":
				b, err := json.MarshalIndent(rs, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			case "flat":
				tv := new(sdcpb.TypedValue)
				err = proto.Unmarshal(rs.GetValue().GetValue(), tv)
				if err != nil {
					return err
				}
				fmt.Printf("%s: %s\n", strings.Join(rs.GetPath(), "/"), tvSPrint(tv))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&cacheName, "name", "n", "", "cache name")
	readCmd.Flags().StringArrayVarP(&readPath, "path", "p", []string{}, "paths to read")
	readCmd.Flags().StringVarP(&storeName, "store", "s", "config", "cache store to read from")
	readCmd.Flags().StringVarP(&format, "format", "", "", "print format, '', 'flat' or 'json'")
}

// TODO: finish all types
func tvSPrint(tv *sdcpb.TypedValue) string {
	switch v := tv.Value.(type) {
	case *sdcpb.TypedValue_AnyVal:
	case *sdcpb.TypedValue_AsciiVal:
		return v.AsciiVal
	case *sdcpb.TypedValue_BoolVal:
		return fmt.Sprintf("%t", v.BoolVal)
	case *sdcpb.TypedValue_BytesVal:
		return fmt.Sprintf("%b", v.BytesVal)
	case *sdcpb.TypedValue_DecimalVal:
	case *sdcpb.TypedValue_DoubleVal:
	case *sdcpb.TypedValue_FloatVal:
	case *sdcpb.TypedValue_IntVal:
		return fmt.Sprintf("%d", v.IntVal)
	case *sdcpb.TypedValue_JsonIetfVal:
		return string(v.JsonIetfVal)
	case *sdcpb.TypedValue_JsonVal:
		return string(v.JsonVal)
	case *sdcpb.TypedValue_LeaflistVal:
	case *sdcpb.TypedValue_ProtoBytes:
	case *sdcpb.TypedValue_StringVal:
		return v.StringVal
	case *sdcpb.TypedValue_UintVal:
		return fmt.Sprintf("%d", v.UintVal)
	}
	return ""
}
