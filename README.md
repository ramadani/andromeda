# andromeda

andromeda is a golang package that can be implemented on an object that has quotas.

## Feature Overview

- Add and reduce quota usage
- Handle race conditions
- Can be applied to multiple quotas
- Reversible if an error occurs
- Handling if quota usage is not exists in redis
- Can be implemented according to your needs

### Use cases

- Flash sale
- Claim vouchers
- etc

## Getting Started

### Prerequisites

andromeda requires [redis](https://redis.io/) to store quota data, make sure your machine has it.

### Installation

```
go get github.com/ramadani/andromeda
```

## Guide

A quota consists of limits and usage. so the entity that you have must have these 2 attributes.

There are 4 functions that you need to prepare for each of your quota.

1. Get quota limit  
Get the limit of a quota. Limit is used to keep usage from exceeding the limit.

2. Get quota usage  
Get updated usage of a quota. This function will be called if the quota usage is not in Redis. After the value of the quota usage is obtained, the value will be saved to redis according to its key and expiration.

3. Get quota usage key  
Generate key that will be used to store and get quota usage stored in redis.

4. Get quota usage expiration  
Expiration is used to set how long the duration of a quota usage is stored in Redis.

### Example

For example, we will use claim voucher use case.

To get quota limit and usage, we create code using `GetQuota` interface

**Get quota limit**

```go
type getVoucherQuotaLimit struct {
	voucherRepo VoucherRepository
}

func (v *getVoucherQuotaLimit) Do(ctx context.Context, req *andromeda.QuotaRequest) (int64, error) {
	voucher, err := v.voucherRepo.FindByID(ctx, req.QuotaID)
	if err != nil {
		return 0, err
	}

	return voucher.Limit, nil
}

func NewGetVoucherQuotaLimit(voucherRepo VoucherRepository) andromeda.GetQuota {
	return &getVoucherQuotaLimit{voucherRepo: voucherRepo}
}
```

**Get quota usage**

```go
type getVoucherQuotaUsage struct {
	voucherRepo VoucherRepository
}

func (v *getVoucherQuotaUsage) Do(ctx context.Context, req *andromeda.QuotaRequest) (int64, error) {
	voucher, err := v.voucherRepo.FindByID(ctx, req.QuotaID)
	if err != nil {
		return 0, err
	}

	return voucher.Usage, nil
}

func NewGetVoucherQuotaUsage(voucherRepo VoucherRepository) andromeda.GetQuota {
	return &getVoucherQuotaUsage{voucherRepo: voucherRepo}
}
```

To get quota usage key, we create code using `GetQuotaKey` interface

**Get quota usage key**

```go
type getVoucherQuotaUsageKey struct {
	keyFormat string
}

func (v *getVoucherQuotaUsageKey) Do(_ context.Context, req *andromeda.QuotaRequest) (string, error) {
	return fmt.Sprintf(v.keyFormat, req.QuotaID), nil
}

func NewGetVoucherQuotaUsageKey(keyFormat string) andromeda.GetQuotaKey {
	return &getVoucherQuotaUsageKey{keyFormat: keyFormat}
}
```

To get quota usage expiration, we create code using the `GetQuotaExpiration` interface

**Get quota usage expiration**

```go
type getVoucherQuotaUsageExpiration struct{}

func (v *getVoucherQuotaUsageExpiration) Do(ctx context.Context, req *andromeda.QuotaRequest) (time.Duration, error) {
	return time.Hour * 5, nil
}

func NewGetVoucherQuotaUsageExpiration() andromeda.GetQuotaExpiration {
	return &getVoucherQuotaUsageExpiration{}
}
```

Then use these 4 code implementations to create add quota usage and reduce quota usage functions

```go
keyUsageFormat := "voucher-quota-usage-%s"
getVoucherQuotaLimit := internal.NewGetVoucherQuotaLimit(voucherRepo)
getVoucherQuotaUsage := internal.NewGetVoucherQuotaUsage(voucherRepo)
getVoucherQuotaUsageKey := internal.NewGetVoucherQuotaUsageKey(keyUsageFormat)
getVoucherQuotaUsageExpiration := internal.NewGetVoucherQuotaUsageExpiration()
getVoucherQuotaUsageConf := andromeda.GetQuotaUsageConfig{
	LockIn:   time.Second * 3,
	MaxRetry: 10,
	RetryIn:  time.Millisecond * 100,
}

addVoucherUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
	Cache:                   cacheRedis,
	GetQuotaLimit:           getVoucherQuotaLimit,
	GetQuotaUsage:           getVoucherQuotaUsage,
	GetQuotaUsageKey:        getVoucherQuotaUsageKey,
	GetQuotaUsageExpiration: getVoucherQuotaUsageExpiration,
	GetQuotaUsageConfig:     getVoucherQuotaUsageConf,
})

reduceVoucherUsage := andromeda.ReduceQuotaUsage(andromeda.ReduceQuotaUsageConfig{
	Cache:                   cacheRedis,
	GetQuotaUsage:           getVoucherQuotaUsage,
	GetQuotaUsageKey:        getVoucherQuotaUsageKey,
	GetQuotaUsageExpiration: getVoucherQuotaUsageExpiration,
	GetQuotaUsageConfig:     getVoucherQuotaUsageConf,
})
```

use these functions in your logic code

```go
func (c *claimVoucher) Do(ctx context.Context, code, userID string) (*History, error) {
	voucher, err := c.voucherRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	history := &History{
		ID:        uuid.NewV4().String(),
		VoucherID: voucher.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	usageReq := andromeda.QuotaUsageRequest{
		QuotaID: voucher.ID,
		Usage:   1,
	}

	if _, err = c.addVoucherUsage.Do(ctx, &usageReq); err != nil {
		return nil, err
	}

	if err = c.historyRepo.Create(ctx, history); err == nil {
		return history, nil
	}

	// has error, do reverse usage quota
	if _, er := c.reduceVoucherUsage.Do(ctx, &usageReq); er != nil {
		err = er
	}

	return nil, err
}
```

#### Listener

You can implement a listener to know the events on success or error using the `UpdateQuotaUsageListener` interface.

```go
type updateVoucherQuotaUsageListener struct{}

func (v *updateVoucherQuotaUsageListener) OnSuccess(ctx context.Context, req *andromeda.QuotaUsageRequest, updatedUsage int64) {
	log.Println("updated quota", updatedUsage)
}

func (v *updateVoucherQuotaUsageListener) OnError(ctx context.Context, req *andromeda.QuotaUsageRequest, err error) {
	log.Println("err", err)
}

func NewUpdateVoucherQuotaUsageListener() andromeda.UpdateQuotaUsageListener {
	return &updateVoucherQuotaUsageListener{}
}
```

Add the listener to the option

```go
updateVoucherQuotaUsageListener := internal.NewUpdateVoucherQuotaUsageListener()

addVoucherUsage := andromeda.AddQuotaUsage(andromeda.AddQuotaUsageConfig{
	Cache:                   cacheRedis,
	GetQuotaLimit:           getVoucherQuotaLimit,
	GetQuotaUsage:           getVoucherQuotaUsage,
	GetQuotaUsageKey:        getVoucherQuotaUsageKey,
	GetQuotaUsageExpiration: getVoucherQuotaUsageExpiration,
	GetQuotaUsageConfig:     getVoucherQuotaUsageConf,
	Option: andromeda.AddUsageOption{
		Listener: updateVoucherQuotaUsageListener,
	},
})

reduceVoucherUsage := andromeda.ReduceQuotaUsage(andromeda.ReduceQuotaUsageConfig{
	Cache:                   cacheRedis,
	GetQuotaUsage:           getVoucherQuotaUsage,
	GetQuotaUsageKey:        getVoucherQuotaUsageKey,
	GetQuotaUsageExpiration: getVoucherQuotaUsageExpiration,
	GetQuotaUsageConfig:     getVoucherQuotaUsageConf,
	Option: andromeda.ReduceUsageOption{
		Listener: updateVoucherQuotaUsageListener,
	},
})
```

Check out the [examples](example) to find out more

### Tips

1. Use job scheduling to update usage from redis to database
2. Set expiration is longer than the original quota time. For example, the quota period is only 3 days, the set expiration is more than 3 days so that the value in redis will still be there when the job scheduling period is still running
3. To get the quota limit, you can save it to redis so that it doesn't always get it from the database

## Contributing

1. Fork the Project
2. Create your Feature Branch (git checkout -b feature/AmazingFeature)
3. Commit your Changes (git commit -m 'Add some AmazingFeature')
4. Push to the Branch (git push origin feature/AmazingFeature)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.

## References

- [Race condition](https://en.wikipedia.org/wiki/Race_condition)
- [Redis incr/decr](https://redis.io/)