package database

import (
	"log"

	"proceive/internal/modules/division"
	"proceive/internal/modules/item"
	"proceive/internal/modules/module"
	"proceive/internal/modules/permission"
	"proceive/internal/modules/position"
	"proceive/internal/modules/role"
	"proceive/internal/modules/user"
	"proceive/internal/modules/vendor"
	"proceive/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Seed menginisialisasi master data awal sistem Proceive
func Seed(db *gorm.DB) error {
	log.Println("🌱 Memulai seeding master data...")

	// Memastikan kolom DDL termutakhir di database
	db.Exec("ALTER TABLE vendors ADD COLUMN IF NOT EXISTS category VARCHAR(100);")
	db.Exec("ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(150);")

	// 1. Seed Divisions
	divisionsData := []division.Division{
		{Code: "IT", Name: "Information Technology"},
		{Code: "HRD", Name: "Human Resources Department"},
		{Code: "PROD", Name: "Produksi & Operasional"},
		{Code: "PURCH", Name: "Purchasing & Procurement"},
		{Code: "FIN", Name: "Finance & Accounting"},
		{Code: "PPIC", Name: "PPIC & Warehouse"},
	}
	divisionMap := make(map[string]uuid.UUID)
	for _, d := range divisionsData {
		var existing division.Division
		if err := db.Where("code = ?", d.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&d).Error; err != nil {
					return err
				}
				divisionMap[d.Code] = d.ID
			}
		} else {
			divisionMap[existing.Code] = existing.ID
		}
	}
	log.Println("  ✅ Divisions seeded")

	// 2. Seed Positions
	positionsData := []position.Position{
		{Code: "STF", Name: "Staff"},
		{Code: "SPV", Name: "Supervisor"},
		{Code: "MGR", Name: "Manager"},
		{Code: "DIR", Name: "Director"},
	}
	positionMap := make(map[string]uuid.UUID)
	for _, p := range positionsData {
		var existing position.Position
		if err := db.Where("code = ?", p.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&p).Error; err != nil {
					return err
				}
				positionMap[p.Code] = p.ID
			}
		} else {
			positionMap[existing.Code] = existing.ID
		}
	}
	log.Println("  ✅ Positions seeded")

	// 3. Seed Modules
	modulesData := []module.Module{
		{Slug: "auth", Name: "Authentication", Description: "Autentikasi dan login akun pengguna"},
		{Slug: "user", Name: "User Management", Description: "Pengelolaan pengguna internal dan vendor"},
		{Slug: "division", Name: "Division", Description: "Master divisi organisasi"},
		{Slug: "position", Name: "Position", Description: "Master jabatan organisasi"},
		{Slug: "role", Name: "Role & Permission", Description: "Manajemen peran hak akses sistem"},
		{Slug: "vendor", Name: "Vendor Management", Description: "Katalog supplier rekanan perusahaan"},
		{Slug: "item", Name: "Item Catalog", Description: "Katalog barang siap dipesan"},
		{Slug: "purchase_requisition", Name: "Purchase Requisition", Description: "Pengajuan permintaan barang (PR)"},
		{Slug: "purchase_order", Name: "Purchase Order", Description: "Penerbitan pesanan pembelian resmi (PO)"},
		{Slug: "goods_receipt", Name: "Goods Receipt", Description: "Verifikasi surat jalan dan penerimaan fisik (GR)"},
	}
	moduleMap := make(map[string]uuid.UUID)
	for _, m := range modulesData {
		var existing module.Module
		if err := db.Where("slug = ?", m.Slug).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&m).Error; err != nil {
					return err
				}
				moduleMap[m.Slug] = m.ID
			}
		} else {
			moduleMap[existing.Slug] = existing.ID
		}
	}
	log.Println("  ✅ Modules seeded")

	// 4. Seed Permissions
	type permInput struct {
		ModuleSlug  string
		Code        string
		Name        string
		Description string
	}
	permissionsData := []permInput{
		// Master Users
		{"user", "user.view", "View Users", "Melihat data user"},
		{"user", "user.create", "Create User", "Menambah user baru"},
		{"user", "user.update", "Update User", "Mengubah data user"},
		{"user", "user.delete", "Delete User", "Menghapus data user"},

		// Master Divisions
		{"division", "division.view", "View Divisions", "Melihat data divisi"},
		{"division", "division.create", "Create Division", "Menambah divisi baru"},

		// Master Positions
		{"position", "position.view", "View Positions", "Melihat data jabatan"},
		{"position", "position.create", "Create Position", "Menambah jabatan baru"},

		// Master Vendors
		{"vendor", "vendor.view", "View Vendors", "Melihat data vendor"},
		{"vendor", "vendor.create", "Create Vendor", "Menambah vendor baru"},

		// Master Items
		{"item", "item.view", "View Items", "Melihat katalog barang"},
		{"item", "item.create", "Create Item", "Menambah barang katalog baru"},

		// Purchase Requisitions
		{"purchase_requisition", "purchase_requisition.view", "View Purchase Requisition", "Melihat daftar/detail PR"},
		{"purchase_requisition", "purchase_requisition.create", "Create Purchase Requisition", "Membuat pengajuan PR"},
		{"purchase_requisition", "purchase_requisition.approve", "Approve/Reject Purchase Requisition", "Menyetujui atau menolak PR"},

		// Purchase Orders
		{"purchase_order", "purchase_order.view", "View Purchase Order", "Melihat daftar/detail PO"},
		{"purchase_order", "purchase_order.create", "Create Purchase Order", "Membuat PO dari PR yang approved"},

		// Goods Receipts
		{"goods_receipt", "goods_receipt.view", "View Goods Receipt", "Melihat daftar status GR"},
		{"goods_receipt", "goods_receipt.upload_sj", "Upload Surat Jalan", "Vendor mengunggah dokumen Surat Jalan"},
		{"goods_receipt", "goods_receipt.verify_sj", "Verify Surat Jalan", "Purchasing memverifikasi Surat Jalan dari vendor"},
		{"goods_receipt", "goods_receipt.complete", "Complete Goods Receipt", "PPIC/Warehouse menyelesaikan penerimaan barang fisik"},
	}

	permMap := make(map[string]uuid.UUID)
	for _, p := range permissionsData {
		modID, exists := moduleMap[p.ModuleSlug]
		if !exists {
			continue
		}
		var existing permission.Permission
		if err := db.Where("code = ?", p.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				newPerm := permission.Permission{
					ModuleID:    modID,
					Code:        p.Code,
					Name:        p.Name,
					Description: p.Description,
				}
				if err := db.Create(&newPerm).Error; err != nil {
					return err
				}
				permMap[p.Code] = newPerm.ID
			}
		} else {
			permMap[existing.Code] = existing.ID
		}
	}
	log.Println("  ✅ Permissions seeded")

	// 5. Seed Roles
	rolesData := []role.Role{
		{Code: "ADMIN", Name: "Administrator", Description: "Akses penuh ke seluruh sistem"},
		{Code: "REQUESTER", Name: "Requester (Staf)", Description: "Staf pembuat pengajuan permintaan barang (PR)"},
		{Code: "APPROVER", Name: "Approver (Manager)", Description: "Manager penyetuju atau penolak pengajuan PR"},
		{Code: "PURCHASING", Name: "Purchasing", Description: "Tim pengadaan pembuat PO dan verifikator Surat Jalan"},
		{Code: "VENDOR", Name: "Vendor Supplier", Description: "Penyedia barang yang mengunggah Surat Jalan"},
		{Code: "PPIC", Name: "PPIC / Warehouse", Description: "Petugas gudang penerima fisik barang (Goods Receipt)"},
	}
	roleMap := make(map[string]uuid.UUID)
	for _, r := range rolesData {
		var existing role.Role
		if err := db.Where("code = ?", r.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&r).Error; err != nil {
					return err
				}
				roleMap[r.Code] = r.ID
			}
		} else {
			roleMap[existing.Code] = existing.ID
		}
	}
	log.Println("  ✅ Roles seeded")

	// 6. Assign Permissions to Roles
	rolePermAssignments := map[string][]string{
		"ADMIN": {
			"user.view", "user.create", "user.update", "user.delete",
			"division.view", "division.create",
			"position.view", "position.create",
			"vendor.view", "vendor.create",
			"item.view", "item.create",
			"purchase_requisition.view", "purchase_requisition.create", "purchase_requisition.approve",
			"purchase_order.view", "purchase_order.create",
			"goods_receipt.view", "goods_receipt.upload_sj", "goods_receipt.verify_sj", "goods_receipt.complete",
		},
		"REQUESTER": {
			"item.view",
			"purchase_requisition.view", "purchase_requisition.create",
		},
		"APPROVER": {
			"item.view",
			"purchase_requisition.view", "purchase_requisition.approve",
		},
		"PURCHASING": {
			"item.view", "vendor.view",
			"purchase_requisition.view",
			"purchase_order.view", "purchase_order.create",
			"goods_receipt.view", "goods_receipt.verify_sj",
		},
		"VENDOR": {
			"purchase_order.view",
			"goods_receipt.view", "goods_receipt.upload_sj",
		},
		"PPIC": {
			"purchase_order.view",
			"goods_receipt.view", "goods_receipt.complete",
		},
	}

	for roleCode, permCodes := range rolePermAssignments {
		roleID, ok := roleMap[roleCode]
		if !ok {
			continue
		}
		for _, permCode := range permCodes {
			permID, ok := permMap[permCode]
			if !ok {
				continue
			}
			var rhp role.RoleHasPermission
			if err := db.Where("role_id = ? AND permission_id = ?", roleID, permID).First(&rhp).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					db.Create(&role.RoleHasPermission{RoleID: roleID, PermissionID: permID})
				}
			}
		}
	}
	log.Println("  ✅ Role permissions mapped")

	// 7. Seed Vendors
	_ = db.AutoMigrate(&vendor.Vendor{})
	_ = db.Exec("ALTER TABLE vendors ADD COLUMN IF NOT EXISTS category VARCHAR(100);").Error

	vendorsData := []vendor.Vendor{
		{
			Name:          "PT Maju Jaya Komputindo",
			Category:      "Hardware & Komputer",
			ContactPerson: "Budi Santoso",
			Email:         "vendor1@proceive.com",
			Address:       "Jl. Jend. Sudirman Kav. 45, Jakarta Selatan",
			IsActive:      true,
		},
		{
			Name:          "CV Sumber Berkah Logistik",
			Category:      "Logistik & Ekspedisi",
			ContactPerson: "Siti Rahma",
			Email:         "vendor2@proceive.com",
			Address:       "Kawasan Industri Jababeka Blok C, Cikarang, Bekasi",
			IsActive:      true,
		},
		{
			Name:          "PT Nusantara Solusi Informatika",
			Category:      "Software & Cloud Services",
			ContactPerson: "Eko Prasetyo",
			Email:         "vendor3@proceive.com",
			Address:       "Gedung Cyber 2 Lantai 18, Jl. HR Rasuna Said, Jakarta Selatan",
			IsActive:      true,
		},
		{
			Name:          "PT Tri Megah Stationery",
			Category:      "Alat Tulis Kantor (ATK)",
			ContactPerson: "Lina Marlina",
			Email:         "vendor4@proceive.com",
			Address:       "Komp. Perkantoran Roxy Mas Blok D2/15, Jakarta Pusat",
			IsActive:      true,
		},
		{
			Name:          "CV Cipta Kreasi Furniture",
			Category:      "Perabot & Interior Kantor",
			ContactPerson: "Hendra Wijaya",
			Email:         "vendor5@proceive.com",
			Address:       "Jl. Rungkut Industri III No. 8, Surabaya",
			IsActive:      true,
		},
		{
			Name:          "PT Global Teknik Sentosa",
			Category:      "Mesin Industri & Sparepart",
			ContactPerson: "Agus Setiawan",
			Email:         "vendor6@proceive.com",
			Address:       "Kawasan Industri MM2100 Blok KK, Cikarang Barat",
			IsActive:      true,
		},
		{
			Name:          "PT Sinar Indah Kemasan",
			Category:      "Packaging & Kemasan Industri",
			ContactPerson: "Dewi Lestari",
			Email:         "vendor7@proceive.com",
			Address:       "Jl. Raya Serang Km 14, Cikupa, Tangerang",
			IsActive:      true,
		},
		{
			Name:          "CV Prima Mandiri Elektrik",
			Category:      "Kelistrikan & Genset",
			ContactPerson: "Taufik Hidayat",
			Email:         "vendor8@proceive.com",
			Address:       "Jl. Soekarno Hatta No. 512, Bandung",
			IsActive:      true,
		},
		{
			Name:          "PT Daya Cipta Mandiri Safety",
			Category:      "Alat Pelindung Diri (APD) & K3",
			ContactPerson: "Ahmad Fauzi",
			Email:         "vendor9@proceive.com",
			Address:       "Sentra Niaga Puri Indah Blok T2/8, Jakarta Barat",
			IsActive:      true,
		},
		{
			Name:          "PT Clean Solution Indonesia",
			Category:      "Chemical & Alat Kebersihan",
			ContactPerson: "Maria Ulfah",
			Email:         "vendor10@proceive.com",
			Address:       "Kawasan Industri Candi Blok 10A, Semarang",
			IsActive:      true,
		},
		{
			Name:          "CV Mitra Sejahtera Telematika",
			Category:      "Jaringan Komputer & Fiber Optik",
			ContactPerson: "Rian Nugroho",
			Email:         "vendor11@proceive.com",
			Address:       "Jl. Ring Road Utara No. 99, Sleman, D.I. Yogyakarta",
			IsActive:      true,
		},
		{
			Name:          "PT Pangan Berkah Utama",
			Category:      "Katering & Konsumsi Karyawan",
			ContactPerson: "Nurul Aini",
			Email:         "vendor12@proceive.com",
			Address:       "Jl. Raya Pulogebang Indah No. 24, Jakarta Timur",
			IsActive:      true,
		},
		{
			Name:          "PT Multi Bangun Perkasa",
			Category:      "Material Konstruksi & Sipil",
			ContactPerson: "Joko Susilo",
			Email:         "vendor13@proceive.com",
			Address:       "Jl. Krakatau No. 128, Medan",
			IsActive:      true,
		},
		{
			Name:          "CV Bahari Trans Cargo",
			Category:      "Kargo & Ekspedisi Laut",
			ContactPerson: "Andi Mallarangeng",
			Email:         "vendor14@proceive.com",
			Address:       "Pelabuhan Soekarno Hatta Area Gudang 5, Makassar",
			IsActive:      true,
		},
		{
			Name:          "PT Delta Sukses Otomasi",
			Category:      "Otomasi Industri & IoT",
			ContactPerson: "Kevin Tanjaya",
			Email:         "vendor15@proceive.com",
			Address:       "Batamindo Industrial Park Lot 325, Mukakuning, Batam",
			IsActive:      true,
		},
		{
			Name:          "PT Bintang Terang Percetakan",
			Category:      "Percetakan, Label & Offset",
			ContactPerson: "Sri Wahyuni",
			Email:         "vendor16@proceive.com",
			Address:       "Jl. Slamet Riyadi No. 410, Surakarta",
			IsActive:      true,
		},
		{
			Name:          "CV Aneka Baja Konstruksi",
			Category:      "Besi, Baja & Pipa Logam",
			ContactPerson: "Bambang Irawan",
			Email:         "vendor17@proceive.com",
			Address:       "Kawasan Industri Berbek Blok C-12, Sidoarjo",
			IsActive:      true,
		},
		{
			Name:          "PT Surya Medika Solusindo",
			Category:      "Alat Kesehatan & P3K",
			ContactPerson: "drg. Diana Puspita",
			Email:         "vendor18@proceive.com",
			Address:       "Jl. Ijen No. 77, Malang",
			IsActive:      true,
		},
		{
			Name:          "CV Garda Security Utama",
			Category:      "Sistem Keamanan & CCTV",
			ContactPerson: "Wayan Darmawan",
			Email:         "vendor19@proceive.com",
			Address:       "Jl. Sunset Road No. 88, Kuta, Bali",
			IsActive:      true,
		},
		{
			Name:          "PT Eco Waste Solusindo",
			Category:      "Pengelolaan Limbah & Daur Ulang",
			ContactPerson: "Zulkifli Lubis",
			Email:         "vendor20@proceive.com",
			Address:       "Kawasan Pelabuhan Boom Baru, Palembang",
			IsActive:      true,
		},
	}
	vendorMap := make(map[string]uuid.UUID)
	for _, v := range vendorsData {
		var existing vendor.Vendor
		if err := db.Where("email = ?", v.Email).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&v).Error; err != nil {
					return err
				}
				vendorMap[v.Email] = v.ID
			}
		} else {
			// Selalu sinkronkan kolom category dan field lainnya untuk seluruh vendor yang sudah ada
			db.Model(&vendor.Vendor{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
				"category":       v.Category,
				"name":           v.Name,
				"contact_person": v.ContactPerson,
				"address":        v.Address,
				"is_active":      v.IsActive,
			})
			vendorMap[existing.Email] = existing.ID
		}
	}
	log.Println("  ✅ Vendors seeded with category")

	// 8. Seed Items
	itemsData := []item.Item{
		{Code: "IT-001", Name: "Laptop Lenovo ThinkPad T14", Category: "Hardware", UOM: "UNIT", PriceEstimate: 18500000},
		{Code: "IT-002", Name: "Monitor Dell UltraSharp 27 inch 4K", Category: "Hardware", UOM: "UNIT", PriceEstimate: 5200000},
		{Code: "ATK-001", Name: "Kertas HVS PaperOne A4 80gr", Category: "ATK", UOM: "RIM", PriceEstimate: 55000},
		{Code: "ATK-002", Name: "Spidol Whiteboard Snowman (Pack 12 pcs)", Category: "ATK", UOM: "BOX", PriceEstimate: 95000},
		{Code: "PROD-001", Name: "Pallet Plastik Heavy Duty 120x100 cm", Category: "Gudang", UOM: "PCS", PriceEstimate: 380000},
	}
	for _, itm := range itemsData {
		var existing item.Item
		if err := db.Where("code = ?", itm.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&itm).Error; err != nil {
					return err
				}
			}
		}
	}
	log.Println("  ✅ Items seeded")

	// 9. Seed Users
	defaultPassword := "password123"
	hashedPassword, err := utils.HashPassword(defaultPassword)
	if err != nil {
		return err
	}

	itDivID := divisionMap["IT"]
	purchDivID := divisionMap["PURCH"]
	ppicDivID := divisionMap["PPIC"]

	stfPosID := positionMap["STF"]
	mgrPosID := positionMap["MGR"]

	usersData := []user.User{
		{
			Name:         "Super Administrator",
			Email:        "admin@proceive.com",
			PasswordHash: hashedPassword,
			RoleID:       roleMap["ADMIN"],
			DivisionID:   &itDivID,
			PositionID:   &mgrPosID,
		},
		{
			Name:         "Budi Requester (IT Staff)",
			Email:        "requester@proceive.com",
			PasswordHash: hashedPassword,
			RoleID:       roleMap["REQUESTER"],
			DivisionID:   &itDivID,
			PositionID:   &stfPosID,
		},
		{
			Name:         "Rudi Approver (IT Manager)",
			Email:        "approver@proceive.com",
			PasswordHash: hashedPassword,
			RoleID:       roleMap["APPROVER"],
			DivisionID:   &itDivID,
			PositionID:   &mgrPosID,
		},
		{
			Name:         "Citra Purchasing (Staff)",
			Email:        "purchasing@proceive.com",
			PasswordHash: hashedPassword,
			RoleID:       roleMap["PURCHASING"],
			DivisionID:   &purchDivID,
			PositionID:   &stfPosID,
		},
		{
			Name:         "Doni Warehouse (PPIC Staff)",
			Email:        "ppic@proceive.com",
			PasswordHash: hashedPassword,
			RoleID:       roleMap["PPIC"],
			DivisionID:   &ppicDivID,
			PositionID:   &stfPosID,
		},
	}

	// Buat akun pengguna untuk setiap vendor menggunakan nama perusahaan vendor masing-masing
	vendorRoleID := roleMap["VENDOR"]
	for _, v := range vendorsData {
		vID := vendorMap[v.Email]
		usersData = append(usersData, user.User{
			Name:         v.Name, // Nama asli perusahaan vendor
			Email:        v.Email, // Email asli vendor yang sama persis
			PasswordHash: hashedPassword,
			RoleID:       vendorRoleID,
			VendorID:     &vID,
		})
	}

	// 1. Eksekusi SQL langsung: Sinkronkan SEMUA akun vendor di tabel users agar name dan email PERSIS SAMA dengan nama perusahaan dan email di tabel vendors
	db.Exec(`
		UPDATE users 
		SET name = vendors.name, email = vendors.email 
		FROM vendors 
		WHERE users.vendor_id = vendors.id;
	`)

	// 2. Loop insert / update usersData
	for _, u := range usersData {
		var existing user.User
		var query *gorm.DB
		if u.VendorID != nil {
			query = db.Where("vendor_id = ?", *u.VendorID)
		} else {
			query = db.Where("email = ?", u.Email)
		}

		if err := query.First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&u).Error; err != nil {
					return err
				}
			}
		} else {
			// Perbarui nama dan email agar selalu sinkron dengan nama perusahaan vendor
			db.Model(&user.User{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
				"name":      u.Name,
				"email":     u.Email,
				"role_id":   u.RoleID,
				"vendor_id": u.VendorID,
			})
		}
	}

	// 3. Bersihkan akun legacy seperti 'vendor@proceive.com' jika masih ada
	db.Exec("DELETE FROM users WHERE email = 'vendor@proceive.com'")

	log.Println("  ✅ Users seeded with vendor company names and matching emails")
	log.Println("✨ Seeding data berhasil diselesaikan!")
	return nil
}
