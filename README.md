# gocronexpr

A cron expression generation tool

## Usage

```go

cronExpr := cronexpr.NewCronExpr()
cronExpr.SetCondition(
    cronexpr.WithCustom(cronexpr.CronPosition.Sec, []int{0}),
    cronexpr.WithCustom(cronexpr.CronPosition.Min, []int{0}),
    cronexpr.WithCustom(cronexpr.CronPosition.Hour, []int{12}),
    cronexpr.WithCustom(cronexpr.CronPosition.Day, []int{15}),
    cronexpr.WithCustom(cronexpr.CronPosition.Mon, []int{1}),
    cronexpr.WithNoDesignate(cronexpr.CronPosition.Week),
    cronexpr.WithNoDesignate(cronexpr.CronPosition.Year),
)
specExpr, err := cronExpr.Gen()
// ....
```

## Options
```go
type cronExprOption struct {
	Every       bool  // 每秒(分)
	Range       []int // 指定范围，从n-m，每秒(分..)执行
	Interval    []int // 周期，从n秒开始，每m秒(分..)
	Custom      []int // 自定义日期
	NoDesignate bool  // 不指定，只有在日和周有用
	// 日专用
	NearDay int  // 每月最近的那个工作日
	LastDay bool // 最后一天
	// 周专用
	LastWeekDay int // 本月最后一个星期几

	position Position
	limit    []int // 限制值
}
```