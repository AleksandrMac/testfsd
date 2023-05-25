package main

import (
	"github.com/spf13/cobra"
)

var (
	migrateCMD = &cobra.Command{
		Use:   "migrate",
		Short: "migrate db schema and seed data",
		Run:   migrateCMDRunner,
	}
)

func init() {
	rootCmd.AddCommand(migrateCMD)
}

func migrateCMDRunner(cmd *cobra.Command, agrs []string) {
	// log.Info("Start migrate")
	// db, err := gorm.Open(postgres.Open(config.Default.Database.URL))
	// if err != nil {
	// 	log.Fatal(
	// 		fmt.Sprintf("Failed to connect database[%v]: %v", config.Parse().Database.URL, err))
	// }
	// defer func() {
	// 	if dbSql, err := db.DB(); err != nil {
	// 		dbSql.Close()
	// 	}
	// }()
	// entities.AutoMigrate(db)
}
