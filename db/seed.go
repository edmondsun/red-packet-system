package db

import (
	"fmt"
	"log"
	"math/rand"
	"red-packet-system/model"

	"github.com/bxcodec/faker/v3"
)

// SeedUsers generate fake users
func SeedUsers(count int) {
	db := GetDB()

	for i := 0; i < count; i++ {
		user := model.User{
			Username: faker.Username(),
			Balance:  float64(rand.Intn(1000)) + rand.Float64(),
		}
		db.Create(&user)
		fmt.Printf("Inserted User: %s with Balance: %.2f\n", user.Username, user.Balance)
	}
}

// SeedRedPackets generate fake red packets
func SeedRedPackets(count int) {
	db := GetDB()

	for i := 0; i < count; i++ {
		totalAmount := float64(rand.Intn(500) + 100)
		totalCount := rand.Intn(10) + 1
		redPacket := model.RedPacket{
			TotalAmount:     totalAmount,
			RemainingAmount: totalAmount,
			TotalCount:      totalCount,
			RemainingCount:  totalCount,
			Status:          1,
		}
		db.Create(&redPacket)
		fmt.Printf("Inserted Red Packet: TotalAmount: %.2f, TotalCount: %d\n", totalAmount, totalCount)
	}
}

// RunSeeds execute
func RunSeeds() {
	log.Println("Seeding database...")
	SeedUsers(10)
	SeedRedPackets(5)
	log.Println("Seeding completed!")
}
