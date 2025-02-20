package db

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"red-packet-system/config"
	"red-packet-system/pkg/logger"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// InitDB initializes MySQL read/write separation
func InitDB(cfg *config.Config) error {
	var err error
	log := logger.GetLogger()

	once.Do(func() {
		// Debug logs to verify connection details
		log.Printf("Connecting to MySQL Master: %s", cfg.DBMaster)
		log.Printf("Connecting to MySQL Slave: %s", cfg.DBSlave)

		// MySQL Master DSN (Write)
		dsnMaster := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBMaster, cfg.DBName,
		)

		// MySQL Slave DSN (Read)
		dsnSlave := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBSlave, cfg.DBName,
		)

		// Connect to MySQL Master
		DB, err = gorm.Open(mysql.Open(dsnMaster), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to MySQL Master at %s: %v", cfg.DBMaster, err)
		}

		// Configure read/write separation
		err = DB.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(dsnMaster)}, // Master (Write)
			Replicas: []gorm.Dialector{mysql.Open(dsnSlave)},  // Slave (Read)
			Policy:   dbresolver.RandomPolicy{},               // Load balancing policy
		}))

		if err != nil {
			log.Fatalf("Failed to set up read/write splitting: %v", err)
		}

		log.Println("MySQL Read/Write Splitting Configured")
	})

	return err
}

// GetDB returns the MySQL connection (Singleton)
func GetDB() *gorm.DB {
	log := logger.GetLogger()

	if DB == nil {
		log.Println("MySQL is not initialized, please run InitDB() first")
	}
	return DB
}

// CloseDB closes the MySQL connection
func CloseDB() error {
	log := logger.GetLogger()

	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("Failed to get DB connection: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("Failed to close database: %v", err)
	}

	log.Println("MySQL connection closed")
	return nil
}
