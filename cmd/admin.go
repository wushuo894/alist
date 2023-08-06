/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/spf13/cobra"
)

// PasswordCmd represents the password command
var PasswordCmd = &cobra.Command{
	Use:     "admin",
	Aliases: []string{"password"},
	Short:   "Show admin user's info",
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		admin, err := op.GetAdmin()
		if err != nil {
			utils.Log.Errorf("failed get admin user: %+v", err)
		} else {
			utils.Log.Infof("Admin user's username: %s", admin.Username)
			utils.Log.Infof("The password can only be output at the first startup, and then stored as a hash value, which cannot be reversed")
			utils.Log.Infof("You can reset the password by running 'alist password random' to generate a random password")
			utils.Log.Infof("You can also set the password by running 'alist password set NEW_PASSWORD' to set a new password")
		}
	},
}

var RandomPasswordCmd = &cobra.Command{
	Use:   "random",
	Short: "Reset admin user's password to a random string",
	Run: func(cmd *cobra.Command, args []string) {
		newPwd := random.String(8)
		setAdminPassword(newPwd)
	},
}

var SetPasswordCmd = &cobra.Command{
	Use:   "set",
	Short: "Set admin user's password",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			utils.Log.Errorf("Please enter the new password")
			return
		}
		setAdminPassword(args[0])
	},
}

func setAdminPassword(pwd string) {
	Init()
	admin, err := op.GetAdmin()
	if err != nil {
		utils.Log.Errorf("failed get admin user: %+v", err)
		return
	}
	admin.PwdHash = model.HashPwd(pwd)
	if err := op.UpdateUser(admin); err != nil {
		utils.Log.Errorf("failed update admin user: %+v", err)
		return
	}
	utils.Log.Infof("admin user has been updated:")
	utils.Log.Infof("username: %s", admin.Username)
	utils.Log.Infof("password: %s", pwd)
	DelAdminCacheOnline()
}

func init() {
	RootCmd.AddCommand(PasswordCmd)
	PasswordCmd.AddCommand(RandomPasswordCmd)
	PasswordCmd.AddCommand(SetPasswordCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passwordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passwordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
