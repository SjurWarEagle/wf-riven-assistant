package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"sort"
	"strconv"
)

var ColumnWeaponName = 0
var ColumnAllCount = 1
var ColumnAllMin = 2
var ColumnAllMax = 3
var ColumnAllAverage = 4
var ColumnLowerCount = 5
var ColumnLowerMin = 6
var ColumnLowerMax = 7
var ColumnLowerAverage = 8

func CreateAndShowTable(config Configuration) {
	var resultData []struct {
		name  string
		price RivenPrices
	}
	market := NewWfMarket()

	processing := 0
	for idx := range config.Rivens {
		processing++
		riven := config.Rivens[idx]
		weaponName := riven.Weapon
		log.Printf("Collecting data for '%s' (%d/%d)...", weaponName, idx, len(config.Rivens))
		price, err := market.requestAuctionPrice(weaponName, config)
		log.Printf("...done\n")
		if err != nil {
			panic(err)
		}

		singleData := struct {
			name  string
			price RivenPrices
		}{name: weaponName, price: price}
		resultData = append(resultData, singleData)
	}
	data := &TableData{
		config:       config,
		data:         resultData,
		sortAsc:      true,
		sortedColumn: 0,
	}
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, true).
		SetContent(data)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, column := table.GetSelection()
			if row == 0 {
				SortData(row, column, data)
			}
		}
		return event
	})

	table.Select(0, ColumnLowerAverage).
		SetFixed(1, 1).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				os.Exit(0)
			}
		})
	if err := tview.NewApplication().
		SetRoot(table, true).
		EnableMouse(true).
		Run(); err != nil {
		panic(err)
	}
}

func SortData(row int, column int, data *TableData) {
	if row != 0 {
		return
	}
	dataToSort := data.data
	if data.sortedColumn == column {
		data.sortAsc = !data.sortAsc
	} else {
		data.sortAsc = true
	}
	data.sortedColumn = column

	sort.Slice(dataToSort, func(i, j int) bool {
		if column == ColumnWeaponName {
			if data.sortAsc {
				return dataToSort[i].name < dataToSort[j].name
			}
			return dataToSort[i].name > dataToSort[j].name
		} else if column == ColumnAllAverage {
			if data.sortAsc {
				return dataToSort[i].price.All_Average < dataToSort[j].price.All_Average
			}
			return dataToSort[i].price.All_Average > dataToSort[j].price.All_Average
		} else if column == ColumnAllCount {
			if data.sortAsc {
				return dataToSort[i].price.All_Cnt < dataToSort[j].price.All_Cnt
			}
			return dataToSort[i].price.All_Cnt > dataToSort[j].price.All_Cnt
		} else if column == ColumnAllMin {
			if data.sortAsc {
				return dataToSort[i].price.All_Min < dataToSort[j].price.All_Min
			}
			return dataToSort[i].price.All_Min > dataToSort[j].price.All_Min
		} else if column == ColumnAllMax {
			if data.sortAsc {
				return dataToSort[i].price.All_Max < dataToSort[j].price.All_Max
			}
			return dataToSort[i].price.All_Max > dataToSort[j].price.All_Max
		} else if column == ColumnLowerCount {
			if data.sortAsc {
				return dataToSort[i].price.LowerSection_Cnt < dataToSort[j].price.LowerSection_Cnt
			}
			return dataToSort[i].price.LowerSection_Min > dataToSort[j].price.LowerSection_Min
		} else if column == ColumnLowerMin {
			if data.sortAsc {
				return dataToSort[i].price.LowerSection_Min < dataToSort[j].price.LowerSection_Min
			}
			return dataToSort[i].price.LowerSection_Max > dataToSort[j].price.LowerSection_Max
		} else if column == ColumnLowerMax {
			if data.sortAsc {
				return dataToSort[i].price.LowerSection_Max < dataToSort[j].price.LowerSection_Max
			}
			return dataToSort[i].price.LowerSection_Max > dataToSort[j].price.LowerSection_Max
		} else if column == ColumnLowerAverage {
			if data.sortAsc {
				return dataToSort[i].price.LowerSection_Average < dataToSort[j].price.LowerSection_Average
			}
			return dataToSort[i].price.LowerSection_Average > dataToSort[j].price.LowerSection_Average
		}
		return false
	})

	data.data = dataToSort
}

func (d *TableData) GetRowCount() int {
	//+1 due to header
	return len(d.data) + 1
}

func (d *TableData) GetColumnCount() int {
	return 9
}

func (d *TableData) GetCell(row, column int) *tview.TableCell {
	value := ""
	align := tview.AlignRight
	cellColor := tcell.ColorWhite

	switch column {
	case ColumnWeaponName:
		if row == 0 {
			value = "Weapon Name"
			align = tview.AlignCenter
		} else {
			value = d.data[row-1].name
			align = tview.AlignLeft
		}
		break
	case ColumnAllCount:
		if row == 0 {
			value = "All Cnt"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.All_Cnt)
		}
		break
	case ColumnAllMin:
		if row == 0 {
			value = "All Min"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.All_Min)
		}
		break
	case ColumnAllMax:
		if row == 0 {
			value = "All Max"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.All_Max)
		}
		break
	case ColumnAllAverage:
		if row == 0 {
			value = "All Avg"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.All_Average)
		}
		break
	case ColumnLowerCount:
		if row == 0 {
			value = "Lower Cnt"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.LowerSection_Cnt)
		}
		break
	case ColumnLowerMin:
		if row == 0 {
			value = "Lower Min"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.LowerSection_Min)
		}
		break
	case ColumnLowerMax:
		if row == 0 {
			value = "Lower Max"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.LowerSection_Max)
		}
		break
	case ColumnLowerAverage:
		if row == 0 {
			value = "Lower Avg"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(d.data[row-1].price.LowerSection_Average)
			if d.data[row-1].price.LowerSection_Average < d.config.Setup.LowerSectionAverageHighlightThreshold {
				cellColor = tcell.ColorRed
			}
		}
		break
	}
	if row == 0 {
		cellColor = tcell.ColorYellow
	}
	if column == 0 {
		cellColor = tcell.ColorYellow
	}
	return tview.NewTableCell(fmt.Sprintf("%s", value)).SetTextColor(cellColor).SetAlign(align)
}

type TableData struct {
	config Configuration
	tview.TableContentReadOnly
	sortedColumn int
	sortAsc      bool
	data         []struct {
		name  string
		price RivenPrices
	}
}
