package database

import (
	"fmt"
	"log"

	"proceive/internal/modules/division"
	goodsreceipt "proceive/internal/modules/goods_receipt"
	"proceive/internal/modules/item"
	"proceive/internal/modules/module"
	"proceive/internal/modules/permission"
	"proceive/internal/modules/position"
	purchaseorder "proceive/internal/modules/purchase_order"
	purchaserequisition "proceive/internal/modules/purchase_requisition"
	"proceive/internal/modules/role"
	"proceive/internal/modules/user"
	"proceive/internal/modules/vendor"

	"gorm.io/gorm"
)

// Migrate menjalankan migrasi skema database untuk semua tabel modul
func Migrate(db *gorm.DB) error {
	log.Println("🔄 Memulai migrasi database PostgreSQL...")

	// Daftar entitas sesuai urutan relasi (master data -> RBAC -> transaksional)
	models := []interface{}{
		// Master Data Organisasi
		&division.Division{},
		&position.Position{},

		// RBAC
		&module.Module{},
		&permission.Permission{},
		&role.Role{},
		&role.RoleHasPermission{},

		// Master Katalog & Vendor
		&vendor.Vendor{},
		&item.Item{},

		// Pengguna Sistem
		&user.User{},

		// Transaksional Pengadaan
		&purchaserequisition.PurchaseRequisition{},
		&purchaserequisition.PRLine{},
		&purchaseorder.PurchaseOrder{},
		&purchaseorder.POLine{},

		// Transaksional Logistik Penerimaan
		&goodsreceipt.GoodsReceipt{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("gagal migrasi model %T: %w", model, err)
		}
	}

	log.Println("✅ Semua tabel berhasil dimigrasi!")
	return nil
}
