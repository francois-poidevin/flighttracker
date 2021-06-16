package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	defaults "github.com/mcuadros/go-defaults"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------

var configNewAsEnvFlag bool

// -----------------------------------------------------------------------------

var configNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Initialize a default configuration",
	Run: func(cmd *cobra.Command, args []string) {
		defaults.SetDefaults(conf)

		if !configNewAsEnvFlag {
			btes, err := toml.Marshal(*conf)
			if err != nil {
				fmt.Printf("Error during configuration export")
			}
			fmt.Println(string(btes))
		} else {
			m := asEnvVariables(conf, "EH", true)
			keys := []string{}

			for k := range m {
				keys = append(keys, k)
			}

			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("export %s=\"%s\"\n", k, m[k])
			}
		}
	},
}

// asEnvVariables sets struct values from environment variables
func asEnvVariables(o interface{}, prefix string, skipCommented bool) map[string]string {
	r := map[string]string{}
	prefix = strings.ToUpper(prefix)
	delim := "_"
	if prefix == "" {
		delim = ""
	}
	fields := structs.Fields(o)
	for _, f := range fields {
		if skipCommented {
			tag := f.Tag("commented")
			if tag != "" {
				commented, err := strconv.ParseBool(tag)
				fmt.Println(err)
				// log.CheckErr("Unable to parse tag value", err)
				if commented {
					continue
				}
			}
		}
		if structs.IsStruct(f.Value()) {
			rf := asEnvVariables(f.Value(), prefix+delim+f.Name(), skipCommented)
			for k, v := range rf {
				r[k] = v
			}
		} else {
			r[prefix+"_"+strings.ToUpper(f.Name())] = fmt.Sprintf("%v", f.Value())
		}
	}
	return r
}
