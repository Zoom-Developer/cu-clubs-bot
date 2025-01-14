package config

import (
	"context"
	"fmt"
	postgresStorage "github.com/Badsnus/cu-clubs-bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Config struct {
	Database *gorm.DB
	Redis    *redis.Client
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := os.Setenv("BOT_TOKEN", viper.GetString("bot.token")); err != nil {
		panic(err)
	}
}

func Get() *Config {
	initConfig()

	err := logger.Init(logger.Config{
		Debug:     viper.GetBool("settings.debug"),
		TimeZone:  viper.GetString("settings.timezone"),
		LogToFile: viper.GetBool("settings.log-to-file"),
		LogsDir:   viper.GetString("settings.logs-dir"),
	})
	if err != nil {
		panic(err)
	}

	var gormConfig *gorm.Config
	if viper.GetBool("settings.debug") {
		newLogger := gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				SlowThreshold: time.Second,
				LogLevel:      gormLogger.Info,
				Colorful:      true,
			},
		)
		gormConfig = &gorm.Config{
			Logger: newLogger,
		}
	} else {
		gormConfig = &gorm.Config{}
	}

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=GMT+3",
		viper.GetString("service.database.user"),
		viper.GetString("service.database.password"),
		viper.GetString("service.database.name"),
		viper.GetString("service.database.host"),
		viper.GetInt("service.database.port"),
	)

	database, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		logger.Log.Panicf("Failed to connect to the database: %v", err)
	} else {
		logger.Log.Info("Successfully connected to the database")
	}

	errMigrate := database.AutoMigrate(postgresStorage.Migrations...)
	if errMigrate != nil {
		logger.Log.Panicf("Failed to migrate database: %v", errMigrate)
	}

	redisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("service.redis.host"), viper.GetInt("service.redis.port")),
		Password: viper.GetString("service.redis.password"),
		DB:       0,
	})
	err = redisDB.Ping(context.Background()).Err()
	if err != nil {
		logger.Log.Panicf("Failed to connect to redis: %v", err)
	} else {
		logger.Log.Info("Successfully connected to redis")
	}

	return &Config{
		Database: database,
		Redis:    redisDB,
	}
}
