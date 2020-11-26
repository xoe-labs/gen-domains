/*
Copyright © 2020 David Arnold <dar@xoe.solutions>

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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xoe-labs/ddd-gen/pkg/gen_app"
)

// appCommandCmd represents the app command
var appCommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Generates application command handler wrapper",
	Long: `Generates application command handler wrapper that assert authorization and interface with storage.

  Available Annotations:
    Key "command" | Separator: ";"
      w/o policy              - handler that will not know how to check access against the policy interface
      topic,<topic>           - topic is generated from the last Word, if not desired, it can be mannually overridden
      adapters,key:import/path,key2:import/path2
                              - add additional domain service adapters for this command handler

  Config File: (will be complemented by this command)

    # ./ddd-config.yaml

    # Objects
    entity:                       "github.com/xoe-labs/ddd-gen/internal/test-svc/domain/account.Account"

    # Error Contructors
    authorizationErrorNew:        "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors.NewAuthorizationError"
    targetIdentificationErrorNew: "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors.NewTargetIdentificationError"
    storageLoadingErrorNew:       "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors.NewStorageLoadingError"
    storageSavingErrorNew:        "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors.NewStorageSavingError"
    domainErrorNew:               "github.com/xoe-labs/ddd-gen/internal/test-svc/app/errors.NewDomainError"

  Expected / Recomended Folder Structure:
    ./app
    ├── command
    │   ├── commands.go             // define your tags here (see example)
    │   ├── make_new_account_gen.go // generated by this command
    │   └── ...                     // generated by this command
    ├── storage.go                  // generated storage interface
    ├── policy.go                   // generated policy interface
    ├── domain.go                   // generated domain interface
    ├── identiy.go                  // generated identity assertion interface
    ├── distinguishable.go          // generated stub of distinguishable interface (edit & implement!)
    ├── authorizable.go             // generated stub of authorizable interface (edit & implement!)
    ├── authorizable
    │   └── actor.go                // implement your actor here
    ├── distinguishable
    │   └── target.go               // implement your target here
    ├── error
    │   └── errors.go               // implement your errors & their constructors here
    └── ...`,
	Example: `  Command:
    //go:generate go run github.com/xoe-labs/ddd-gen --config ../../ddd-config.yaml app command --type Commands

  Code:
    type Commands struct {
      MakeNewAccount          MakeNewAccountHandlerWrapper          ` + "`" + `` + "`" + `
      MakeNewAccountWithOutId MakeNewAccountWithOutIdHandlerWrapper ` + "`" + `command:topic,account"` + "`" + `
      DeleteAccount           DeleteAccountHandlerWrapper           ` + "`" + `` + "`" + `
      BlockAccount            BlockAccountHandlerWrapper            ` + "`" + `` + "`" + `
      ValidateHolder          ValidateHandlerWrapper                ` + "`" + `command:"w/o policy"` + "`" + `
      IncreaseBalance         IncreaseBalanceHandlerWrapper         ` + "`" + `` + "`" + `
      IncreaseBalanceFromSvc  IncreaseBalanceFromSvcHandlerWrapper  ` + "`" + `command:"topic,balance; adapters,svc:github.com/xoe-labs/ddd-gen/internal/test-svc/adapter/balancesvc.Balancer"` + "`" + `
    }
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := gen_app.NewConfig(
			viper.GetString("entity"),
			viper.GetString("domain"),
			viper.GetString("authorizationErrorNew"),
			viper.GetString("targetIdentificationErrorNew"),
			viper.GetString("storageLoadingErrorNew"),
			viper.GetString("storageSavingErrorNew"),
			viper.GetString("domainErrorNew"),
		)
		if err != nil {
			return err
		}
		return gen_app.Gen(sourceType, useFactStorage, cfg)
	},
}

func init() {
	appCmd.AddCommand(appCommandCmd)
	appCommandCmd.Flags().BoolVarP(&useFactStorage, "fact-based", "f", false, "Event sourcing variant")
}
