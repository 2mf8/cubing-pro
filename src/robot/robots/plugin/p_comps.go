package plugin

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/guojia99/go-tables/table"
	"sort"
	"strconv"
	"strings"
)

type CompsPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &CompsPlugin{}

func (c *CompsPlugin) ID() []string {
	return []string{"comp", "comps", "比赛", "比赛列表", "比赛赛果", "赛果", "比赛成绩", "比赛打乱", "打乱"}
}

func (c *CompsPlugin) Help() string {
	return `获取比赛信息:
1. 比赛/比赛列表: 可查询当前比赛列表
2. 比赛-{名称/序号}: 可查询比赛的详细信息
3. 比赛成绩/赛果/比赛赛果-{名称/序号} {项目} {轮次} {排名}: 
              可查询比赛成绩详细列表.
              例如 "赛果-1 333 初赛 30", 代表查询比赛1 三阶初赛前30名
`
}

func (c *CompsPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := message.Message
	if strings.Contains(msg, "comps") || strings.Contains(msg, "比赛列表") {
		return c.comps(message)
	}
	if strings.Contains(msg, "赛果") || strings.Contains(msg, "成绩") {
		return c.compResult(message)
	}
	if strings.Contains(msg, "打乱") {
		return c.compScramble(message)
	}

	return c.comp(message)
}

func (c *CompsPlugin) compScramble(message types.InMessage) (*types.OutMessage, error) {
	//comp, err := c._getComps(message)
	//if err != nil {
	//	return message.NewOutMessage(err.Error()), nil
	//}
	//
	////comp.CompJSON.Events
	return nil, nil
}

func (c *CompsPlugin) _getComps(message types.InMessage) (competition.Competition, error) {
	msg := RemoveID(message.Message, c.ID())
	msg = utils.ReplaceAll(msg, "", "-")

	var id = 0
	if number := utils.GetNumbers(msg); len(number) > 0 {
		id = int(number[0])
	}
	var comp competition.Competition
	var err error
	if id == 0 {
		err = c.Svc.DB.Where("status = ?", competition.Running).Order("created_at DESC").First(&comp).Error
	} else {
		err = c.Svc.DB.Where("status = ?", competition.Running).Where("id = ?", id).First(&comp).Error
	}
	if err != nil {
		return comp, fmt.Errorf("找不到比赛%d", id)
	}
	return comp, nil
}

type compResultTable struct {
	Rank   string `table:"排名"`
	Name   string `table:"选手"`
	Single string `table:"单次"`
	Avg    string `table:"平均"`
}

func (c *CompsPlugin) compResult(message types.InMessage) (*types.OutMessage, error) {

	comp, err := c._getComps(message)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}

	msg := RemoveID(message.Message, c.ID())
	msg = utils.ReplaceAll(msg, "", "-")
	group := strings.Split(msg, " ")

	if len(group) < 2 {
		return message.NewOutMessage("请输入一个项目"), nil
	}

	// 比赛成绩/赛果/比赛赛果-{名称/序号} {项目} {轮次} {排名}

	// 项目处理
	ev, _, _, err := GetMessageEvent(GetEvents(c.Svc, ""), group[1])
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}
	var compEv competition.CompetitionEvent
	for _, v := range comp.CompJSON.Events {
		if v.EventID == ev.ID {
			compEv = v
			break
		}
	}
	if len(compEv.EventID) == 0 {
		return message.NewOutMessage("本场比赛未开设该项目"), nil
	}

	// 其他参数处理
	roundStr := "决赛"
	if len(group) >= 3 {
		roundStr = group[2]
	}
	num := 10
	if len(group) >= 4 {
		num, _ = strconv.Atoi(group[3])
		if num < 3 {
			num = 3
		}
		if num > 50 {
			num = 50
		}
	}

	results := c.Svc.Cov.SelectCompsResult(comp.ID)

	rrr, ok := results[ev.ID]
	if !ok {
		return message.NewOutMessage("该项目未有成绩"), nil
	}
	schedule, err := compEv.CurRunningSchedule(roundStr, nil)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}

	rr, ok := rrr[schedule.RoundNum]
	if !ok {
		return message.NewOutMessage(fmt.Sprintf("本场比赛`%s`项目没有 `%s` 轮次", group[1], roundStr)), nil
	}

	out := fmt.Sprintf("比赛: %s\n", comp.Name)
	out += fmt.Sprintf("项目: %s\n", ev.Cn)
	out += fmt.Sprintf("轮次: %s\n", schedule.Round)

	var tabs []compResultTable
	for i := 0; i < num && i < len(rr); i++ {
		r := rr[i]
		tb := compResultTable{
			Rank:   fmt.Sprintf("%d", r.Rank),
			Name:   r.PersonName,
			Single: r.BestString(),
			Avg:    r.BestAvgString(),
		}
		tabs = append(tabs, tb)
	}

	tb, _ := table.SimpleTable(tabs, &table.Option{
		ExpendID: false,
		Align:    table.AlignLeft,
		Contour:  table.EmptyContour,
	})
	out += tb.String()
	return message.NewOutMessage(out), nil
}

func (c *CompsPlugin) comp(message types.InMessage) (*types.OutMessage, error) {
	comp, err := c._getComps(message)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}
	var out = fmt.Sprintf("%s\n\n", comp.Name)

	// 基本信息
	out += fmt.Sprintf("状态: %s\n", comp.StatusName())
	out += fmt.Sprintf("比赛时间: %s ~ %s\n", comp.CompStartTime.Format("20060102"), comp.CompEndTime.Format("20060102"))
	out += fmt.Sprintf("比赛ID: %d\n", comp.ID)
	out += "比赛项目: "
	var events = GetEvents(c.Svc, comp.EventMin)
	for _, ev := range events {
		out += ev.Cn + " "
	}

	evs := GetEvents(c.Svc, "")

	// 赛果
	if comp.IsDone {
		var cr []result.Record
		var gr []result.Record

		c.Svc.DB.Where("comps_id = ?", comp.ID).Where("d_type = ?", result.RecordTypeWithCubingPro).Find(&cr)
		c.Svc.DB.Where("comps_id = ?", comp.ID).Where("d_type = ?", result.RecordTypeWithGroup).Find(&gr)

		if len(cr) > 0 || len(gr) > 0 {
			out += "\n本场比赛打破的记录:\n"
		}
		if len(gr) > 0 {
			out += "======== GR =========\n"
			out += _recordTable(gr, evs, len(gr)).String()
		}
		if len(cr) > 0 {
			out += "======== CR =========\n"
			out += _recordTable(cr, evs, len(cr)).String()
		}
	}

	return message.NewOutMessage(out), nil
}

func (c *CompsPlugin) comps(message types.InMessage) (*types.OutMessage, error) {
	var comps []competition.Competition

	var page = 0
	numbers := utils.GetNumbers(message.Message)
	if len(numbers) > 0 {
		page = int(numbers[0])
	}
	offset := (page - 1) * 10

	if err := c.Svc.DB.Where("status = ?", competition.Running).Order("created_at DESC").Offset(offset).Limit(10).Find(&comps).Error; err != nil {
		return nil, err
	}

	var group competition.CompetitionGroup
	c.Svc.DB.Where("qq_groups LIKE ?", fmt.Sprintf("%%%d%%", message.GroupID)).First(&group)

	out := "===== 比赛列表 =====\n"
	sort.Slice(comps, func(i, j int) bool {
		if comps[i].IsDone {
			if !comps[j].IsDone {
				return true
			}
		}
		return comps[i].ID > comps[j].ID
	})

	done := false
	out += "--- 已结束的比赛:\n"
	for _, comp := range comps {
		if !comp.IsDone && !done {
			out += "--- 进行中的比赛:\n"
			done = true
		}

		thisGroup := ""
		if comp.GroupID == group.ID {
			thisGroup = "[本群]"
		}

		out += fmt.Sprintf("%d.[%s] %s %s\n", comp.ID, comp.CompStartTime.Format("01月02日"), thisGroup, comp.Name)

	}
	return message.NewOutMessage(out), nil
}