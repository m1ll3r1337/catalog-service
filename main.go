package main

import (
	"context"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/m1ll3r1337/catalog-service/internal/app/config"
	"github.com/m1ll3r1337/catalog-service/internal/app/entity"
	rhealth "github.com/m1ll3r1337/catalog-service/internal/app/handler/http/health"
	rprocessor "github.com/m1ll3r1337/catalog-service/internal/app/processor/http"
	pcategory "github.com/m1ll3r1337/catalog-service/internal/app/repository/category"
	rcpostgres "github.com/m1ll3r1337/catalog-service/internal/app/repository/conn/postgres"
	pproduct "github.com/m1ll3r1337/catalog-service/internal/app/repository/product"
)

func main() {
	ctx := context.Background()
	config.Load()
	cfg := config.Root

	pgClient, err := rcpostgres.NewClient(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if oldVer != newVer {
		log.Printf("Database migrated old_version=%d new_version=%d", oldVer, newVer)
	} else {
		log.Printf("Database is up to date version=%d", newVer)
	}

	// === Временная проверка репозитория ===

	categoryRepo := pcategory.NewRepoFromPostgres(pgClient)
	productRepo := pproduct.NewRepoFromPostgres(pgClient)

	// 1. Создание категории
	cat := entity.Category{
		GUID:      uuid.Must(uuid.NewV4()),
		Name:      "Электроника",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := categoryRepo.Create(ctx, cat); err != nil {
		log.Fatalf("Category create failed: %v", err)
	}
	log.Printf("Category created guid=%s", cat.GUID.String())

	// 2. Получение категории
	foundCats, err := categoryRepo.GetByGUIDs(ctx, []uuid.UUID{cat.GUID})
	if err != nil {
		log.Fatalf("Category GetByGUIDs failed: %v", err)
	}
	if len(foundCats) == 0 {
		log.Fatalf("Category not found guid=%s", cat.GUID.String())
	}
	found := foundCats[0]
	log.Printf("Category found name=%s", found.Name)

	// 3. Обновление категории
	found.Name = "Бытовая техника"
	found.UpdatedAt = time.Now()
	if err := categoryRepo.Update(ctx, found); err != nil {
		log.Fatalf("Category update failed: %v", err)
	}
	log.Println("Category updated")

	// 4. List — все категории
	allCats, err := categoryRepo.List(ctx, nil)
	if err != nil {
		log.Fatalf("Category list failed: %v", err)
	}
	log.Printf("All categories count=%d", len(allCats))

	// 5. List — фильтр по имени
	filterName := "Бытовая техника"
	filtered, err := categoryRepo.List(ctx, &filterName)
	if err != nil {
		log.Fatalf("Category list by name failed: %v", err)
	}
	log.Printf("Filtered categories count=%d", len(filtered))

	// 6. Создание продукта
	desc := "Мощный пылесос"
	prod := entity.Product{
		GUID:         uuid.Must(uuid.NewV4()),
		Name:         "Пылесос Dyson V15",
		Description:  &desc,
		Price:        4999999,
		CategoryGUID: cat.GUID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := productRepo.Create(ctx, prod); err != nil {
		log.Fatalf("Product create failed: %v", err)
	}
	log.Printf("Product created guid=%s", prod.GUID.String())

	// 7. List продуктов по категории
	products, err := productRepo.List(ctx, nil, &cat.GUID)
	if err != nil {
		log.Fatalf("Product list by category failed: %v", err)
	}
	log.Printf("Products in category count=%d", len(products))

	// 8. Удаление продукта, затем категории
	if err := productRepo.Delete(ctx, prod.GUID); err != nil {
		log.Fatalf("Product delete failed: %v", err)
	}
	if err := categoryRepo.Delete(ctx, cat.GUID); err != nil {
		log.Fatalf("Category delete failed: %v", err)
	}
	log.Println("Cleanup complete")

	// === Конец проверки ===

	hHealth := rhealth.NewHandler()

	httpServer := rprocessor.NewHTTP(hHealth, cfg.Processor.WebServer)
	if err := httpServer.Serve(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
