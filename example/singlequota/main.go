package main

import (
	"context"
	"flag"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ramadani/andromeda"
	"github.com/ramadani/andromeda/cache"
	"github.com/ramadani/andromeda/example/singlequota/internal"
	"github.com/ramadani/andromeda/example/singlequota/internal/repository"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Address          string           `yaml:"address"`
	DB               string           `yaml:"db"`
	SyncIn           time.Duration    `yaml:"syncIn"`
	Redis            RedisConfig      `yaml:"redis"`
	QuotaUsageConfig QuotaUsageConfig `yaml:"quotaUsage"`
}

type RedisConfig struct {
	Address string `yaml:"address"`
}

type QuotaUsageConfig struct {
	LockIn   time.Duration `yaml:"lockIn"`
	MaxRetry int           `yaml:"maxRetry"`
	RetryIn  time.Duration `yaml:"retryIn"`
}

func main() {
	ctx := context.Background()
	address := flag.String("address", "", "server address")

	flag.Parse()

	file, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	conf := new(Config)

	err = yaml.Unmarshal(fileContent, &conf)
	if err != nil {
		panic(err)
	}

	if addr := *address; strings.TrimSpace(addr) != "" {
		conf.Address = addr
	}

	db, err := sqlx.Connect("mysql", conf.DB)
	if err != nil {
		panic(err)
	}

	voucherRepo := repository.NewVoucherRepositoryMySqlx(db)
	historyRepo := repository.NewHistoryRepositoryMySqlx(db)

	redisClient := redis.NewClient(&redis.Options{Addr: conf.Redis.Address})
	cacheRedis := cache.NewCacheRedis(redisClient)

	keyUsageFormat := "voucher-quota-usage-%s"
	getVoucherQuotaUsageParams := internal.NewGetVoucherQuotaParams(keyUsageFormat)
	getVoucherQuotaLimit := internal.NewGetVoucherQuotaLimit(voucherRepo)
	getVoucherQuotaUsage := internal.NewGetVoucherQuotaUsage(voucherRepo)
	getCachedVoucherQuotaUsage := andromeda.NewGetCachedQuota(cacheRedis, getVoucherQuotaUsageParams)
	getVoucherQuotaUsageConf := andromeda.GetQuotaUsageConfig{
		LockIn:   conf.QuotaUsageConfig.LockIn,
		MaxRetry: conf.QuotaUsageConfig.MaxRetry,
		RetryIn:  conf.QuotaUsageConfig.RetryIn,
	}

	addVoucherUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
		Cache:               cacheRedis,
		GetQuotaCacheParams: getVoucherQuotaUsageParams,
		GetQuotaLimit:       getVoucherQuotaLimit,
		GetQuotaUsage:       getVoucherQuotaUsage,
		GetQuotaUsageConfig: getVoucherQuotaUsageConf,
	})

	reduceVoucherUsage := andromeda.ReduceQuotaUsage(andromeda.ReduceQuotaUsageConfig{
		Cache:               cacheRedis,
		GetQuotaCacheParams: getVoucherQuotaUsageParams,
		GetQuotaUsage:       getVoucherQuotaUsage,
		GetQuotaUsageConfig: getVoucherQuotaUsageConf,
	})

	claimVoucher := internal.NewClaimVoucher(voucherRepo, historyRepo, addVoucherUsage, reduceVoucherUsage)

	// sync voucher usage
	go func() {
		for {
			vouchers, err := voucherRepo.FindAll(ctx)
			if err != nil {
				panic(err)
			}

			for _, voucher := range vouchers {
				usage, err := getCachedVoucherQuotaUsage.Do(ctx, &andromeda.QuotaRequest{QuotaID: voucher.ID})
				if err == nil {
					if usage > 0 {
						voucher.Usage = usage
						err = voucherRepo.Update(ctx, voucher)
						if err != nil {
							log.Println("err update voucher", err)
						}
					}
				}
			}

			time.Sleep(conf.SyncIn)
		}
	}()

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	type dataReq struct {
		Code   string `json:"code"`
		UserID string `json:"userId"`
	}

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/claim", func(c echo.Context) error {
		ctx := c.Request().Context()
		input := new(dataReq)
		if err := c.Bind(&input); err != nil {
			return err
		}

		res, err := claimVoucher.Do(ctx, input.Code, input.UserID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	})

	// Start server
	e.Logger.Fatal(e.Start(conf.Address))
}
