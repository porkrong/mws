package submit

import (
	"encoding/csv"
	"io"
)

type csvTem struct {
	w        io.Writer
	csvWrite *csv.Writer
}

func NewCSV(w io.Writer) (obj *csvTem) {
	obj = new(csvTem)
	obj.w = w
	obj.csvWrite = csv.NewWriter(obj.w)
	return
}

//设置换行符
func (c *csvTem) SetDelimiter(char string) (err error) {
	c.csvWrite.Comma = '\t'
	return
}

//设置字符集
func (c *csvTem) setInputEncoding(charset string) {
	switch charset {
	case "UTF8":
		c.w.Write([]byte("\xEF\xBB\xBF"))
	case "Unicode":
		c.w.Write([]byte("\xFF\xFE"))
		return

	}
}

//写入一行
func (c *csvTem) InsertOne(rows []string) (err error) {
	err = c.csvWrite.Write(rows)
	return
}

func (c *csvTem) Flush() {
	c.csvWrite.Flush()
}

type buffer struct {
	data []byte
}

func NewBuffer() (obj *buffer) {
	obj = new(buffer)
	return
}

func (b *buffer) Write(data []byte) (n int, err error) {
	b.data = append(b.data, data...)
	return
}

func (b *buffer) Get() []byte {
	return b.data
}
