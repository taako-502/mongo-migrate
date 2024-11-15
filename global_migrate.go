package migrate

import (
	"context"
	"fmt"
	"runtime"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var globalMigrate = NewMigrate(nil)

func internalRegister(up, down MigrationFunc, skip int) error {
	_, file, _, _ := runtime.Caller(skip)
	version, description, err := extractVersionDescription(file)
	if err != nil {
		return err
	}
	if hasVersion(globalMigrate.migrations, version) {
		return fmt.Errorf("migration with version %v already registered", version)
	}
	globalMigrate.migrations = append(globalMigrate.migrations, Migration{
		Version:     version,
		Description: description,
		Up:          up,
		Down:        down,
	})
	return nil
}

// Register performs migration registration.
// Use case of this function:
//
// - Create a file called like "1_setup_indexes.go" ("<version>_<comment>.go").
//
// - Use the following template inside:
//
//	 package migrations
//
//	 import (
//		 "go.mongodb.org/mongo-driver/v2/bson"
//		 "go.mongodb.org/mongo-driver/v2/mongo"
//		 "go.mongodb.org/mongo-driver/v2/mongo/options"
//		 migrate "github.com/xakep666/mongo-migrate"
//	 )
//
//	 func init() {
//		 migrate.Register(func(ctx context.Context, db *mongo.Database) error {
//		 	 opt := options.Index().SetName("my-index")
//		 	 keys := bson.D{{Key: "my-key", Value: 1}}
//		 	 model := mongo.IndexModel{Keys: keys, Options: opt}
//		 	 _, err := db.Collection("my-coll").Indexes().CreateOne(ctx, model)
//		 	 if err != nil {
//		 		 return err
//		 	 }
//		 	 return nil
//		 }, func(db *mongo.Database) error {
//		 	 _, err := db.Collection("my-coll").Indexes().DropOne(ctx, "my-index")
//		 	 if err != nil {
//		 		 return err
//		 	 }
//		 	 return nil
//		 })
//	 }
func Register(up, down MigrationFunc) error {
	return internalRegister(up, down, 2)
}

// MustRegister acts like Register but panics on errors.
func MustRegister(up, down MigrationFunc) {
	if err := internalRegister(up, down, 2); err != nil {
		panic(err)
	}
}

// RegisteredMigrations returns all registered migrations.
func RegisteredMigrations() []Migration {
	ret := make([]Migration, len(globalMigrate.migrations))
	copy(ret, globalMigrate.migrations)
	return ret
}

// SetDatabase sets database for global migrate.
func SetDatabase(db *mongo.Database) {
	globalMigrate.db = db
}

// SetMigrationsCollection changes default collection name for migrations history.
func SetMigrationsCollection(name string) {
	globalMigrate.SetMigrationsCollection(name)
}

// Version returns current database version.
func Version(ctx context.Context) (uint64, string, error) {
	return globalMigrate.Version(ctx)
}

// Up performs "up" migration using registered migrations.
// Detailed description available in Migrate.Up().
func Up(ctx context.Context, n int) error {
	return globalMigrate.Up(ctx, n)
}

// Down performs "down" migration using registered migrations.
// Detailed description available in Migrate.Down().
func Down(ctx context.Context, n int) error {
	return globalMigrate.Down(ctx, n)
}

// SetLogger sets a logger to print the migration process
func SetLogger(log Logger) {
	globalMigrate.SetLogger(log)
}
