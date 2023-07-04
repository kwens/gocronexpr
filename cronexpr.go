/**
 * @Author: kwens
 * @Date: 2022-09-02 09:42:47
 * @Description:
 */
package gocronexpr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	every = "*"
	no    = "?"
	last  = "L"
)

const (
	cronPositionSec = iota
	cronPositionMin
	cronPositionHour
	cronPositionDay
	cronPositionMon
	cronPositionWeek
	cronPositionYear
)

type Position int

var CronPosition = struct {
	Sec  Position
	Min  Position
	Hour Position
	Day  Position
	Mon  Position
	Week Position
	Year Position
}{
	cronPositionSec,
	cronPositionMin,
	cronPositionHour,
	cronPositionDay,
	cronPositionMon,
	cronPositionWeek,
	cronPositionYear,
}

type PositionLimit []int

var positionLimit = struct {
	sec  PositionLimit
	min  PositionLimit
	hour PositionLimit
	day  PositionLimit
	mon  PositionLimit
	week PositionLimit
}{
	PositionLimit{0, 59},
	PositionLimit{0, 59},
	PositionLimit{0, 23},
	PositionLimit{1, 31},
	PositionLimit{1, 12},
	PositionLimit{1, 7}, // 1是周日，以此类推
}

type CronExpr struct {
	sec  cronExprOption // 秒
	min  cronExprOption // 分
	hour cronExprOption // 时
	day  cronExprOption // 日
	mon  cronExprOption // 月
	week cronExprOption // 周
	year cronExprOption // 年
}

func NewCronExpr() *CronExpr {
	return &CronExpr{
		sec:  cronExprOption{Every: true, limit: positionLimit.sec, position: CronPosition.Sec},
		min:  cronExprOption{Every: true, limit: positionLimit.min, position: CronPosition.Min},
		hour: cronExprOption{Every: true, limit: positionLimit.hour, position: CronPosition.Hour},
		day:  cronExprOption{Every: true, limit: positionLimit.day, position: CronPosition.Day},
		mon:  cronExprOption{Every: true, limit: positionLimit.mon, position: CronPosition.Mon},
		week: cronExprOption{NoDesignate: true, limit: positionLimit.week, position: CronPosition.Week},
		year: cronExprOption{NoDesignate: true, position: CronPosition.Year},
	}
}

// SetCondition 设置表达式条件，同个位置的条件只会生效最后一个
func (ce *CronExpr) SetCondition(opts ...CronExprOption) error {
	for i, opt := range opts {
		if opt == nil {
			return fmt.Errorf("condition index[%d] error, please checkout your condition", i+1)
		}
		opt.apply(ce)
	}
	return nil
}

// Gen 生成表达式
// 如果某个位置的表达式条件不适用规则，则默认返回“*”，比如在min位置适用WithNoDesignate
// 如果传入的表达式条件为范围值，当不符合范围要求时，直接报错
func (ce *CronExpr) Gen() (string, error) {
	secSpec, err := ce.gen(ce.sec)
	if err != nil {
		return "", err
	}
	minSpec, err := ce.gen(ce.min)
	if err != nil {
		return "", err
	}
	hourSpec, err := ce.gen(ce.hour)
	if err != nil {
		return "", err
	}
	daySpec, err := ce.gen(ce.day)
	if err != nil {
		return "", err
	}
	monSpec, err := ce.gen(ce.mon)
	if err != nil {
		return "", err
	}
	weekSpec, err := ce.gen(ce.week)
	if err != nil {
		return "", err
	}
	yearSpec, err := ce.gen(ce.year)
	if err != nil {
		return "", err
	}
	if err := ce.checkDayAndWeek(daySpec, weekSpec); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s %s %s %s %s", secSpec, minSpec, hourSpec, daySpec, monSpec, weekSpec, yearSpec), nil
}
func (ce *CronExpr) checkDayAndWeek(daySpec, weekSpec string) error {
	if (daySpec == no && weekSpec != no) || (daySpec != no && weekSpec == no) {
		return nil
	}
	return errors.New("day and week can't set conditon at the same time")
}

func (ce *CronExpr) gen(opt cronExprOption) (string, error) {
	if opt.rangeCheck() {
		return "", fmt.Errorf("position:%d, condition range check error", opt.position)
	}
	if opt.Every {
		return opt.genEvery(), nil
	}
	if len(opt.Range) > 0 {
		return opt.genRange(), nil
	}
	if len(opt.Interval) > 0 {
		return opt.genInterval(), nil
	}
	if len(opt.Custom) > 0 {
		return opt.genCustom(), nil
	}
	if opt.NoDesignate {
		return opt.genNoDesignate(), nil
	}
	if opt.NearDay > 0 {
		return opt.genNearDay(), nil
	}
	if opt.LastDay {
		return opt.genLastDay(), nil
	}
	if opt.LastWeekDay > 0 {
		return opt.genLastWeekDay(), nil
	}
	return every, nil
}

type CronExprOption interface {
	apply(ce *CronExpr)
}

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

// outRange 范围检查
func (cep cronExprOption) rangeCheck() bool {
	if len(cep.Range) == 1 {
		return false
	}
	if len(cep.Range) > 1 {
		for _, v := range cep.Range {
			if v > cep.limit[1] || v < cep.limit[0] {
				return true
			}
		}
	}
	if len(cep.Interval) == 1 {
		return false
	}
	if len(cep.Interval) > 1 {
		if cep.position == CronPosition.Week {
			if cep.Interval[0] > 4 || cep.Interval[0] < 1 {
				return true
			}
			if cep.Interval[1] > cep.limit[1] || cep.Interval[1] < cep.limit[0] {
				return true
			}
		} else {
			for _, v := range cep.Interval {
				if v > cep.limit[1] || v < cep.limit[0] {
					return true
				}
			}
		}
	}
	if len(cep.Custom) > 0 {
		for _, v := range cep.Custom {
			if v > cep.limit[1] || v < cep.limit[0] {
				return true
			}
		}
	}
	if cep.NearDay > 0 {
		if cep.NearDay > cep.limit[1] || cep.NearDay < cep.limit[0] {
			return true
		}
	}
	if cep.LastWeekDay > 0 {
		if cep.LastWeekDay > cep.limit[1] || cep.LastWeekDay < cep.limit[0] {
			return true
		}
	}
	return false
}

func (cep cronExprOption) apply(ce *CronExpr) {
	switch cep.position {
	case CronPosition.Sec:
		cep.limit = ce.sec.limit
		ce.sec = cep
	case CronPosition.Min:
		cep.limit = ce.min.limit
		ce.min = cep
	case CronPosition.Hour:
		cep.limit = ce.hour.limit
		ce.hour = cep
	case CronPosition.Day:
		cep.limit = ce.day.limit
		ce.day = cep
	case CronPosition.Mon:
		cep.limit = ce.mon.limit
		ce.mon = cep
	case CronPosition.Week:
		cep.limit = ce.week.limit
		ce.week = cep
	case CronPosition.Year:
		cep.limit = ce.year.limit
		ce.year = cep
	}
}

// genEvery 生成每秒、分、时、天、月、周、年
func (cep cronExprOption) genEvery() string {
	return every
}

func (cep cronExprOption) genRange() string {
	if cep.position == CronPosition.Week {
		return fmt.Sprintf("%d/%d", cep.Range[0], cep.Range[1])
	}
	return fmt.Sprintf("%d-%d", cep.Range[0], cep.Range[1])
}

// genInterval 生成周期从第Range[0]秒(分、时、天、月日(当位置在月份的时候，表示从哪一日))开始，每Range[1]秒(分、时、天、月)执行一次
// 针对周的，表示第Range[0]周的星期Range[1], 其中Range[0] 范围1-4
func (cep cronExprOption) genInterval() string {
	if cep.position == CronPosition.Week {
		return fmt.Sprintf("%d#%d", cep.Interval[0], cep.Interval[1])
	}
	return fmt.Sprintf("%d/%d", cep.Interval[0], cep.Interval[1])
}

func (cep cronExprOption) genCustom() string {
	var spec string
	for i, v := range cep.Custom {
		if i == 0 {
			spec = fmt.Sprintf("%d", v)
		} else {
			spec += fmt.Sprintf(",%d", v)
		}
	}
	return spec
}

func (cep cronExprOption) genNoDesignate() string {
	if cep.position == CronPosition.Year {
		return ""
	}
	return no
}

func (cep cronExprOption) genNearDay() string {
	return fmt.Sprintf("%dW", cep.NearDay)
}

func (cep cronExprOption) genLastDay() string {
	return last
}

func (cep cronExprOption) genLastWeekDay() string {
	return fmt.Sprintf("%d%s", cep.LastWeekDay, last)
}

// ================== 选项方法 ================== //
func emptyOption(position Position) *cronExprOption {
	return &cronExprOption{
		position: position,
	}
}
func WithEvery(position Position) CronExprOption {
	opt := emptyOption(position)
	opt.Every = true
	return opt
}

func WithRange(begin, end int, position Position) CronExprOption {
	opt := emptyOption(position)
	opt.Range = []int{begin, end}
	return opt
}

func WithInterval(begin, sep int, position Position) CronExprOption {
	opt := emptyOption(position)
	opt.Interval = []int{begin, sep}
	return opt
}

// WithCustom
func WithCustom(custom []int, position Position) CronExprOption {
	opt := emptyOption(position)
	opt.Custom = custom
	return opt
}

// WithNoDesignate 不指定，只有日/星期有, 传入其他位置则返回nil
func WithNoDesignate(position Position) CronExprOption {
	switch position {
	case CronPosition.Day, CronPosition.Week, CronPosition.Year:
		opt := emptyOption(position)
		opt.NoDesignate = true
		return opt
	default:
		return nil
	}
}

// WithNearWorkDay 每月几号最近的工作日
// @param.day: 1-31
func WithNearWorkDay(day int) CronExprOption {
	opt := emptyOption(CronPosition.Day)
	opt.NearDay = day
	return opt
}

// WithLastDay 本月最后一天
func WithLastDay() CronExprOption {
	opt := emptyOption(CronPosition.Day)
	opt.LastDay = true
	return opt
}

// WithLastWeek 本月最后一个星期几
// @param.weekDay: 1-7
func WithLastWeek(weekDay int) CronExprOption {
	opt := emptyOption(CronPosition.Week)
	opt.LastWeekDay = weekDay
	return opt
}

// With24TimeStr 24小时制的时间设置，eg: 12:20:00
func With24TimeStr(timestr string) []CronExprOption {
	timeSlice := strings.Split(timestr, ":")
	h, _ := strconv.Atoi(timeSlice[0])
	m, _ := strconv.Atoi(timeSlice[1])
	s, _ := strconv.Atoi(timeSlice[2])
	hourOpt := emptyOption(CronPosition.Hour)
	hourOpt.Custom = append(hourOpt.Custom, h)
	minOpt := emptyOption(CronPosition.Min)
	minOpt.Custom = append(hourOpt.Custom, m)
	secOpt := emptyOption(CronPosition.Sec)
	secOpt.Custom = append(hourOpt.Custom, s)
	return []CronExprOption{hourOpt, minOpt, secOpt}
}
