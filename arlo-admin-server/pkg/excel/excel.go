package excel

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// Sheet 简单表格：首行表头 + 数据行
type Sheet struct {
	Name    string
	Headers []string
	Rows    [][]interface{}
}

// Write 生成 xlsx 字节
func Write(sheet Sheet) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	name := sheet.Name
	if name == "" {
		name = "Sheet1"
	}
	idx, err := f.NewSheet(name)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(idx)
	_ = f.DeleteSheet("Sheet1")

	for i, h := range sheet.Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(name, cell, h)
	}
	for r, row := range sheet.Rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(name, cell, v)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadRows 读取第一个工作表（跳过表头），返回每行字符串切片
func ReadRows(data []byte) (headers []string, rows [][]string, err error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("空工作簿")
	}
	all, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, nil
	}
	headers = all[0]
	for _, r := range all[1:] {
		rows = append(rows, r)
	}
	return headers, rows, nil
}

// WriteDownload 写出下载响应
func WriteDownload(c *gin.Context, filename string, data []byte) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// Cell 安全取单元格
func Cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}
