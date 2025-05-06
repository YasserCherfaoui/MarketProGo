package database

import (
	"fmt"
	"os"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {

	var dsn string
	if gin.Mode() == gin.ReleaseMode {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s TimeZone=Asia/Shanghai",
			os.Getenv("PGHOST"),
			os.Getenv("PGUSER"),
			os.Getenv("PGPASSWORD"),
			os.Getenv("PGDATABASE"),
			os.Getenv("PGPORT"),
		)
	} else {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
		fmt.Printf("RUNNING IN DEV MODE %s", os.Getenv("DB_NAME"))
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s TimeZone=Asia/Shanghai",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Address{},
		&models.Product{},
		&models.ProductImage{},
		&models.Category{},
		&models.InventoryItem{},
		&models.Warehouse{},
		&models.ProductSpecification{},
		&models.Order{},
		&models.OrderItem{},
		&models.Invoice{},
		&models.PurchaseOrder{},
		&models.POItem{},
		&models.Supplier{},
		&models.SupplierContact{},
		&models.Document{},
		&models.Contract{},
		&models.ContractItem{},
		&models.ContractSchedule{},
		&models.ContractOrder{},
	)

	return db, nil
}
